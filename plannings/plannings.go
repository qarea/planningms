// Package plannings checks and saves
package plannings

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

var (
	timeNowFunc = func() int64 {
		return time.Now().Unix()
	}

	errOfflineSpentTime = errors.New("offline spent time")
)

const fiveMinutesInSeconds = 60 * 5

// NewService creates new plannings.Service instance
func NewService(config PlanningServiceCfg) *Service {
	return &Service{
		spentTimeStorage: config.SpentTimeStorage,
		planningStorage:  config.PlanningStorage,

		maxPlanningAge:    config.MaxPlanningAge,
		maxFromLastUpdate: config.MaxPeriodFromLastUpdate,
	}
}

// PlanningServiceCfg config for Service
type PlanningServiceCfg struct {
	SpentTimeStorage SpentTimeStorage
	PlanningStorage  PlanningStorage

	MaxPlanningAge          time.Duration
	MaxPeriodFromLastUpdate time.Duration
}

// SpentTimeStorage required api
type SpentTimeStorage interface {
	NewSpentTime(context.Context, entities.SpentTime) error
	Modify(context.Context, ctxtg.UserID, entities.ModifySpentTimeFunc) error
}

// PlanningStorage required api
type PlanningStorage interface {
	AddSpentTime(context.Context, entities.SpentTimeHistory) error
	ClosePlanning(context.Context, ctxtg.UserID, entities.PlanningReport) error
	PlanningCreatedAt(context.Context, entities.PlanningID) (int64, error)
	LastActivity(context.Context, ctxtg.UserID) (int64, error)
	Planning(context.Context, entities.PlanningID) (*entities.Planning, error)
	OpenedPlannings(context.Context, ctxtg.UserID) ([]entities.ExtendedPlanning, error)
	SpentTimeByUserIDTimeRange(ctx context.Context, uid ctxtg.UserID, from, to int64) (int, error)
}

// Service contains implements all needed business logic
type Service struct {
	spentTimeStorage SpentTimeStorage
	planningStorage  PlanningStorage

	maxPlanningAge    time.Duration
	maxFromLastUpdate time.Duration
}

// OpenedPlannings returns all opened plannings for uid and marks outdated
func (s *Service) OpenedPlannings(ctx context.Context, uid ctxtg.UserID) ([]entities.ExtendedPlanning, error) {
	ps, err := s.planningStorage.OpenedPlannings(ctx, uid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load openedPlannings")
	}
	now := timeNowFunc()
	for i, p := range ps {
		ps[i].Outdated = s.isOutdated(now, p.CreatedAt, p.LastActivity)
	}
	return ps, nil
}

// AddSpentTime registers spent time.
// Returns err if:
// - no active planning
// - invalid report
// - failed to save offline time to database
func (s *Service) AddSpentTime(ctx context.Context, report entities.SpentTimeReport) error {
	var spentTime entities.SpentTime
	spentTimeFunc := checkNotEmpty(func(st entities.SpentTime) (*entities.SpentTime, error) {
		spentTime = st
		return &st, nil
	})
	err := s.spentTimeStorage.Modify(ctx, report.UserID, spentTimeFunc)
	if err != nil {
		return err
	}
	newReport, err := s.checkReport(spentTime, report)
	if err == errOfflineSpentTime {
		saveErr := s.planningStorage.AddSpentTime(ctx, reportToHistory(newReport, entities.Offline))
		if saveErr != nil {
			return errors.Wrapf(err, "fail to save spenttime history: %+v", saveErr)
		}
	} else if err != nil {
		return errors.Wrap(err, "invalid report")
	}
	incrementTimeFunc := checkNotEmpty(func(st entities.SpentTime) (*entities.SpentTime, error) {
		st.Last = newReport.Time
		st.SpentOnline += newReport.Spent
		return &st, nil
	})
	return s.spentTimeStorage.Modify(ctx, report.UserID, incrementTimeFunc)
}

// ClosePlanning close planning for userID with report
// Return error if planning id is invalid it returns error
func (s *Service) ClosePlanning(ctx context.Context, userID ctxtg.UserID, report entities.PlanningReport) error {
	modifyFunc := ifNotEmpty(ifPlanningID(report.PlanningID, s.toHistory(ctx)))
	err := s.spentTimeStorage.Modify(ctx, userID, modifyFunc)
	if err != nil {
		return errors.Wrap(err, "failed to modify spentTime storage")
	}
	err = s.planningStorage.ClosePlanning(ctx, userID, report)
	if err != nil {
		return errors.Wrap(err, "failed ot close planning")
	}
	return nil
}

// SpentTime returns total SpentTime amount for time period in seconds
func (s *Service) SpentTime(ctx context.Context, uid ctxtg.UserID, from, to int64) (int, error) {
	var onlineSpent int
	err := s.spentTimeStorage.Modify(ctx, uid, ifNotEmpty(func(st entities.SpentTime) (*entities.SpentTime, error) {
		if st.Started >= from && st.Started <= to {
			onlineSpent = st.SpentOnline
		}
		return &st, nil
	}))
	if err != nil {
		return 0, errors.Wrap(err, "cache error")
	}
	spent, err := s.planningStorage.SpentTimeByUserIDTimeRange(ctx, uid, from, to)
	if err != nil {
		return 0, errors.Wrap(err, "planning storage error")
	}
	return onlineSpent + spent, nil
}

// SetActive save previous active planning's spent time from spenttime storage to planning storage
// and register new spent time storage instance
func (s *Service) SetActive(ctx context.Context, a entities.NewActivePlanning) error {
	err := s.spentTimeStorage.Modify(ctx, a.UserID, ifNotEmpty(s.toHistory(ctx)))
	if err != nil {
		return errors.Wrap(err, "failed spentTime storage modification")
	}
	if a.PlanningID == 0 {
		return nil
	}
	lastActivity, err := s.planningStorage.LastActivity(ctx, a.UserID)
	if err != nil {
		return errors.Wrap(err, "failed to load planning lastActivity")
	}
	if lastActivity > a.Time {
		return entities.ErrOutdatedReport
	}
	createdAt, err := s.planningStorage.PlanningCreatedAt(ctx, a.PlanningID)
	if err != nil {
		return errors.Wrap(err, "failed to load planning StartedAt")
	}
	err = s.spentTimeStorage.NewSpentTime(ctx, entities.SpentTime{
		UserID:            a.UserID,
		PlanningID:        a.PlanningID,
		PlanningCreatedAt: createdAt,
		Started:           a.Time,
		Last:              a.Time,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create new spent time")
	}
	return nil
}

func (s *Service) toHistory(ctx context.Context) spentTimeFunc {
	return func(st entities.SpentTime) (*entities.SpentTime, error) {
		err := s.spentTimeToHistory(ctx, st, entities.Online)
		if err != nil {
			return &st, err
		}
		return nil, nil
	}
}

func (s *Service) spentTimeToHistory(ctx context.Context, spentTime entities.SpentTime, status entities.SpentTimeStatus) error {
	history := spentTimeToHistory(spentTime, status)
	err := s.planningStorage.AddSpentTime(ctx, history)
	if err != nil {
		return errors.Wrap(err, "failed to save spent time history to storage")
	}
	return nil
}

type spentTimeFunc func(st entities.SpentTime) (*entities.SpentTime, error)

func checkNotEmpty(f spentTimeFunc) entities.ModifySpentTimeFunc {
	return func(st *entities.SpentTime) (*entities.SpentTime, error) {
		if st == nil {
			return nil, entities.ErrNoActivePlanning
		}
		return f(*st)
	}
}

func ifNotEmpty(f spentTimeFunc) entities.ModifySpentTimeFunc {
	return func(st *entities.SpentTime) (*entities.SpentTime, error) {
		if st == nil {
			return nil, nil
		}
		return f(*st)
	}
}

func ifPlanningID(planningID entities.PlanningID, f spentTimeFunc) spentTimeFunc {
	return func(st entities.SpentTime) (*entities.SpentTime, error) {
		if st.PlanningID != planningID {
			return &st, nil
		}
		return f(st)
	}
}

func (s *Service) checkReport(st entities.SpentTime, report entities.SpentTimeReport) (entities.SpentTimeReport, error) {
	now := timeNowFunc()
	if now < st.Last || now < st.Started {
		return report, errors.Errorf("Invalid spentTime now: %d , instance: %+v", now, st)
	}
	if st.PlanningID != report.PlanningID {
		return report, entities.ErrInvalidPlanningID
	}
	if report.Time < st.Last {
		return report, entities.ErrOutdatedReport
	}
	if report.Spent < 0 {
		return report, entities.ErrNegativeSpentTime
	}
	if s.isOutdated(report.Time, st.PlanningCreatedAt, st.Last) {
		return report, entities.ErrPlanningOutdated
	}
	if report.Time-st.Last < int64(report.Spent) {
		report.Spent = int(report.Time - st.Last)
	}
	if now-st.Last > fiveMinutesInSeconds {
		return report, errOfflineSpentTime
	}
	return report, nil
}

func (s *Service) isOutdated(time, created, last int64) bool {
	return time-created > int64(s.maxPlanningAge.Seconds()) ||
		(last > 0 && time-last > int64(s.maxFromLastUpdate.Seconds()))
}

func reportToHistory(r entities.SpentTimeReport, status entities.SpentTimeStatus) entities.SpentTimeHistory {
	return entities.SpentTimeHistory{
		PlanningID: r.PlanningID,
		Spent:      r.Spent,
		StartedAt:  r.Time - int64(r.Spent),
		EndedAt:    r.Time,
		Status:     status,
	}
}

func spentTimeToHistory(t entities.SpentTime, s entities.SpentTimeStatus) entities.SpentTimeHistory {
	return entities.SpentTimeHistory{
		PlanningID: t.PlanningID,
		Spent:      t.SpentOnline,
		StartedAt:  t.Started,
		EndedAt:    t.Last,
		Status:     s,
	}
}

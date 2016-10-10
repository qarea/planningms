// Package storage provide API to access database
package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/powerman/narada-go/narada"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

const mysqlForeignKeyErrorCode = 1452

var timeNowFunc = func() int64 {
	return time.Now().Unix()
}

// NewPlanningStorage setups and created PlanningStorage
func NewPlanningStorage(db *sqlx.DB, sharedLock time.Duration) *PlanningStorage {
	return &PlanningStorage{
		sharedLockDuration: sharedLock,
		db:                 db,
	}
}

// PlanningStorage provide needed operations on plannings
type PlanningStorage struct {
	sharedLockDuration time.Duration
	db                 *sqlx.DB
}

// Planning return planning by pid
func (p *PlanningStorage) Planning(_ context.Context, pid entities.PlanningID) (*entities.Planning, error) {
	var planning *entities.Planning
	err := p.withSharedLock(func() error {
		var err error
		planning, err = findPlanning(p.db, pid)
		return err
	})
	return planning, err
}

// OpenedPlannings returned all opened plannings for uid
func (p *PlanningStorage) OpenedPlannings(_ context.Context, uid ctxtg.UserID) ([]entities.ExtendedPlanning, error) {
	ps, err := openedPlannings(p.db, uid)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load opened plannings")
	}
	estimations, err := estimationsForPlannings(p.db, ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load estimations")
	}
	lastActivities, err := lastActivityForPlannings(p.db, ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load last activities")
	}
	var plannings []entities.ExtendedPlanning
	for _, p := range ps {
		plannings = append(plannings, entities.ExtendedPlanning{
			Planning:     p,
			Estimation:   estimations[p.ID],
			LastActivity: lastActivities[p.ID],
		})
	}
	return plannings, nil
}

// PlanningCreatedAt returns createdAt field for pid
func (p *PlanningStorage) PlanningCreatedAt(_ context.Context, pid entities.PlanningID) (int64, error) {
	var createdAt int64
	err := p.withSharedLock(func() error {
		var err error
		createdAt, err = planningCreatedAt(p.db, pid)
		return err
	})
	return createdAt, err
}

// AddExtraTime create new estimation for planning
func (p *PlanningStorage) AddExtraTime(_ context.Context, uid ctxtg.UserID, np entities.PlannedTime) error {
	return p.withSharedLockAndTransaction(func(tx sqlx.Ext) error {
		p, err := findPlanning(tx, np.PlanningID)
		if err != nil {
			return errors.Wrap(err, "failed to load planning")
		}
		if p == nil {
			return entities.ErrInvalidPlanningID
		}
		if p.UserID != uid {
			return entities.ErrInvalidUserID
		}
		np.CreatedAt = timeNowFunc()
		_, err = savePlannedTime(tx, np)
		return err
	})
}

// AddSpentTime save new SpentTimeHistory
func (p *PlanningStorage) AddSpentTime(_ context.Context, h entities.SpentTimeHistory) error {
	return p.withSharedLockAndTransaction(func(tx sqlx.Ext) error {
		_, err := saveHistory(tx, h)
		if err != nil {
			return errors.Wrap(err, "failed to save history")
		}
		err = addSpentTimeToPlanning(tx, h)
		if err != nil {
			return errors.Wrap(err, "failed to add spent time to planning")
		}
		return nil
	})
}

// CreatePlanning create new planning and new planned time
func (p *PlanningStorage) CreatePlanning(_ context.Context, np entities.NewPlanning) (entities.PlanningID, error) {
	var id entities.PlanningID
	err := p.withSharedLockAndTransaction(func(tx sqlx.Ext) error {
		var err error
		id, err = savePlanning(tx, newPlanningToPlanning(np))
		if err != nil {
			return errors.Wrap(err, "failed to save planning")
		}
		_, err = savePlannedTime(tx, entities.PlannedTime{
			PlanningID: id,
			Estimation: np.Estimation,
			CreatedAt:  timeNowFunc(),
		})
		if err != nil {
			return errors.Wrap(err, "failed to save planned time")
		}
		return nil
	})
	return id, err
}

// LastActivity for user id
func (p *PlanningStorage) LastActivity(_ context.Context, uid ctxtg.UserID) (int64, error) {
	var last int64
	err := p.withSharedLock(func() error {
		var err error
		last, err = lastActivityForUser(p.db, uid)
		return err
	})
	return last, err
}

// SpentTimeByUserIDTimeRange return total spent time for user for time range
func (p *PlanningStorage) SpentTimeByUserIDTimeRange(ctx context.Context, uid ctxtg.UserID, from, to int64) (int, error) {
	return spentTime(p.db, uid, from, to)
}

// ClosePlanning check user id, save history and update planning
func (p *PlanningStorage) ClosePlanning(_ context.Context, uid ctxtg.UserID, report entities.PlanningReport) error {
	return p.withSharedLockAndTransaction(func(tx sqlx.Ext) error {
		planning, err := findPlanning(tx, report.PlanningID)
		if err != nil {
			return errors.Wrap(err, "failed to find planning")
		}
		if planning == nil {
			return entities.ErrInvalidPlanningID
		}
		if planning.UserID != uid {
			return entities.ErrInvalidUserID
		}
		if planning.Status == entities.Closed {
			return entities.ErrPlanningClosed
		}
		histories, err := findHistories(tx, report.PlanningID)
		if err != nil {
			return errors.Wrap(err, "failed to load histories")
		}
		planning.SpentOnline, planning.SpentOffline = countTime(histories)
		planning.Status = entities.Closed
		planning.Reported = report.Time
		planning.IssueDone = report.Progress

		err = updatePlanning(tx, *planning)
		if err != nil {
			return errors.Wrap(err, "failed to update planning")
		}
		return nil
	})
}

func (p *PlanningStorage) withSharedLockAndTransaction(f func(tx sqlx.Ext) error) error {
	return p.withSharedLock(func() error {
		tx, err := p.db.Beginx()
		if err != nil {
			return errors.Wrap(err, "failed to start transaction")
		}
		defer tx.Rollback()
		if err := f(tx); err != nil {
			return err
		}
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "failed to commit transaction")
		}
		return nil
	})
}

func (p *PlanningStorage) withSharedLock(f func() error) error {
	l, err := narada.SharedLock(p.sharedLockDuration)
	if err != nil {
		return errors.Wrapf(entities.ErrMaintenance, "can't obtain lock %v", err)
	}
	defer l.UnLock()
	return f()
}

func newPlanningToPlanning(np entities.NewPlanning) entities.Planning {
	return entities.Planning{
		UserID:          np.UserID,
		Status:          entities.Open,
		ProjectID:       np.ProjectID,
		TrackerID:       np.TrackerID,
		IssueID:         np.IssueID,
		IssueTitle:      np.IssueTitle,
		IssueDueDate:    np.IssueDueDate,
		IssueEstimation: np.IssueEstimation,
		IssueDone:       np.IssueDone,
		IssueURL:        np.IssueURL,
		ActivityID:      np.ActivityID,
		CreatedAt:       timeNowFunc(),
		SpentOnline:     0,
		SpentOffline:    0,
		Reported:        0,
	}
}

func countTime(hs []entities.SpentTimeHistory) (spentOnline, spentOffline int) {
	for _, h := range hs {
		switch h.Status {
		case entities.Online:
			spentOnline += h.Spent
		case entities.Offline:
			spentOffline += h.Spent
		}
	}
	return spentOnline, spentOffline
}

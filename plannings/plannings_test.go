package plannings

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

var ctx = context.Background()

func TestMain(m *testing.M) {
	rand.Seed(time.Now().Unix())
	os.Exit(m.Run())
}

func TestOpenedPlanningsEmpty(t *testing.T) {
	ps := newPlanningStorage()
	svc := &Service{
		planningStorage: ps,
	}
	pls, err := svc.OpenedPlannings(ctx, 123)
	if err != nil {
		t.Fatal(err)
	}
	if len(pls) != 0 {
		t.Error("Should be empty")
	}
}

func TestOpenedPlannings(t *testing.T) {
	maxPlanningAge := 20 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25)()
	uid := randomUserID()

	ps := newPlanningStorage()
	ps.addPlanning(entities.Planning{
		ID: randomPlanningID(),
	})
	ps.addPlanning(entities.Planning{
		ID: randomPlanningID(),
	})
	svc := &Service{
		planningStorage: ps,

		maxPlanningAge:    maxPlanningAge,
		maxFromLastUpdate: maxFromLastUpdate,
	}
	pls, err := svc.OpenedPlannings(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(pls) != 2 {
		t.Error("Should have 2")
	}
	if ps.userID != uid {
		t.Error("Invalid user id passed")
	}
	for _, p := range pls {
		if !p.Outdated {
			t.Error("Should be outdated")
		}
	}
}

func TestAddSpentTimeNoActivePlanning(t *testing.T) {
	spentTimeStorage := newSpentTimeStorage()
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
	}
	testAmount := entities.SpentTimeReport{
		UserID:     1,
		PlanningID: 2,
	}
	err := svc.AddSpentTime(ctx, testAmount)
	if err != entities.ErrNoActivePlanning {
		t.Error("Invalid err", err)
	}
}

func TestAddSpentTimeInvalidPlanningID(t *testing.T) {
	defer mockTimeNow(10)()
	userID := ctxtg.UserID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		Started:    0,
		PlanningID: 5,
		Last:       5,
	}
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
	}
	testAmount := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: 2,
		Spent:      3,
	}
	err := svc.AddSpentTime(ctx, testAmount)
	if errors.Cause(err) != entities.ErrInvalidPlanningID {
		t.Error("Invalid err", err)
	}
}

func TestAddSpentTimePlanningTooOldOnline(t *testing.T) {
	maxPlanningAge := 20 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              20,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       25,
	}
	err := svc.AddSpentTime(ctx, report)
	if errors.Cause(err) != entities.ErrPlanningOutdated {
		t.Error("Unexpected error", err)
	}
	if *spentTimeStorage.spentTime[userID] != initialSpentTime {
		t.Errorf("Spent time shouldn't be changed %+v", spentTimeStorage.spentTime[userID])
	}
}

func TestAddSpentTimePlanningTooOldOffline(t *testing.T) {
	maxPlanningAge := 20 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25 + fiveMinutesInSeconds)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              20,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       25,
	}
	err := svc.AddSpentTime(ctx, report)
	if errors.Cause(err) != entities.ErrPlanningOutdated {
		t.Error("Unexpected error", err)
	}
	if *spentTimeStorage.spentTime[userID] != initialSpentTime {
		t.Error("Spent time shouldn't be changed")
	}
}

func TestAddSpentTimePlanningTooLongAfterLastUpdateOnline(t *testing.T) {
	maxPlanningAge := 400 * time.Second
	maxFromLastUpdate := 5 * time.Second
	defer mockTimeNow(25)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              15,
	}
	spentTime := initialSpentTime
	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       25,
	}
	err := svc.AddSpentTime(ctx, report)
	if errors.Cause(err) != entities.ErrPlanningOutdated {
		t.Error("Unexpected error", err)
	}
	if *spentTimeStorage.spentTime[userID] != initialSpentTime {
		t.Error("Spent time shouldn't be changed")
	}
}

func TestAddSpentTimePlanningTooLongAfterLastUpdateOffline(t *testing.T) {
	maxPlanningAge := 400 * time.Second
	maxFromLastUpdate := 5 * time.Second
	defer mockTimeNow(25 + fiveMinutesInSeconds)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              15,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       25,
	}
	err := svc.AddSpentTime(ctx, report)
	if errors.Cause(err) != entities.ErrPlanningOutdated {
		t.Error("Unexpected error", err)
	}
	if *spentTimeStorage.spentTime[userID] != initialSpentTime {
		t.Error("Spent time shouldn't be changed")
	}
}

func TestAddSpentTimePlanningOnlineMissedSecond(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	nowTime := int64(25)
	defer mockTimeNow(nowTime)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		SpentOnline:       10,
		Last:              15,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      10,
		Time:       25,
	}
	err := svc.AddSpentTime(ctx, report)
	if err != nil {
		t.Fatal(err)
	}
	if spentTimeStorage.spentTime[userID].SpentOnline != 20 {
		t.Error("Unexpected spent value", spentTimeStorage.spentTime[userID].SpentOnline)
	}
}

func TestAddSpentTimePlanningOnline(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	var initialSpentOnline = 10
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		SpentOnline:       initialSpentOnline,
		Last:              15,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       20,
	}
	err := svc.AddSpentTime(ctx, report)
	if err != nil {
		t.Fatal(err)
	}
	if spentTimeStorage.spentTime[userID].SpentOnline != initialSpentOnline+report.Spent ||
		spentTimeStorage.spentTime[userID].Last != report.Time {
		t.Error("Invalid spent time", initialSpentTime.SpentOnline, report.Spent, spentTimeStorage.spentTime[userID].SpentOnline, spentTimeStorage.spentTime[userID].Last)
	}
}

func TestAddSpentTimePlanningOffline(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25 + fiveMinutesInSeconds)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	planningStorage := newPlanningStorage()
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              15,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		PlanningStorage:         planningStorage,
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       20,
	}
	err := svc.AddSpentTime(ctx, report)
	if err != nil {
		t.Fatal(err)
	}
	if spentTimeStorage.spentTime[userID].SpentOnline != initialSpentTime.SpentOnline+report.Spent ||
		spentTimeStorage.spentTime[userID].Last != report.Time {
		t.Error("Invalid spent time", spentTimeStorage.spentTime[userID].SpentOnline)
	}
	expectedSpentTimeHistory := entities.SpentTimeHistory{
		PlanningID: planningID,
		Spent:      report.Spent,
		StartedAt:  report.Time - int64(report.Spent),
		EndedAt:    report.Time,
		Status:     entities.Offline,
	}
	if planningStorage.histories[0] != expectedSpentTimeHistory {
		t.Errorf("Invalid history in planning storage %+v", planningStorage.histories[0])
	}
}

func TestAddSpentTimePlanningTimeLessThenLast(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25 + fiveMinutesInSeconds)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	planningStorage := newPlanningStorage()
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              15,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		PlanningStorage:         planningStorage,
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      5,
		Time:       10,
	}
	err := svc.AddSpentTime(ctx, report)
	if errors.Cause(err) != entities.ErrOutdatedReport {
		t.Error("Unexpected err", err)
	}
	if *spentTimeStorage.spentTime[userID] != initialSpentTime {
		t.Error("Spent time shouldn't be changed")
	}
}

func TestAddSpentTimePlanningSpentNegative(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25 + fiveMinutesInSeconds)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	planningStorage := newPlanningStorage()
	spentTimeStorage := newSpentTimeStorage()
	initialSpentTime := entities.SpentTime{
		UserID:            userID,
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		Last:              15,
	}
	spentTime := initialSpentTime

	spentTimeStorage.spentTime[userID] = &spentTime
	svc := NewService(PlanningServiceCfg{
		PlanningStorage:         planningStorage,
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      -5,
		Time:       20,
	}
	err := svc.AddSpentTime(ctx, report)
	if errors.Cause(err) != entities.ErrNegativeSpentTime {
		t.Error("Unexpected err", err)
	}
	if *spentTimeStorage.spentTime[userID] != initialSpentTime {
		t.Error("Spent time shouldn't be changed")
	}
}

func TestAddSpentTimePlanningOnlineTooMuchTime(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	var initialSpentOnline = 15
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		Started:           0,
		SpentOnline:       initialSpentOnline,
		Last:              15,
	}
	svc := NewService(PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      10,
		Time:       20,
	}
	err := svc.AddSpentTime(ctx, report)
	if err != nil {
		t.Fatal(err)
	}
	if spentTimeStorage.spentTime[userID].SpentOnline != 20 ||
		spentTimeStorage.spentTime[userID].Last != report.Time {
		t.Error("Invalid spent time", spentTimeStorage.spentTime[userID].SpentOnline, spentTimeStorage.spentTime[userID].Last)
	}
}

func TestAddSpentTimePlanningOfflineTooMuchTime(t *testing.T) {
	maxPlanningAge := 35 * time.Second
	maxFromLastUpdate := 10 * time.Second
	defer mockTimeNow(25 + fiveMinutesInSeconds)()
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	planningStorage := newPlanningStorage()
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		PlanningID:        planningID,
		PlanningCreatedAt: 0,
		SpentOnline:       15,
		Started:           0,
		Last:              15,
	}
	svc := NewService(PlanningServiceCfg{
		PlanningStorage:         planningStorage,
		SpentTimeStorage:        spentTimeStorage,
		MaxPlanningAge:          maxPlanningAge,
		MaxPeriodFromLastUpdate: maxFromLastUpdate,
	})
	report := entities.SpentTimeReport{
		UserID:     userID,
		PlanningID: planningID,
		Spent:      10,
		Time:       20,
	}
	err := svc.AddSpentTime(ctx, report)
	if err != nil {
		t.Fatal(err)
	}
	if spentTimeStorage.spentTime[userID].SpentOnline != 20 ||
		spentTimeStorage.spentTime[userID].Last != report.Time {
		t.Error("Invalid spent time")
	}
	expectedSpentTimeHistory := entities.SpentTimeHistory{
		PlanningID: planningID,
		Spent:      5,
		StartedAt:  15,
		EndedAt:    report.Time,
		Status:     entities.Offline,
	}
	if planningStorage.histories[0] != expectedSpentTimeHistory {
		t.Errorf("Invalid history in planning storage %+v", planningStorage.histories[0])
	}
}

func TestClosePlanningStorageErr(t *testing.T) {
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	spentTime := entities.SpentTime{
		UserID:     userID,
		PlanningID: planningID,
		Started:    rand.Int63(),
		Last:       rand.Int63(),
	}
	planningStorage := newPlanningStorage()
	planningStorage.err = errors.New("Planning storage")
	spentTimeStorage := newSpentTimeStorage()
	err := spentTimeStorage.NewSpentTime(ctx, spentTime)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
		planningStorage:  planningStorage,
	}
	err = svc.ClosePlanning(ctx, userID, entities.PlanningReport{
		PlanningID: planningID,
	})
	if errors.Cause(err) != planningStorage.err {
		t.Error("Invalid error", err)
	}
}

func TestClosePlanningTracker(t *testing.T) {
	userID := ctxtg.UserID(rand.Int63())
	planningID := entities.PlanningID(rand.Int63())
	activityID := rand.Int63()
	progress := rand.Int()
	reportTime := rand.Int63()
	spentTime := entities.SpentTime{
		UserID:     userID,
		PlanningID: planningID,
		Started:    rand.Int63(),
		Last:       rand.Int63(),
	}
	planningStorage := newPlanningStorage()
	p := entities.Planning{
		ID:         planningID,
		UserID:     userID,
		ActivityID: entities.ActivityID(activityID),
		TrackerID:  entities.TrackerID(rand.Int63()),
		ProjectID:  entities.ProjectID(rand.Int63()),
		IssueID:    entities.IssueID(rand.Int63()),
		Status:     entities.Open,
	}
	planningStorage.addPlanning(p)
	spentTimeStorage := newSpentTimeStorage()
	err := spentTimeStorage.NewSpentTime(ctx, spentTime)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
		planningStorage:  planningStorage,
	}
	planningReport := entities.PlanningReport{
		PlanningID: planningID,
		Progress:   progress,
		Time:       reportTime,
	}
	err = svc.ClosePlanning(ctx, userID, planningReport)
	if err != nil {
		t.Error("Invalid error", err)
	}
	if planningStorage.plannings[planningID].Status != entities.Closed {
		t.Error("Invalid status")
	}
}

func TestSetActiveTimeStorageErr(t *testing.T) {
	planningID := entities.PlanningID(rand.Int63())
	userID := ctxtg.UserID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.err = errors.New("spent time err")

	svc := &Service{
		spentTimeStorage: spentTimeStorage,
	}
	err := svc.SetActive(ctx, entities.NewActivePlanning{
		UserID:     userID,
		PlanningID: planningID,
		Time:       rand.Int63(),
	})
	if errors.Cause(err) != spentTimeStorage.err {
		t.Error("Invalid err", err)
	}
}

func TestSetActivePlanningStorageErr(t *testing.T) {
	planningID := entities.PlanningID(rand.Int63())
	userID := ctxtg.UserID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		UserID:     userID,
		PlanningID: planningID,
	}
	planningStorage := newPlanningStorage()
	planningStorage.err = errors.New("planning err")

	svc := &Service{
		planningStorage:  planningStorage,
		spentTimeStorage: spentTimeStorage,
	}
	err := svc.SetActive(ctx, entities.NewActivePlanning{
		UserID:     userID,
		PlanningID: planningID,
		Time:       rand.Int63(),
	})

	if errors.Cause(err) != planningStorage.err {
		t.Error("Invalid err", err)
	}
}

func TestSetActivePlanningNoActive(t *testing.T) {
	ts := rand.Int63()
	createdAt := rand.Int63()
	planningID := entities.PlanningID(rand.Int63())
	userID := ctxtg.UserID(rand.Int63())
	spentTimeStorage := newSpentTimeStorage()
	planningStorage := newPlanningStorage()
	planningStorage.plannings[planningID] = &entities.Planning{
		CreatedAt: createdAt,
	}

	svc := &Service{
		planningStorage:  planningStorage,
		spentTimeStorage: spentTimeStorage,
	}
	err := svc.SetActive(ctx, entities.NewActivePlanning{
		UserID:     userID,
		PlanningID: planningID,
		Time:       ts,
	})
	if err != nil {
		t.Error("Invalid err", err)
	}
	st := spentTimeStorage.spentTime[userID]
	if st.Last != ts ||
		st.UserID != userID ||
		st.PlanningID != planningID ||
		st.Started != ts ||
		st.PlanningCreatedAt != createdAt {
		t.Errorf("Invalid spent time %+v", st)
	}
}

func TestSetActivePlanning(t *testing.T) {
	var now int64 = 20
	defer mockTimeNow(now)()
	planningID := entities.PlanningID(rand.Int63())
	userID := ctxtg.UserID(rand.Int63())
	createdAt := rand.Int63()
	startedAt := rand.Int63()
	spent := rand.Int()
	last := rand.Int63()
	ts := now + last
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		PlanningID:  planningID,
		Started:     startedAt,
		Last:        last,
		SpentOnline: spent,
		UserID:      userID,
	}
	planningStorage := newPlanningStorage()
	planningStorage.plannings[planningID] = &entities.Planning{
		CreatedAt: createdAt,
	}

	svc := &Service{
		planningStorage:  planningStorage,
		spentTimeStorage: spentTimeStorage,
	}
	err := svc.SetActive(ctx, entities.NewActivePlanning{
		UserID:     userID,
		PlanningID: planningID,
		Time:       ts,
	})

	if err != nil {
		t.Error("Invalid err", err)
	}
	p := planningStorage.histories[0]
	if p.Status != entities.Online ||
		p.PlanningID != planningID ||
		p.Spent != spent ||
		p.StartedAt != startedAt ||
		p.EndedAt != last {
		t.Errorf("Invalid history %+v", p)
	}

	st := spentTimeStorage.spentTime[userID]

	if st.UserID != userID ||
		st.PlanningID != planningID ||
		st.PlanningCreatedAt != createdAt ||
		st.Started != ts ||
		st.Last != ts ||
		st.SpentOnline != 0 {
		t.Errorf("Invalid spent time %+v", st)
	}
}

func TestSetActivePlanningOutdatedReport(t *testing.T) {
	var now int64 = 20
	defer mockTimeNow(now)()
	planningID := entities.PlanningID(rand.Int63())
	userID := ctxtg.UserID(rand.Int63())
	startedAt := rand.Int63()
	spent := rand.Int()
	last := rand.Int63()
	var ts int64 = 9
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		PlanningID:  planningID,
		Started:     startedAt,
		Last:        last,
		SpentOnline: spent,
		UserID:      userID,
	}
	planningStorage := newPlanningStorage()
	planningStorage.lastActivity = 10

	svc := &Service{
		planningStorage:  planningStorage,
		spentTimeStorage: spentTimeStorage,
	}
	err := svc.SetActive(ctx, entities.NewActivePlanning{
		UserID:     userID,
		PlanningID: planningID,
		Time:       ts,
	})

	if err != entities.ErrOutdatedReport {
		t.Error("Unexpected error", err)
	}
	if spentTimeStorage.spentTime[userID] != nil {
		t.Error("Should be empty")
	}
}

func TestSetActivePlanningZeroID(t *testing.T) {
	var now int64 = 20
	defer mockTimeNow(now)()
	ts := rand.Int63()
	planningID := entities.PlanningID(rand.Int63())
	userID := ctxtg.UserID(rand.Int63())
	startedAt := rand.Int63()
	spent := rand.Int()
	last := rand.Int63()
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		PlanningID:  planningID,
		Started:     startedAt,
		Last:        last,
		SpentOnline: spent,
		UserID:      userID,
	}
	planningStorage := newPlanningStorage()

	svc := &Service{
		planningStorage:  planningStorage,
		spentTimeStorage: spentTimeStorage,
	}
	err := svc.SetActive(ctx, entities.NewActivePlanning{
		UserID:     userID,
		PlanningID: 0,
		Time:       ts,
	})

	if err != nil {
		t.Error("Invalid err", err)
	}
	p := planningStorage.histories[0]
	if p.Status != entities.Online ||
		p.PlanningID != planningID ||
		p.Spent != spent ||
		p.StartedAt != startedAt ||
		p.EndedAt != last {
		t.Errorf("Invalid history %+v", p)
	}
	st := spentTimeStorage.spentTime[userID]
	if st != nil {
		t.Error("Should be empty")
	}
}

func TestSpentTimeSpentTimeStorageErr(t *testing.T) {
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.err = errors.New("sts err")
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
	}
	_, err := svc.SpentTime(ctx, 1, 2, 3)
	if errors.Cause(err) != spentTimeStorage.err {
		t.Error("Unexpected error", err)
	}
}

func TestSpentTimePlanningStorageErr(t *testing.T) {
	spentTimeStorage := newSpentTimeStorage()
	planningStorage := newPlanningStorage()
	planningStorage.err = errors.New("err")
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
		planningStorage:  planningStorage,
	}
	_, err := svc.SpentTime(ctx, 1, 2, 3)
	if errors.Cause(err) != planningStorage.err {
		t.Error("Unexpected error", err)
	}
}

func TestSpentTime(t *testing.T) {
	var userID ctxtg.UserID = 1
	spentTimeStorage := newSpentTimeStorage()
	spentTimeStorage.spentTime[userID] = &entities.SpentTime{
		Started:     2,
		SpentOnline: 10,
	}

	planningStorage := newPlanningStorage()
	planningStorage.spent = 20

	var from int64 = 1
	var to int64 = 4
	svc := &Service{
		spentTimeStorage: spentTimeStorage,
		planningStorage:  planningStorage,
	}
	spent, err := svc.SpentTime(ctx, userID, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if spent != 10+20 {
		t.Error("Invalid spent", spent)
	}
	if planningStorage.from != from {
		t.Error("Invalid from", planningStorage.from)
	}
	if planningStorage.to != to {
		t.Error("Invalid to", planningStorage.to)
	}

}

func TestConvertion(t *testing.T) {
	spent, status := randomSpentTimeAndStatus()
	h := spentTimeToHistory(spent, status)
	if h.PlanningID != spent.PlanningID ||
		h.StartedAt != spent.Started ||
		h.EndedAt != spent.Last ||
		h.Status != status {
		t.Error("Invalid fields")
	}
}

func TestOutdated(t *testing.T) {
	s := &Service{}
	s.maxPlanningAge = 5 * time.Second
	s.maxFromLastUpdate = 3 * time.Second
	type test struct {
		time, created, last int64
		expected            bool
	}
	var tests = []test{
		{10, 7, 0, false},
		{10, 7, 7, false},
		{10, 3, 3, true},
	}
	for i, test := range tests {
		result := s.isOutdated(test.time, test.created, test.last)
		if result != test.expected {
			t.Errorf("Expected %v, value %v in test %d", test.expected, result, i)
		}
	}
}

//func TestOutdated(t *testing.T) {
//	s := &Service{}
//	s.maxPlanningAge = 5 * time.Second
//	s.maxFromLastUpdate = 3 * time.Second
//	if s.isOutdated(10, 7, 0) {
//		t.Error("Should not be outdated")
//	}
//}
//
//func TestOutdated2(t *testing.T) {
//	s := &Service{}
//	s.maxPlanningAge = 5 * time.Second
//	s.maxFromLastUpdate = 3 * time.Second
//	if s.isOutdated(10, 7, 7) {
//		t.Error("Should not be outdated")
//	}
//}
//
//func TestOutdated3(t *testing.T) {
//	s := &Service{}
//	s.maxPlanningAge = 5 * time.Second
//	s.maxFromLastUpdate = 3 * time.Second
//	if s.isOutdated(10, 3, 3) {
//		t.Error("Should be outdated")
//	}
//
//}

func randomSpentTimeAndStatus() (entities.SpentTime, entities.SpentTimeStatus) {
	return entities.SpentTime{
		UserID:     ctxtg.UserID(rand.Int63()),
		PlanningID: entities.PlanningID(rand.Int63()),
		Started:    rand.Int63(),
		Last:       rand.Int63(),
	}, randomStatus()
}

func randomStatus() entities.SpentTimeStatus {
	if 1 != rand.Int63()%2 {
		return entities.Online
	}
	return entities.Offline
}

func randomUserID() ctxtg.UserID {
	return ctxtg.UserID(rand.Int63())
}

func randomPlanningID() entities.PlanningID {
	return entities.PlanningID(rand.Int63())
}

func mockTimeNow(timeToReturn int64) func() {
	timeNowFunc = func() int64 { return timeToReturn }
	return func() {
		timeNowFunc = func() int64 {
			return time.Now().Unix()
		}
	}
}

func newPlanningStorage() *testPlanningStorage {
	return &testPlanningStorage{
		plannings: make(map[entities.PlanningID]*entities.Planning),
	}
}

type testPlanningStorage struct {
	histories []entities.SpentTimeHistory
	plannings map[entities.PlanningID]*entities.Planning
	userID    ctxtg.UserID
	from      int64
	to        int64

	spent        int
	lastActivity int64
	err          error
}

type planningsKey struct {
	UserID     ctxtg.UserID
	PlanningID entities.PlanningID
}

func (t *testPlanningStorage) addPlanning(p entities.Planning) {
	t.plannings[p.ID] = &p
}

func (t *testPlanningStorage) Planning(_ context.Context, pid entities.PlanningID) (*entities.Planning, error) {
	return t.plannings[pid], t.err
}

func (t *testPlanningStorage) AddSpentTime(_ context.Context, history entities.SpentTimeHistory) error {
	t.histories = append(t.histories, history)
	return t.err
}

func (t *testPlanningStorage) ClosePlanning(_ context.Context, userID ctxtg.UserID, r entities.PlanningReport) error {
	p := t.plannings[r.PlanningID]
	if p != nil && p.UserID == userID {
		p.Status = entities.Closed
		p.Reported = r.Time
		t.plannings[p.ID] = p
	}
	return t.err
}

func (t *testPlanningStorage) OpenedPlannings(_ context.Context, userID ctxtg.UserID) ([]entities.ExtendedPlanning, error) {
	t.userID = userID
	var ps []entities.ExtendedPlanning
	for _, p := range t.plannings {
		ps = append(ps, entities.ExtendedPlanning{
			Planning: *p,
		})
	}
	return ps, t.err
}

func (t *testPlanningStorage) LastActivity(_ context.Context, userID ctxtg.UserID) (int64, error) {
	return t.lastActivity, t.err
}

func (t *testPlanningStorage) PlanningCreatedAt(_ context.Context, pid entities.PlanningID) (int64, error) {
	p := t.plannings[pid]
	if p != nil {
		return p.CreatedAt, t.err
	}
	return 0, t.err
}

func (t *testPlanningStorage) SpentTimeByUserIDTimeRange(_ context.Context, uid ctxtg.UserID, from, to int64) (int, error) {
	t.userID = uid
	t.from = from
	t.to = to
	return t.spent, t.err

}

func newSpentTimeStorage() *testSpentTimeStorage {
	return &testSpentTimeStorage{
		spentTime: make(map[ctxtg.UserID]*entities.SpentTime),
	}
}

type testSpentTimeStorage struct {
	spentTime map[ctxtg.UserID]*entities.SpentTime

	err error
}

func (t *testSpentTimeStorage) NewSpentTime(_ context.Context, spentTime entities.SpentTime) error {
	t.spentTime[spentTime.UserID] = &spentTime
	return t.err
}

func (t *testSpentTimeStorage) Modify(_ context.Context, userID ctxtg.UserID, f entities.ModifySpentTimeFunc) error {
	origS := t.spentTime[userID]
	if origS != nil {
		origS.UserID = userID
	}
	s, err := f(origS)
	t.spentTime[userID] = s
	if err != nil {
		return err
	}
	return t.err
}

///  +build integration

package storage

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
	"github.com/qarea/planningms/mysqldb"
)

var (
	ctx    = context.Background()
	second = time.Second
)

func TestOpenedPlanningsEmptyTable(t *testing.T) {
	defer prepareDB()()
	st := NewPlanningStorage(mysqldb.New(), second)
	ps, err := st.OpenedPlannings(ctx, 2)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	if len(ps) != 0 {
		t.Error("Invalid amount", len(ps))
	}
}

func TestOpenedPlannings(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	uid := ctxtg.UserID(rand.Int63())
	st := NewPlanningStorage(db, second)
	p1 := saveTestPlanningOpened(db, t, uid)
	p2 := saveTestPlanningOpened(db, t, uid)
	p3 := saveTestPlanningOpened(db, t, uid)
	p4 := saveTestPlanningClosed(db, t, uid)
	pts1 := saveExtraTimeForPlanning(db, t, p1.ID)
	pts2 := saveExtraTimeForPlanning(db, t, p2.ID)
	pts3 := saveExtraTimeForPlanning(db, t, p3.ID)
	saveExtraTimeForPlanning(db, t, p4.ID)
	latestEstimations := map[entities.PlanningID]int64{
		p1.ID: latestEstimation(pts1),
		p2.ID: latestEstimation(pts2),
		p3.ID: latestEstimation(pts3),
	}
	ps, err := st.OpenedPlannings(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 3 {
		t.Error("Invalid amount", len(ps))
	}
	var createdAt int64
	for _, p := range ps {
		if p.ID != p1.ID && p.ID != p2.ID && p.ID != p3.ID {
			t.Error("Invalid ID", p.ID)
		}
		if latestEstimations[p.ID] != p.Estimation {
			t.Error("Invalid estimation", p.Estimation, latestEstimations[p.ID])
		}
		if createdAt < p.CreatedAt {
			createdAt = p.CreatedAt
		} else {
			t.Error("Invalid sorting")
		}
	}
}

func TestSpentTimeByUserIDTimeRangeNoInfo(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	uid := ctxtg.UserID(rand.Int63())
	spent, err := st.SpentTimeByUserIDTimeRange(ctx, uid, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if spent != 0 {
		t.Error("Invalid spent", spent)
	}
}

func TestSpentTimeByUserIDTimeRange(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	uid := ctxtg.UserID(rand.Int63())
	st := NewPlanningStorage(db, second)
	p1 := entities.Planning{
		UserID:       uid,
		SpentOnline:  4,
		SpentOffline: 5,
		CreatedAt:    0,
		Status:       entities.Open,
	}
	p2 := entities.Planning{
		UserID:       uid,
		SpentOnline:  5,
		SpentOffline: 6,
		CreatedAt:    1,
		Status:       entities.Closed,
	}
	p3 := entities.Planning{
		UserID:       uid,
		SpentOnline:  6,
		SpentOffline: 7,
		CreatedAt:    2,
		Status:       entities.Open,
	}
	p4 := entities.Planning{
		UserID:       uid,
		SpentOnline:  7,
		SpentOffline: 8,
		CreatedAt:    3,
		Status:       entities.Closed,
	}
	p5 := entities.Planning{
		UserID:       uid - 1,
		SpentOnline:  6,
		SpentOffline: 7,
		CreatedAt:    2,
		Status:       entities.Open,
	}
	saveTestPlanning(db, t, p1)
	saveTestPlanning(db, t, p2)
	saveTestPlanning(db, t, p3)
	saveTestPlanning(db, t, p4)
	saveTestPlanning(db, t, p5)

	spent, err := st.SpentTimeByUserIDTimeRange(ctx, uid, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if int(spent) != p2.SpentOnline+p2.SpentOffline+p3.SpentOnline+p3.SpentOffline {
		t.Error("Invalid spent", spent)
	}
}

func latestEstimation(ps []entities.PlannedTime) int64 {
	var latest int64
	var estimation int64
	for _, p := range ps {
		if latest < p.CreatedAt {
			latest = p.CreatedAt
			estimation = p.Estimation
		}
	}
	return estimation
}

func saveExtraTimeForPlanning(db *sqlx.DB, t *testing.T, pid entities.PlanningID) []entities.PlannedTime {
	return []entities.PlannedTime{
		saveTestPlannedTime(db, t, pid),
		saveTestPlannedTime(db, t, pid),
		saveTestPlannedTime(db, t, pid),
		saveTestPlannedTime(db, t, pid),
		saveTestPlannedTime(db, t, pid),
	}
}

func saveTestPlannedTime(db *sqlx.DB, t *testing.T, pid entities.PlanningID) entities.PlannedTime {
	p := randPlannedTime()
	p.PlanningID = pid
	id, err := savePlannedTime(db, p)
	if err != nil {
		t.Fail()
	}
	p.ID = id
	return p
}

func saveTestPlanningOpened(db *sqlx.DB, t *testing.T, uid ctxtg.UserID) entities.Planning {
	p := randPlanning()
	p.Status = entities.Open
	p.UserID = uid
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	p.ID = id
	return p
}

func saveTestPlanningClosed(db *sqlx.DB, t *testing.T, uid ctxtg.UserID) entities.Planning {
	p := randPlanning()
	p.Status = entities.Closed
	p.UserID = uid
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	p.ID = id
	return p
}

func saveTestPlanning(db *sqlx.DB, t *testing.T, p entities.Planning) entities.PlanningID {
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestPlanning(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	p.ID = id
	plan, err := st.Planning(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if *plan != p {
		t.Error("Invalid value")
	}
}

func TestPlanningNoPlanning(t *testing.T) {
	defer prepareDB()()
	st := NewPlanningStorage(mysqldb.New(), second)
	p, err := st.Planning(ctx, entities.PlanningID(rand.Int63()))
	if err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Error("Should be nil")
	}
}

func TestCreatePlanning(t *testing.T) {
	defer prepareDB()()
	var now int64 = 50
	db := mysqldb.New()
	defer mockTimeNow(now)()
	st := NewPlanningStorage(db, second)
	p := randNewPlanning()
	id, err := st.CreatePlanning(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	var p2 planning
	err = db.Get(&p2, `SELECT * FROM Planning WHERE id=?`, id)
	if err != nil {
		t.Fatal(err)
	}
	expectedPlanning := entities.Planning{
		ID:              id,
		UserID:          p.UserID,
		Status:          entities.Open,
		ProjectID:       p.ProjectID,
		TrackerID:       p.TrackerID,
		IssueID:         p.IssueID,
		IssueTitle:      p.IssueTitle,
		IssueURL:        p.IssueURL,
		IssueEstimation: p.IssueEstimation,
		IssueDueDate:    p.IssueDueDate,
		IssueDone:       p.IssueDone,
		ActivityID:      p.ActivityID,
		CreatedAt:       now,
	}
	planning := p2.Planning
	planning.Status = entities.PlanningStatus(p2.Status)
	if planning != expectedPlanning {
		t.Errorf("Invalid planning %+v", p2)
	}

	expectedPlannedTime := entities.PlannedTime{
		PlanningID: id,
		Estimation: p.Estimation,
		Reason:     "",
		CreatedAt:  now,
	}
	var pt entities.PlannedTime
	err = db.Get(&pt, `SELECT * FROM PlannedTime WHERE id=?`, id)
	if err != nil {
		t.Fatal(err)
	}
	pt.ID = 0 //To check with expectedPlannedTime
	if pt != expectedPlannedTime {
		t.Errorf("Unexpected planned time %+v", pt)
	}
}

func TestAddExtraTimeInvalidID(t *testing.T) {
	defer prepareDB()()
	pt := randPlannedTime()
	st := NewPlanningStorage(mysqldb.New(), second)
	err := st.AddExtraTime(ctx, 1, pt)
	if err != entities.ErrInvalidPlanningID {
		t.Error("Unexpected error", err)
	}
}

func TestAddExtraTimeInvalidUserID(t *testing.T) {
	defer prepareDB()()
	var now int64 = 1
	defer mockTimeNow(now)()
	db := mysqldb.New()

	pt := randPlannedTime()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	pt.PlanningID = id
	err = st.AddExtraTime(ctx, p.UserID/2, pt)
	if err != entities.ErrInvalidUserID {
		t.Error("Unexpected error", err)
	}
	var plannedTime entities.PlannedTime
	err = db.Get(&plannedTime, `SELECT * FROM PlannedTime`)
	if err != sql.ErrNoRows {
		t.Error("Should be empty")
	}
}

func TestAddExtraTimeInvalidPlanningID(t *testing.T) {
	defer prepareDB()()
	var now int64 = 1
	defer mockTimeNow(now)()
	db := mysqldb.New()

	pt := randPlannedTime()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	pt.PlanningID = id / 2
	err = st.AddExtraTime(ctx, p.UserID, pt)
	if err != entities.ErrInvalidPlanningID {
		t.Error("Unexpected error", err)
	}

	var plannedTime entities.PlannedTime
	err = db.Get(&plannedTime, `SELECT * FROM PlannedTime`)
	if err != sql.ErrNoRows {
		t.Error("Should be empty")
	}
}

func TestAddExtraTime(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	var now int64 = 1
	defer mockTimeNow(now)()

	pt := randPlannedTime()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	pt.PlanningID = id
	err = st.AddExtraTime(ctx, p.UserID, pt)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	var plannedTime entities.PlannedTime
	err = db.Get(&plannedTime, `SELECT * FROM PlannedTime`)
	if err != nil {
		t.Fatal(err)
	}
	expectedFirstPlannedTime := entities.PlannedTime{
		PlanningID: id,
		Estimation: pt.Estimation,
		Reason:     pt.Reason,
		CreatedAt:  now,
	}
	plannedTime.ID = 0 //To compare with plannedTime
	if plannedTime != expectedFirstPlannedTime {
		t.Errorf("Invalid planned time %v", plannedTime)
	}
}

func TestAddSpentTimeInvalid(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	h := randSpentTimeHistory()
	err := st.AddSpentTime(ctx, h)
	if errors.Cause(err) != entities.ErrInvalidPlanningID {
		t.Error("Unexpected error", err)
	}
}

func TestAddSpentTimeIncPlanning(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	p.SpentOffline = 0
	p.SpentOnline = 0
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	online1 := randSpentTimeHistory()
	online1.PlanningID = id
	online1.Status = entities.Online
	online2 := randSpentTimeHistory()
	online2.PlanningID = id
	online2.Status = entities.Online

	offline1 := randSpentTimeHistory()
	offline1.PlanningID = id
	offline1.Status = entities.Offline
	offline2 := randSpentTimeHistory()
	offline2.PlanningID = id
	offline2.Status = entities.Offline

	err = st.AddSpentTime(ctx, online1)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	err = st.AddSpentTime(ctx, online2)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	err = st.AddSpentTime(ctx, offline1)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	err = st.AddSpentTime(ctx, offline2)
	if err != nil {
		t.Error("Unexpected error", err)
	}

	planning, err := findPlanning(db, id)
	if err != nil {
		t.Fatal(err)
	}
	if planning.SpentOnline != online1.Spent+online2.Spent ||
		planning.SpentOffline != offline1.Spent+offline2.Spent {
		t.Errorf("Add invalid amount to planning %+v", planning)
	}
}

func TestPlanningCreatedAtInvalidPlanningID(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	_, err := st.PlanningCreatedAt(ctx, entities.PlanningID(rand.Int63()))
	if err != entities.ErrInvalidPlanningID {
		t.Error("Unexpected error", err)
	}
}

func TestPlanningCreatedAt(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	createdAt, err := st.PlanningCreatedAt(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if createdAt != p.CreatedAt {
		t.Error("Invalid createdAt", createdAt)
	}
}

func TestAddSpentTime(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	h := randSpentTimeHistory()
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	h.PlanningID = id
	err = st.AddSpentTime(ctx, h)
	if err != nil {
		t.Error("Unexpected error", err)
	}
	var sth spentTimeHistory
	err = db.Get(&sth, `SELECT * FROM SpentTimeHistory`)
	if err != nil {
		t.Fatal(err)
	}
	expectedHistory := spentTimeHistory{
		SpentTimeHistory: entities.SpentTimeHistory{
			PlanningID: id,
			Spent:      h.Spent,
			StartedAt:  h.StartedAt,
			EndedAt:    h.EndedAt,
		},
		Status: string(h.Status),
	}
	if expectedHistory != sth {
		t.Errorf("Invalid history %+v", sth)
	}
}

func TestClosePlanningInvalidUserID(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	p.Status = entities.Open
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	var openedPlanning planning
	err = db.Get(&openedPlanning, "SELECT * FROM Planning WHERE id = ?", id)
	if err != nil {
		t.Fatal(err)
	}

	err = st.ClosePlanning(ctx, p.UserID/2, entities.PlanningReport{PlanningID: id})
	if err != entities.ErrInvalidUserID {
		t.Error("Unexpected err", err)
	}
	var shouldBeOpened planning
	err = db.Get(&shouldBeOpened, "SELECT * FROM Planning WHERE id = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	if openedPlanning != shouldBeOpened {
		t.Error("Planning modified")
	}
}

func TestClosePlanningInvalidPlanningID(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	p.Status = entities.Open
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	var openedPlanning planning
	err = db.Get(&openedPlanning, "SELECT * FROM Planning WHERE id = ?", id)
	if err != nil {
		t.Fatal(err)
	}

	err = st.ClosePlanning(ctx, p.UserID, entities.PlanningReport{PlanningID: id / 2})
	if err != entities.ErrInvalidPlanningID {
		t.Error("Unexpected err", err)
	}
	var shouldBeOpened planning
	err = db.Get(&shouldBeOpened, "SELECT * FROM Planning WHERE id = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	if openedPlanning != shouldBeOpened {
		t.Error("Planning modified")
	}
}

func TestClosePlanningClosed(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	p.Status = entities.Closed
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	err = st.ClosePlanning(ctx, p.UserID, entities.PlanningReport{PlanningID: id})
	if err != entities.ErrPlanningClosed {
		t.Error("Unexpected err", err)
	}

}

func TestClosePlanning(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	p.Status = entities.Open
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}
	var createdAt int64
	err = db.Get(&createdAt, `SELECT created_at FROM Planning WHERE id = ?`, id)
	if err != nil {
		t.Fatal(err)
	}

	p2 := randPlanning()
	p2.Status = entities.Open
	id2, err := savePlanning(db, p2)
	if err != nil {
		t.Fatal(err)
	}

	h1 := randSpentTimeHistory()
	h1.Status = entities.Online
	h1.PlanningID = id
	h2 := randSpentTimeHistory()
	h2.Status = entities.Online
	h2.PlanningID = id
	h3 := randSpentTimeHistory()
	h3.Status = entities.Offline
	h3.PlanningID = id

	h4 := randSpentTimeHistory()
	h4.Status = entities.Offline
	h4.PlanningID = id2

	err = st.AddSpentTime(ctx, h1)
	if err != nil {
		t.Fatal(err)
	}
	err = st.AddSpentTime(ctx, h2)
	if err != nil {
		t.Fatal(err)
	}
	err = st.AddSpentTime(ctx, h3)
	if err != nil {
		t.Fatal(err)
	}
	err = st.AddSpentTime(ctx, h4)
	if err != nil {
		t.Fatal(err)
	}

	report := entities.PlanningReport{
		PlanningID: id,
		Progress:   int(rand.Int31()),
		Time:       int64(rand.Int31()),
	}
	err = st.ClosePlanning(ctx, p.UserID, report)
	if err != nil {
		t.Error("Unexpected err", err)
	}

	var closedPlanning planning
	err = db.Get(&closedPlanning, "SELECT * FROM Planning WHERE id = ?", id)
	if err != nil {
		t.Fatal(err)
	}
	expectedPlanning := planning{
		Status: string(entities.Closed),
		Planning: entities.Planning{
			UserID:          p.UserID,
			ID:              id,
			ProjectID:       p.ProjectID,
			TrackerID:       p.TrackerID,
			IssueID:         p.IssueID,
			IssueTitle:      p.IssueTitle,
			IssueURL:        p.IssueURL,
			IssueDueDate:    p.IssueDueDate,
			IssueEstimation: p.IssueEstimation,
			IssueDone:       report.Progress,
			ActivityID:      p.ActivityID,
			SpentOnline:     h1.Spent + h2.Spent,
			SpentOffline:    h3.Spent,
			Reported:        report.Time,
			CreatedAt:       createdAt,
		},
	}
	if expectedPlanning != closedPlanning {
		t.Errorf("Invalid closed planning %+v %+v", closedPlanning, expectedPlanning)
	}
}

func TestLastActivityZero(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	_, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}

	ts, err := st.LastActivity(ctx, p.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if ts != 0 {
		t.Error("Invalid timestamp", ts)
	}
}

func TestLastActivity(t *testing.T) {
	defer prepareDB()()
	db := mysqldb.New()
	st := NewPlanningStorage(db, second)
	p := randPlanning()
	id, err := savePlanning(db, p)
	if err != nil {
		t.Fatal(err)
	}

	h1 := randSpentTimeHistory()
	h1.EndedAt = 1
	h1.PlanningID = id
	h2 := randSpentTimeHistory()
	h2.EndedAt = 2
	h2.PlanningID = id
	h3 := randSpentTimeHistory()
	h3.EndedAt = 3
	h3.PlanningID = id
	h4 := randSpentTimeHistory()
	h4.EndedAt = 4
	h4.PlanningID = id

	if err := st.AddSpentTime(ctx, h1); err != nil {
		t.Fatal(err)
	}
	if err := st.AddSpentTime(ctx, h2); err != nil {
		t.Fatal(err)
	}
	if err := st.AddSpentTime(ctx, h3); err != nil {
		t.Fatal(err)
	}
	if err := st.AddSpentTime(ctx, h4); err != nil {
		t.Fatal(err)
	}
	ts, err := st.LastActivity(ctx, p.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if ts != h4.EndedAt {
		t.Error("Invalid timestamp", ts)
	}
}

func randSpentTimeHistory() entities.SpentTimeHistory {
	return entities.SpentTimeHistory{
		PlanningID: entities.PlanningID(rand.Int63()),
		Spent:      int(rand.Int31() / 10),
		StartedAt:  rand.Int63(),
		EndedAt:    rand.Int63(),
		Status:     randSpentTimeStatus(),
	}
}

func randPlannedTime() entities.PlannedTime {
	return entities.PlannedTime{
		ID:         rand.Int63(),
		Estimation: int64(math.Abs(float64(rand.Int31()))),
		Reason:     randString(),
		CreatedAt:  int64(math.Abs(float64(rand.Int31()))),
	}
}

func randNewPlanning() entities.NewPlanning {
	return entities.NewPlanning{
		UserID:          ctxtg.UserID(rand.Int63()),
		ProjectID:       entities.ProjectID(rand.Int31()),
		TrackerID:       entities.TrackerID(rand.Int31()),
		IssueID:         entities.IssueID(rand.Int31()),
		IssueTitle:      randString(),
		IssueURL:        randString(),
		IssueEstimation: int64(rand.Int31()),
		IssueDueDate:    rand.Int63(),
		IssueDone:       int(rand.Int31()),
		ActivityID:      entities.ActivityID(rand.Int31()),
		Estimation:      int64(rand.Int31()),
	}
}

func randPlanning() entities.Planning {
	return entities.Planning{
		ID:              entities.PlanningID(rand.Int63()),
		UserID:          ctxtg.UserID(rand.Int63()),
		ProjectID:       entities.ProjectID(rand.Int31()),
		TrackerID:       entities.TrackerID(rand.Int31()),
		IssueID:         entities.IssueID(rand.Int31()),
		IssueTitle:      randString(),
		IssueURL:        randString(),
		IssueEstimation: int64(rand.Int31()),
		IssueDueDate:    rand.Int63(),
		ActivityID:      entities.ActivityID(rand.Int31()),
		SpentOnline:     int(rand.Int31() / 2),
		SpentOffline:    int(rand.Int31() / 2),
		Reported:        int64(rand.Int31()),
		CreatedAt:       rand.Int63(),
		Status:          randStatus(),
	}
}

func randSpentTimeStatus() entities.SpentTimeStatus {
	if 1 == rand.Int63()%2 {
		return entities.Online
	}
	return entities.Offline

}

func randStatus() entities.PlanningStatus {
	if 1 == rand.Int63()%2 {
		return entities.Open
	}
	return entities.Closed
}

func randString() string {
	return fmt.Sprint(rand.Int63())
}

func mockTimeNow(timeToReturn int64) func() {
	timeNowFunc = func() int64 { return timeToReturn }
	return func() {
		timeNowFunc = func() int64 {
			return time.Now().Unix()
		}
	}
}

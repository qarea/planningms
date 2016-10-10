package rpcsvc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/powerman/narada-go/narada"
	"github.com/qarea/ctxtg"
	"github.com/qarea/ctxtg/ctxtgtest"
	"github.com/qarea/planningms/entities"
)

func TestVERSION(t *testing.T) {
	api := &API{}
	var args struct{}
	var res string
	err := api.Version(&args, &res)
	if err != nil {
		t.Errorf("Version(), err = %v", err)
	}
	if ver, _ := narada.Version(); res != ver {
		t.Errorf("Version() = %v, want %v", res, ver)
	}
}

func TestCreatePlanningTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	var resp CreatePlanningResp
	err := api.CreatePlanning(&CreatePlanningReq{
		Context: ctx,
	}, &resp)
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestCreatePlanningServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningStorage{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningStorage: ps,
		TokenParser:     p,
	})
	var resp CreatePlanningResp
	err := api.CreatePlanning(&CreatePlanningReq{
		Context: ctx,
	}, &resp)
	if err != ps.err {
		t.Error("Service error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestCreatePlanning(t *testing.T) {
	userID := ctxtg.UserID(rand.Int63())
	ctx := testContext()
	ps := &testPlanningStorage{}
	p := &ctxtgtest.Parser{
		Claims: ctxtg.Claims{
			UserID: userID,
		},
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningStorage: ps,
		TokenParser:     p,
	})
	expectedNewPlanning := entities.NewPlanning{
		UserID:          userID,
		ProjectID:       entities.ProjectID(rand.Int63()),
		TrackerID:       entities.TrackerID(rand.Int63()),
		IssueID:         entities.IssueID(rand.Int63()),
		IssueTitle:      randomString(),
		IssueURL:        randomString(),
		IssueDueDate:    rand.Int63(),
		IssueEstimation: rand.Int63(),
		ActivityID:      entities.ActivityID(rand.Int63()),
		Estimation:      rand.Int63(),
	}
	var resp CreatePlanningResp
	err := api.CreatePlanning(&CreatePlanningReq{
		Context:         ctx,
		ProjectID:       expectedNewPlanning.ProjectID,
		TrackerID:       expectedNewPlanning.TrackerID,
		IssueID:         expectedNewPlanning.IssueID,
		IssueTitle:      expectedNewPlanning.IssueTitle,
		IssueURL:        expectedNewPlanning.IssueURL,
		ActivityID:      expectedNewPlanning.ActivityID,
		Estimation:      expectedNewPlanning.Estimation,
		IssueEstimation: expectedNewPlanning.IssueEstimation,
		IssueDueDate:    expectedNewPlanning.IssueDueDate,
	}, &resp)
	if err != ps.err {
		t.Error("Service error expected", err)
	}
	if ps.newPlanning != expectedNewPlanning {
		t.Errorf("Invalid new planning passed %+v", ps.newPlanning)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestGetPlanningsTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	var resp GetOpenPlanningsResp
	err := api.GetOpenPlannings(&GetOpenPlanningsReg{
		Context: ctx,
	}, &resp)
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestGetPlanningsServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningService{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	var resp GetOpenPlanningsResp
	err := api.GetOpenPlannings(&GetOpenPlanningsReg{
		Context: ctx,
	}, &resp)
	if err != ps.err {
		t.Error("Service error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestGetPlannings(t *testing.T) {
	claims := testClaims()
	ctx := testContext()
	ps := &testPlanningService{
		plannings: randPlannings(),
	}
	p := &ctxtgtest.Parser{
		Claims:        claims,
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	var resp GetOpenPlanningsResp
	err := api.GetOpenPlannings(&GetOpenPlanningsReg{
		Context: ctx,
	}, &resp)
	if err != nil {
		t.Fatal(err)
	}
	if ps.userID != claims.UserID {
		t.Error("Invalid user ID", ps.userID)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
	if !reflect.DeepEqual(ps.plannings, resp.Plannings) {
		t.Error("Invalid response")
	}
}

func TestSetExtraTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	err := api.SetExtra(&SetExtraReq{
		Context: ctx,
	}, nil)
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetExtraServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningStorage{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningStorage: ps,
		TokenParser:     p,
	})
	err := api.SetExtra(&SetExtraReq{
		Context: ctx,
	}, nil)
	if err != ps.err {
		t.Error("TimeManager error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetExtra(t *testing.T) {
	planningID := entities.PlanningID(rand.Int63())
	estimation := rand.Int63()
	reason := randomString()
	claims := testClaims()
	ctx := testContext()
	expectedExtra := entities.PlannedTime{
		PlanningID: planningID,
		Estimation: estimation,
		Reason:     reason,
	}
	ps := &testPlanningStorage{}
	p := &ctxtgtest.Parser{
		Claims:        claims,
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningStorage: ps,
		TokenParser:     p,
	})
	err := api.SetExtra(&SetExtraReq{
		Context:    ctx,
		PlanningID: planningID,
		Estimation: estimation,
		Reason:     reason,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ps.extraTime != expectedExtra {
		t.Error("Invalid args passed")
	}
	if ps.userID != claims.UserID {
		t.Error("Invalid args passed")
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetActiveTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	err := api.SetActive(&SetActiveReq{
		Context: ctx,
	}, nil)
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetActiveServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningService{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.SetActive(&SetActiveReq{
		Context: ctx,
	}, nil)
	if err != ps.err {
		t.Error("TimeManager error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetActive(t *testing.T) {
	ts := rand.Int63()
	planningID := entities.PlanningID(rand.Int63())
	claims := testClaims()
	ctx := testContext()
	ps := &testPlanningService{}
	p := &ctxtgtest.Parser{
		Claims:        claims,
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.SetActive(&SetActiveReq{
		Context:    ctx,
		PlanningID: planningID,
		Time:       ts,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if ps.planningID != planningID || ps.userID != claims.UserID || ps.time != ts {
		t.Error("Unexpected args")
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetSpentTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	err := api.SetSpent(&SetSpentReq{
		Context: ctx,
	}, nil)
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetSpentServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningService{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.SetSpent(&SetSpentReq{
		Context: ctx,
	}, nil)
	if err != ps.err {
		t.Error("TimeManager error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSetSpent(t *testing.T) {
	planningID := entities.PlanningID(rand.Int63())
	claims := testClaims()
	ctx := testContext()
	spent := rand.Int()
	expectedSpentTime := entities.SpentTimeReport{
		UserID:     claims.UserID,
		PlanningID: planningID,
		Spent:      spent,
	}
	ps := &testPlanningService{}
	p := &ctxtgtest.Parser{
		Claims:        claims,
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.SetSpent(&SetSpentReq{
		Context:    ctx,
		PlanningID: planningID,
		Spent:      spent,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if expectedSpentTime != ps.spentTime {
		t.Error("Invalid args passed")
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestClosingPlanningTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	err := api.ClosePlanning(&ClosePlanningReq{
		Context: ctx,
	}, nil)
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestClosePlanningServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningService{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.ClosePlanning(&ClosePlanningReq{
		Context: ctx,
	}, nil)
	if err != ps.err {
		t.Error("TimeManager error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestClosePlanning(t *testing.T) {
	claims := testClaims()
	ctx := testContext()
	planningID := entities.PlanningID(rand.Int63())
	progress := rand.Int()
	spent := rand.Int63()
	ps := &testPlanningService{}
	p := &ctxtgtest.Parser{
		Claims:        claims,
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.ClosePlanning(&ClosePlanningReq{
		Context:    ctx,
		PlanningID: planningID,
		Progress:   progress,
		Time:       spent,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	expectedReport := entities.PlanningReport{
		PlanningID: planningID,
		Progress:   progress,
		Time:       spent,
	}
	if ps.userID != claims.UserID || expectedReport != ps.report {
		t.Error("Invalid args passed")
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSpentTimeTokenErr(t *testing.T) {
	ctx := testContext()
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
		Err:           errors.New("parser err"),
	}
	api := newPlanningServiceRPC(RPCConfig{
		TokenParser: p,
	})
	err := api.SpentTime(&SpentTimeReq{
		Context: ctx,
	}, &SpentTimeResp{})
	if err != p.Err {
		t.Error("Parser error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSpentTimeServiceErr(t *testing.T) {
	ctx := testContext()
	ps := &testPlanningService{
		err: errors.New("Service err"),
	}
	p := &ctxtgtest.Parser{
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	err := api.SpentTime(&SpentTimeReq{
		Context: ctx,
	}, &SpentTimeResp{})
	if err != ps.err {
		t.Error("Service error expected", err)
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func TestSpentTime(t *testing.T) {
	claims := testClaims()
	ctx := testContext()
	spent := rand.Int()
	from := rand.Int63()
	to := rand.Int63()

	ps := &testPlanningService{
		spent: spent,
	}
	p := &ctxtgtest.Parser{
		Claims:        claims,
		TokenExpected: ctx.Token,
	}
	api := newPlanningServiceRPC(RPCConfig{
		PlanningService: ps,
		TokenParser:     p,
	})
	result := &SpentTimeResp{}
	err := api.SpentTime(&SpentTimeReq{
		Context: ctx,
		From:    from,
		To:      to,
	}, result)
	if err != nil {
		t.Fatal(err)
	}
	if result.Spent != spent {
		t.Error("Invalid result", result.Spent)
	}
	if ps.userID != claims.UserID || ps.from != from || ps.to != to {
		t.Error("Invalid args passed")
	}
	if err := p.Error(); err != nil {
		t.Error("Unexpected parser error", err)
	}
}

func randPlannings() []entities.ExtendedPlanning {
	var ps []entities.ExtendedPlanning
	for i := 0; i < rand.Intn(10); i++ {
		ps = append(ps, randPlanning())
	}
	return ps
}

func randPlanning() entities.ExtendedPlanning {
	return entities.ExtendedPlanning{
		Estimation: rand.Int63(),
		Planning: entities.Planning{
			ID:         entities.PlanningID(rand.Int63()),
			IssueTitle: randomString(),
		},
	}
}

func testClaims() ctxtg.Claims {
	return ctxtg.Claims{
		UserID: ctxtg.UserID(rand.Int63()),
	}
}

func testContext() ctxtg.Context {
	return ctxtg.Context{
		Token: randomToken(),
	}
}

func randomToken() ctxtg.Token {
	return ctxtg.Token(randomString())
}

func randomString() string {
	return fmt.Sprint(rand.Int63())
}

type testPlanningService struct {
	userID     ctxtg.UserID
	planningID entities.PlanningID
	extraTime  entities.PlannedTime
	report     entities.PlanningReport
	plannings  []entities.ExtendedPlanning
	spentTime  entities.SpentTimeReport
	time       int64
	from       int64
	to         int64
	spent      int

	err error
}

func (t *testPlanningService) CreatePlanningNewIssue(_ context.Context) error {
	return t.err
}

func (t *testPlanningService) SetActive(_ context.Context, a entities.NewActivePlanning) error {
	t.userID = a.UserID
	t.planningID = a.PlanningID
	t.time = a.Time
	return t.err
}

func (t *testPlanningService) AddSpentTime(_ context.Context, time entities.SpentTimeReport) error {
	t.spentTime = time
	return t.err
}

func (t *testPlanningService) ClosePlanning(_ context.Context, userID ctxtg.UserID, r entities.PlanningReport) error {
	t.report = r
	t.userID = userID
	return t.err
}
func (t *testPlanningService) OpenedPlannings(_ context.Context, uid ctxtg.UserID) ([]entities.ExtendedPlanning, error) {
	t.userID = uid
	return t.plannings, t.err
}

func (t *testPlanningService) SpentTime(_ context.Context, uid ctxtg.UserID, from, to int64) (int, error) {
	t.userID = uid
	t.from = from
	t.to = to
	return t.spent, t.err
}

type testPlanningStorage struct {
	err         error
	newPlanning entities.NewPlanning
	id          entities.PlanningID
	userID      ctxtg.UserID
	extraTime   entities.PlannedTime
}

func (t *testPlanningStorage) CreatePlanning(_ context.Context, np entities.NewPlanning) (entities.PlanningID, error) {
	t.newPlanning = np
	return t.id, t.err
}

func (t *testPlanningStorage) AddExtraTime(_ context.Context, userID ctxtg.UserID, et entities.PlannedTime) error {
	t.userID = userID
	t.extraTime = et
	return t.err
}

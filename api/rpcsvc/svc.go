// Package rpcsvc provides handlers for JSON-RPC 2.0.
package rpcsvc

import (
	"context"
	"net/http"
	"net/rpc"

	"github.com/powerman/must"
	"github.com/powerman/narada-go/narada"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/cfg"
	"github.com/qarea/planningms/entities"

	"github.com/pkg/errors"

	"github.com/powerman/rpc-codec/jsonrpc2"
)

var log = narada.NewLog("rpcsvc: ")

// Init setups and registers JSON-RPC handlers
func Init(c RPCConfig) {
	must.AbortIf(rpc.Register(newPlanningServiceRPC(c)))
	http.Handle(cfg.HTTP.BasePath+"/rpc", jsonrpc2.HTTPHandler(nil))
}

// RPCConfig is dependencies configuration for rpcsvc
type RPCConfig struct {
	TokenParser     ctxtg.TokenParser
	PlanningService PlanningService
	PlanningStorage PlanningStorage
}

func newPlanningServiceRPC(c RPCConfig) *API {
	return &API{
		tokenParser:     c.TokenParser,
		planningService: c.PlanningService,
		planningStorage: c.PlanningStorage,
	}
}

// API struct for JSON-RPC 2.0
type API struct {
	tokenParser     ctxtg.TokenParser
	planningService PlanningService
	planningStorage PlanningStorage
}

// PlanningService is required dependency for API
type PlanningService interface {
	ClosePlanning(context.Context, ctxtg.UserID, entities.PlanningReport) error
	SetActive(context.Context, entities.NewActivePlanning) error
	AddSpentTime(context.Context, entities.SpentTimeReport) error
	OpenedPlannings(context.Context, ctxtg.UserID) ([]entities.ExtendedPlanning, error)
	SpentTime(context.Context, ctxtg.UserID, int64, int64) (int, error)
}

// PlanningStorage is required dependency for API
type PlanningStorage interface {
	CreatePlanning(context.Context, entities.NewPlanning) (entities.PlanningID, error)
	AddExtraTime(context.Context, ctxtg.UserID, entities.PlannedTime) error
}

// Version returns current project narada version
func (*API) Version(args *struct{}, res *string) error {
	log.DEBUG("RPC: VERSION")
	*res, _ = narada.Version()
	return nil
}

// CreatePlanningReq is input argument to CreatePlanning
type CreatePlanningReq struct {
	Context         ctxtg.Context
	ProjectID       entities.ProjectID
	TrackerID       entities.TrackerID
	IssueID         entities.IssueID
	IssueTitle      string
	IssueURL        string
	IssueEstimation int64
	IssueDueDate    int64
	IssueDone       int
	ActivityID      entities.ActivityID
	Estimation      int64
}

// CreatePlanningResp is response from CreatePlanning
type CreatePlanningResp struct {
	PlanningID entities.PlanningID
}

// CreatePlanning creates new planning and return PlanningID
func (p *API) CreatePlanning(req *CreatePlanningReq, resp *CreatePlanningResp) error {
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		id, err := p.planningStorage.CreatePlanning(ctx, entities.NewPlanning{
			UserID:          c.UserID,
			ProjectID:       req.ProjectID,
			TrackerID:       req.TrackerID,
			IssueID:         req.IssueID,
			IssueTitle:      req.IssueTitle,
			IssueURL:        req.IssueURL,
			IssueEstimation: req.IssueEstimation,
			IssueDueDate:    req.IssueDueDate,
			IssueDone:       req.IssueDone,
			ActivityID:      req.ActivityID,
			Estimation:      req.Estimation,
		})
		*resp = CreatePlanningResp{
			PlanningID: id,
		}
		return err
	})
	return errWithLog(req.Context, "failed to CreatePlanning", err)
}

// GetOpenPlanningsReg is input parameter to GetOpenPlannings
type GetOpenPlanningsReg struct {
	Context ctxtg.Context
}

// GetOpenPlanningsResp is output from GetOpenPlannings
type GetOpenPlanningsResp struct {
	Plannings []entities.ExtendedPlanning
}

// GetOpenPlannings returns all open plannings for user
func (p *API) GetOpenPlannings(req *GetOpenPlanningsReg, resp *GetOpenPlanningsResp) error {
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		ps, err := p.planningService.OpenedPlannings(ctx, c.UserID)
		*resp = GetOpenPlanningsResp{
			Plannings: ps,
		}
		return err
	})
	return errWithLog(req.Context, "failed to GetOpenPlannings", err)
}

// SetExtraReq is input parameter to SetExtra
type SetExtraReq struct {
	Context    ctxtg.Context
	PlanningID entities.PlanningID
	Estimation int64
	Reason     string
}

// SetExtra add extra time to estimation of planning
func (p *API) SetExtra(req *SetExtraReq, _ *struct{}) error {
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		return p.planningStorage.AddExtraTime(ctx, c.UserID, entities.PlannedTime{
			PlanningID: req.PlanningID,
			Estimation: req.Estimation,
			Reason:     req.Reason,
		})
	})
	return errWithLog(req.Context, "failed to SetExtra", err)
}

// SetActiveReq is input parameter to SetActive
type SetActiveReq struct {
	Context    ctxtg.Context
	PlanningID entities.PlanningID
	Time       int64
}

// SetActive set active planning for user
func (p *API) SetActive(req *SetActiveReq, resp *struct{}) error {
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		return p.planningService.SetActive(ctx, entities.NewActivePlanning{
			UserID:     c.UserID,
			PlanningID: req.PlanningID,
			Time:       req.Time,
		})
	})
	return errWithLog(req.Context, "failed to SetActive", err)
}

// SetSpentReq is input parameter ti SetSpent
type SetSpentReq struct {
	Context    ctxtg.Context
	PlanningID entities.PlanningID
	Spent      int
	Time       int64
}

// SetSpent add spent time for planning
func (p *API) SetSpent(req *SetSpentReq, resp *struct{}) error {
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		return p.planningService.AddSpentTime(ctx, entities.SpentTimeReport{
			UserID:     c.UserID,
			PlanningID: req.PlanningID,
			Spent:      req.Spent,
			Time:       req.Time,
		})
	})
	return errWithLog(req.Context, "failed to SetSpent", err)
}

// ClosePlanningReq is input parameter to ClosePlanning
type ClosePlanningReq struct {
	Context    ctxtg.Context
	PlanningID entities.PlanningID
	Progress   int
	Time       int64
}

// ClosePlanning close planning and registers additional info about it
func (p *API) ClosePlanning(req *ClosePlanningReq, _ *struct{}) error {
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		return p.planningService.ClosePlanning(ctx, c.UserID, entities.PlanningReport{
			PlanningID: req.PlanningID,
			Progress:   req.Progress,
			Time:       req.Time,
		})
	})
	return errWithLog(req.Context, "failed to ClosePlanning", err)
}

// SpentTimeReq is input parameter to SpentTime
type SpentTimeReq struct {
	Context ctxtg.Context
	From    int64
	To      int64
}

// SpentTimeResp is output from SpentTime
type SpentTimeResp struct {
	Spent int
}

// SpentTime returns user's spent time for period
func (p *API) SpentTime(req *SpentTimeReq, resp *SpentTimeResp) error {
	var spent int
	err := p.tokenParser.ParseCtxWithClaims(req.Context, func(ctx context.Context, c ctxtg.Claims) error {
		var err error
		spent, err = p.planningService.SpentTime(ctx, c.UserID, req.From, req.To)
		return err
	})
	*resp = SpentTimeResp{
		Spent: spent,
	}
	return errWithLog(req.Context, "failed to ClosePlanning", err)

}

func errWithLog(ctx ctxtg.Context, prefix string, err error) error {
	if err == nil {
		return nil
	}
	log.ERR("tracking id: %s, token: %s, %s: %+v", ctx.TracingID, ctx.Token, prefix, err)
	err = errors.Cause(err)
	if err == context.DeadlineExceeded {
		return entities.ErrTimeout
	}
	return err
}

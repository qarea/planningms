package entities

import "github.com/powerman/rpc-codec/jsonrpc2"

//Global error codes
var (
	ErrTimeout     = jsonrpc2.NewError(0, "TIMEOUT")
	ErrMaintenance = jsonrpc2.NewError(4, "MAINTENANCE")
)

//Application specific codes
var (
	ErrInvalidPlanningID = jsonrpc2.NewError(100, "INVALID_PLANNING_ID")
	ErrNoActivePlanning  = jsonrpc2.NewError(101, "NO_ACTIVE_PLANNING")
	ErrInvalidUserID     = jsonrpc2.NewError(102, "INVALID_USER_ID")
	ErrPlanningClosed    = jsonrpc2.NewError(103, "PLANNING_CLOSED")
	ErrPlanningOutdated  = jsonrpc2.NewError(104, "PLANNING_OUTDATED")
	ErrOutdatedReport    = jsonrpc2.NewError(105, "OUTDATED_REPORT")
	ErrNegativeSpentTime = jsonrpc2.NewError(106, "NEGATIVE_SPENT_TIME")
)

package entities

import "github.com/qarea/ctxtg"

// Planning represents user plan for today
type Planning struct {
	ID              PlanningID     `db:"id"`
	UserID          ctxtg.UserID   `db:"user_id"`
	Status          PlanningStatus `db:"-"`
	ProjectID       ProjectID      `db:"project_id"`
	TrackerID       TrackerID      `db:"tracker_id"`
	IssueID         IssueID        `db:"issue_id"`
	IssueTitle      string         `db:"issue_title"`
	IssueURL        string         `db:"issue_url"`
	IssueEstimation int64          `db:"issue_estim"`
	IssueDueDate    int64          `db:"issue_due_date"`
	IssueDone       int            `db:"issue_done"`
	ActivityID      ActivityID     `db:"activity_id"`
	SpentOnline     int            `db:"spent_online"`
	SpentOffline    int            `db:"spent_offline"`
	Reported        int64          `db:"reported"`
	CreatedAt       int64          `db:"created_at"`
}

// ExtendedPlanning is Planning with additional information
type ExtendedPlanning struct {
	Planning
	Estimation   int64
	LastActivity int64
	Outdated     bool
}

// NewPlanning is representation of new user plan for today
type NewPlanning struct {
	UserID          ctxtg.UserID
	ProjectID       ProjectID
	TrackerID       TrackerID
	IssueID         IssueID
	IssueTitle      string
	IssueURL        string
	IssueEstimation int64
	IssueDueDate    int64
	IssueDone       int
	ActivityID      ActivityID
	Estimation      int64
}

// PlannedTime represents expected time to spent on PlanningID by user
type PlannedTime struct {
	ID         int64      `db:"id"`
	PlanningID PlanningID `db:"planning_id"`
	Estimation int64      `db:"estimation"`
	Reason     string     `db:"reason"`
	CreatedAt  int64      `db:"created_at"`
}

// SpentTime represents user's spent time on planning
type SpentTime struct {
	UserID            ctxtg.UserID
	PlanningID        PlanningID
	PlanningCreatedAt int64
	Started           int64
	Last              int64
	SpentOnline       int
}

// SpentTimeReport represents spent time on planning report
type SpentTimeReport struct {
	UserID     ctxtg.UserID
	PlanningID PlanningID
	Spent      int
	Time       int64
}

// SpentTimeHistory represents small amount of history
type SpentTimeHistory struct {
	PlanningID PlanningID      `db:"planning_id"`
	Spent      int             `db:"spent"`
	StartedAt  int64           `db:"started_at"`
	EndedAt    int64           `db:"ended_at"`
	Status     SpentTimeStatus `db:"-"`
}

// PlanningReport represents final report in the end of planning
type PlanningReport struct {
	PlanningID PlanningID
	Progress   int
	Time       int64
}

// NewActivePlanning represents new active planning for user
type NewActivePlanning struct {
	UserID     ctxtg.UserID
	PlanningID PlanningID
	Time       int64
}

// ModifySpentTimeFunc is function to modify SpentTime for user
type ModifySpentTimeFunc func(*SpentTime) (*SpentTime, error)

// PlanningID is helper type to avoid invalid int usage
type PlanningID int64

// ProjectID is helper type to avoid invalid int usage
type ProjectID int64

// TrackerID is helper type to avoid invalid int usage
type TrackerID int64

// IssueID is helper type to avoid invalid int usage
type IssueID int64

// ActivityID is helper type to avoid invalid int usage
type ActivityID int64

// Status is helper type to avoid invalid string usage
type Status string

// SpentTimeStatus is type for online or offline spent time status
type SpentTimeStatus string

// Available spent time statuses
const (
	Online  SpentTimeStatus = "ONLINE"
	Offline SpentTimeStatus = "OFFLINE"
)

// PlanningStatus is type for open or closed planning statuses
type PlanningStatus string

// Available planning statuses
const (
	Open   PlanningStatus = "OPEN"
	Closed PlanningStatus = "CLOSED"
)

package storage

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

const (
	savePlanningsStmt = `
		INSERT INTO Planning (user_id,
							  status,
							  project_id,
							  tracker_id,
							  issue_id,
							  issue_title,
							  issue_url,
							  issue_estim,
							  issue_due_date,
							  issue_done,
							  activity_id,
							  spent_online,
							  spent_offline,
							  reported,
							  created_at)
		VALUES				 (:user_id,
							  :status,
							  :project_id,
                              :tracker_id,
                              :issue_id,
                              :issue_title,
                              :issue_url,
							  :issue_estim,
							  :issue_due_date,
							  :issue_done,
                              :activity_id,
                              :spent_online,
                              :spent_offline,
                              :reported,
                              :created_at)
	`
	updatePlanningsStmt = `
		UPDATE Planning
		   SET user_id       = :user_id,
			   status        = :status,
			   project_id    = :project_id,
               tracker_id    = :tracker_id,
               issue_id      = :issue_id,
               issue_title   = :issue_title,
               issue_url     = :issue_url,
               issue_done    = :issue_done,
               activity_id   = :activity_id,
               spent_online  = :spent_online,
               spent_offline = :spent_offline,
               reported      = :reported,
               created_at    = :created_at
         WHERE id            = :id
	`
	findPlanningByIDStmt = `
		SELECT *
		  FROM Planning
		 WHERE id = ?
	`
	createdAtStmt = `
		SELECT created_at
		  FROM Planning
		 WHERE id = ?
	`
	incrementOnlineStmt = `
		UPDATE Planning
		   SET spent_online = spent_online + ?
		 WHERE id = ?
	`
	incrementOfflineStmt = `
		UPDATE Planning
		   SET spent_offline = spent_offline + ?
		 WHERE id = ?
	`
	openedPlanningsStmt = `
		SELECT *
		  FROM Planning
		 WHERE user_id = ?
           AND status = "OPEN"
         ORDER BY created_at ASC
	`
	spentSumStmt = `
	 SELECT SUM(spent_online) + SUM(spent_offline) AS sum
	   FROM Planning
	  WHERE user_id = ?
		AND created_at >= ?
		AND created_at <= ?
	`
)

type planning struct {
	entities.Planning
	Status string `db:"status"`
}

func spentTime(ex sqlx.Ext, uid ctxtg.UserID, from, to int64) (int, error) {
	var spent sql.NullInt64
	err := sqlx.Get(ex, &spent, spentSumStmt, uid, from, to)
	return int(spent.Int64), err
}

func openedPlannings(ex sqlx.Ext, uid ctxtg.UserID) ([]entities.Planning, error) {
	var plannings []planning
	err := sqlx.Select(ex, &plannings, openedPlanningsStmt, uid)
	if err != nil {
		return nil, err
	}
	var ps []entities.Planning
	for _, p := range plannings {
		ps = append(ps, fromDBPlanning(p))
	}
	return ps, nil
}

func addSpentTimeToPlanning(ex sqlx.Ext, h entities.SpentTimeHistory) error {
	switch h.Status {
	case entities.Online:
		_, err := ex.Exec(incrementOnlineStmt, h.Spent, h.PlanningID)
		return err
	case entities.Offline:
		_, err := ex.Exec(incrementOfflineStmt, h.Spent, h.PlanningID)
		return err
	}
	return errors.New("invalid status")
}

func savePlanning(ex sqlx.Ext, p entities.Planning) (entities.PlanningID, error) {
	planningDB := toDBPlanning(p)
	res, err := sqlx.NamedExec(ex, savePlanningsStmt, planningDB)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return entities.PlanningID(id), nil
}

func findPlanning(ex sqlx.Ext, pid entities.PlanningID) (*entities.Planning, error) {
	row := ex.QueryRowx(findPlanningByIDStmt, pid)
	var p planning
	err := row.StructScan(&p)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	planning := fromDBPlanning(p)
	return &planning, err
}

func updatePlanning(ex sqlx.Ext, p entities.Planning) error {
	_, err := sqlx.NamedExec(ex, updatePlanningsStmt, toDBPlanning(p))
	return err
}

func planningCreatedAt(ex sqlx.Ext, pid entities.PlanningID) (int64, error) {
	row := ex.QueryRowx(createdAtStmt, int64(pid))
	var createdAt int64
	err := row.Scan(&createdAt)
	if err == sql.ErrNoRows {
		return 0, entities.ErrInvalidPlanningID
	}
	return createdAt, err
}

func toDBPlanning(p entities.Planning) planning {
	return planning{
		Planning: p,
		Status:   string(p.Status),
	}
}

func fromDBPlanning(p planning) entities.Planning {
	planning := p.Planning
	planning.Status = entities.PlanningStatus(p.Status)
	return planning
}

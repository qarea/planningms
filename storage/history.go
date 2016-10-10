package storage

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/entities"
)

const (
	saveSpentTimeHistoryStmt = `
		INSERT INTO SpentTimeHistory (planning_id,
									  status,
									  spent,
									  started_at,
									  ended_at)
		VALUES						 (:planning_id,
									  :status,
									  :spent,
                                      :started_at,
                                      :ended_at)
	`
	findHistoriesByPlanningID = `
		SELECT *
		  FROM SpentTimeHistory
		 WHERE planning_id = ?
	`
	findLastActivityStmt = `
		SELECT ended_at
		  FROM Planning AS p INNER JOIN SpentTimeHistory AS s
			ON p.id = s.planning_id
		 WHERE p.user_id = ?
		 ORDER BY ended_at DESC
		 LIMIT 1
	`
	findLastActivitiesStmt = `
		SELECT planning_id, MAX(ended_at) AS ended_at
		  FROM SpentTimeHistory
		 WHERE planning_id IN (?)
		 GROUP BY planning_id
	`
)

type spentTimeHistory struct {
	entities.SpentTimeHistory
	Status string `db:"status"`
}

func saveHistory(ex sqlx.Ext, h entities.SpentTimeHistory) (int64, error) {
	sth := spentTimeHistory{
		SpentTimeHistory: h,
		Status:           string(h.Status),
	}
	res, err := sqlx.NamedExec(ex, saveSpentTimeHistoryStmt, sth)
	if err, ok := err.(*mysql.MySQLError); ok {
		if err.Number == mysqlForeignKeyErrorCode {
			return 0, entities.ErrInvalidPlanningID
		}
	}
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func findHistories(ex sqlx.Ext, pid entities.PlanningID) ([]entities.SpentTimeHistory, error) {
	var histories []spentTimeHistory
	err := sqlx.Select(ex, &histories, findHistoriesByPlanningID, pid)
	if err != nil {
		return nil, err
	}
	var hs []entities.SpentTimeHistory
	for _, h := range histories {
		h.SpentTimeHistory.Status = entities.SpentTimeStatus(h.Status)
		hs = append(hs, h.SpentTimeHistory)
	}
	return hs, nil
}

func lastActivityForUser(ex sqlx.Ext, uid ctxtg.UserID) (int64, error) {
	var lastActivity int64
	err := sqlx.Get(ex, &lastActivity, findLastActivityStmt, uid)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return lastActivity, err
}

func lastActivityForPlannings(ex sqlx.Ext, ps []entities.Planning) (map[entities.PlanningID]int64, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	var ids []entities.PlanningID
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	q, args, err := sqlx.In(findLastActivitiesStmt, ids)
	if err != nil {
		return nil, err
	}
	var results []struct {
		PlanningID entities.PlanningID `db:"planning_id"`
		EndedAt    int64               `db:"ended_at"`
	}
	err = sqlx.Select(ex, &results, q, args...)
	activities := map[entities.PlanningID]int64{}
	for _, v := range results {
		activities[v.PlanningID] = v.EndedAt
	}
	return activities, err
}

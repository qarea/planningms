package storage

import (
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/qarea/planningms/entities"
)

const (
	savePlannedTimeStmt = `
		INSERT INTO PlannedTime (planning_id,
							     estimation,
							     reason,
							     created_at)
	    VALUES				    (:planning_id,
							     :estimation, 
							     :reason,     
                                 :created_at) 
	`
	latestEstimationStmt = `
		SELECT pt.planning_id, pt.estimation
          FROM PlannedTime AS pt
         INNER JOIN (
	                 SELECT planning_id, MAX(created_at) AS created_at
                       FROM PlannedTime
                      WHERE planning_id IN (?)
	                  GROUP BY planning_id
	     ) AS latest 
	        ON pt.planning_id = latest.planning_id 
	       AND pt.created_at = latest.created_at`
)

func estimationsForPlannings(ex sqlx.Ext, ps []entities.Planning) (map[entities.PlanningID]int64, error) {
	if len(ps) == 0 {
		return nil, nil
	}
	var ids []entities.PlanningID
	for _, p := range ps {
		ids = append(ids, p.ID)
	}
	q, args, err := sqlx.In(latestEstimationStmt, ids)
	if err != nil {
		return nil, err
	}
	var results []struct {
		PlanningID entities.PlanningID `db:"planning_id"`
		Estimation int64               `db:"estimation"`
	}
	err = sqlx.Select(ex, &results, q, args...)
	ests := map[entities.PlanningID]int64{}
	for _, v := range results {
		ests[v.PlanningID] = v.Estimation
	}
	return ests, err
}

func savePlannedTime(ex sqlx.Ext, p entities.PlannedTime) (int64, error) {
	res, err := sqlx.NamedExec(ex, savePlannedTimeStmt, p)
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

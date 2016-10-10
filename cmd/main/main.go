// Package main provides main service.
package main

import (
	"net/http"

	"github.com/powerman/narada-go/narada/bootstrap"
	"github.com/qarea/ctxtg"
	"github.com/qarea/planningms/api/rpcsvc"
	"github.com/qarea/planningms/cache"
	"github.com/qarea/planningms/cfg"
	"github.com/qarea/planningms/mysqldb"
	"github.com/qarea/planningms/plannings"
	"github.com/qarea/planningms/storage"

	"github.com/powerman/narada-go/narada"
	"github.com/prometheus/client_golang/prometheus"

	_ "github.com/go-sql-driver/mysql"
)

var log = narada.NewLog("")

func main() {
	db := mysqldb.New()

	parser, err := ctxtg.NewRSATokenParser(cfg.RSAPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	spentTimeStorage := cache.NewSpentTimeInMemory(
		cfg.TimeSpent.Frequency,
		cfg.TimeSpent.Folder,
		cfg.LockTimeout,
	)

	planningStorage := storage.NewPlanningStorage(
		db,
		cfg.LockTimeout,
	)

	svc := plannings.NewService(plannings.PlanningServiceCfg{
		SpentTimeStorage:        spentTimeStorage,
		PlanningStorage:         planningStorage,
		MaxPlanningAge:          cfg.Plannings.MaxAge,
		MaxPeriodFromLastUpdate: cfg.Plannings.OldestLastUpdate,
	})

	rpcsvc.Init(rpcsvc.RPCConfig{
		TokenParser:     parser,
		PlanningService: svc,
		PlanningStorage: planningStorage,
	})

	if err := bootstrap.Unlock(); err != nil {
		log.Fatal(err)
	}
	http.Handle(cfg.HTTP.BasePath+"/metrics", prometheus.Handler())
	log.NOTICE("Listening on %s", cfg.HTTP.Listen+cfg.HTTP.BasePath)
	log.Fatal(http.ListenAndServe(cfg.HTTP.Listen, nil))
}

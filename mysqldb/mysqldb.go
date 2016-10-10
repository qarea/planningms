package mysqldb

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/qarea/planningms/cfg"
)

//New creates new database connection for mysql database
func New() *sqlx.DB {
	return sqlx.MustConnect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.MySQL.Login,
		cfg.MySQL.Password,
		cfg.MySQL.Host,
		cfg.MySQL.Port,
		cfg.MySQL.DB,
	))
}

package oracle

import (
	"fmt"
	"time"

	_ "github.com/godror/godror"
	"github.com/jmoiron/sqlx"

	"github.com/presnalex/go-micro/v3/service"
)

func Connect(cfg *service.OracleConfig) (*sqlx.DB, error) {
	// format connection string
	connStr := fmt.Sprintf("%s/%s@%s/%s", cfg.Login, cfg.Passw, cfg.Addr, cfg.DBName)

	db, err := sqlx.Connect("godror", connStr)
	if err != nil {
		return nil, err
	}

	if cfg.ConnMax > 0 {
		db.SetMaxOpenConns(int(cfg.ConnMax))
	}
	if cfg.ConnMaxIdle > 0 {
		db.SetMaxIdleConns(int(cfg.ConnMaxIdle))
	}
	if cfg.ConnLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnLifetime) * time.Second)
	}
	if cfg.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}

	return db, nil
}

package postgres

import (
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/presnalex/go-micro/v3/service"
)

func Connect(cfg *service.PostgresConfig) (*sqlx.DB, error) {
	// format connection string
	dbstr := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable&statement_cache_mode=describe",
		cfg.Login,
		url.QueryEscape(cfg.Passw),
		cfg.Addr,
		cfg.DBName,
	)

	// parse connection string
	dbConf, err := pgx.ParseConfig(dbstr)
	if err != nil {
		return nil, err
	}

	// needed for pgbouncer
	dbConf.RuntimeParams = map[string]string{
		"standard_conforming_strings": "on",
		"application_name":            cfg.AppName,
	}
	// may be needed for pbbouncer, needs to check
	// dbConf.PreferSimpleProtocol = true
	// register pgx conn
	connStr := stdlib.RegisterConnConfig(dbConf)

	db, err := sqlx.Connect("pgx", connStr)
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

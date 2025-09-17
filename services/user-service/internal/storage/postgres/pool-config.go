package postgres

import (
	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newPoolConfig(cfg *config.Config) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = cfg.Postgres.MaxConns
	config.MinConns = cfg.Postgres.MinConns
	config.MaxConnLifetime = cfg.Postgres.MaxConnLifeTime
	config.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.Postgres.CheckPeriod

	return config, nil
}

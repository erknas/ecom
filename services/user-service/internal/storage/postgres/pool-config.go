package postgres

import (
	"fmt"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newPoolConfig(cfg *config.Config) (*pgxpool.Config, error) {
	connString := dsn(cfg)

	config, err := pgxpool.ParseConfig(connString)
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

func dsn(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.SSLMode,
	)
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	uniqueConstraintCode = "23505"
)

type PostgresPool struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *config.Config) (*PostgresPool, error) {
	config, err := newPoolConfig(cfg)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &PostgresPool{pool: pool}, nil
}

func (p *PostgresPool) Insert(ctx context.Context, user *models.User) (int64, error) {
	query := "INSERT INTO users(first_name, email, password_hash, created_at) VALUES($1, $2, $3, $4) RETURNING id"

	var id int64
	if err := p.pool.QueryRow(
		ctx,
		query,
		user.FirstName,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
	).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueConstraintCode {
				return 0, storage.ErrUserExists
			}
		}
		return 0, fmt.Errorf("%w: %s", storage.ErrInternalDatabase, err)
	}

	return id, nil
}

func (p *PostgresPool) UserByID(ctx context.Context, id int64) (*models.User, error) {
	query := "SELECT id, first_name, email, created_at FROM users WHERE id = $1"

	row := p.pool.QueryRow(ctx, query, id)

	user := new(models.User)
	if err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.Email,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %s", storage.ErrInternalDatabase, err)
	}

	return user, nil
}

func (p *PostgresPool) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := "SELECT id, first_name, email, password_hash, created_at FROM users WHERE email = $1"

	row := p.pool.QueryRow(ctx, query, email)

	user := new(models.User)
	if err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("%w: %s", storage.ErrInternalDatabase, err)
	}

	return user, nil
}

func (p *PostgresPool) Update(ctx context.Context, id int64, user *models.UpdatedUser) error {
	columns := []string{}
	args := []any{}
	argPos := 1

	if user.FirstName != nil {
		columns = append(columns, fmt.Sprintf("first_name = $%d", argPos))
		args = append(args, *user.FirstName)
		argPos++
	}

	if user.Email != nil {
		columns = append(columns, fmt.Sprintf("email = $%d", argPos))
		args = append(args, *user.Email)
		argPos++
	}

	if user.PasswordHash != nil {
		columns = append(columns, fmt.Sprintf("password_hash = $%d", argPos))
		args = append(args, user.PasswordHash)
		argPos++
	}

	query := "UPDATE users SET " + strings.Join(columns, ", ") + fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	_, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueConstraintCode {
				return storage.ErrUserExists
			}
		}
		return fmt.Errorf("%w: %s", storage.ErrInternalDatabase, err)
	}

	return nil
}

package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/erknas/ecom/user-service/internal/config"
	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/jackc/pgx/v5/pgxpool"
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

func (p *PostgresPool) InsertUser(ctx context.Context, user *models.User) (int64, error) {
	query := "INSERT INTO users(first_name, phone_number, email, password_hash, created_at) VALUES($1, $2, $3, $4, $5) RETURNING id"

	var id int64
	if err := p.pool.QueryRow(
		ctx,
		query,
		user.FirstName,
		user.PhoneNumber,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
	).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (p *PostgresPool) UserByID(ctx context.Context, id int64) (*models.User, error) {
	query := "SELECT id, first_name, phone_number, email, password_hash, created_at FROM users WHERE id = $1"

	row := p.pool.QueryRow(ctx, query, id)

	user := new(models.User)
	if err := row.Scan(&user.ID,
		&user.FirstName,
		&user.PhoneNumber,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

func (p *PostgresPool) UserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	query := "SELECT id, first_name, phone_number, email, password_hash, created_at FROM users WHERE phone_number = $1"

	row := p.pool.QueryRow(ctx, query, phoneNumber)

	user := new(models.User)
	if err := row.Scan(&user.ID,
		&user.FirstName,
		&user.PhoneNumber,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

func (p *PostgresPool) Update(ctx context.Context, id int64, user *models.User) error {
	var (
		query       = "UPDATE users SET "
		args        = []any{}
		argsCounter = 1
	)

	if user.FirstName != "" {
		query += fmt.Sprintf("first_name = $%d, ", argsCounter)
		args = append(args, user.FirstName)
		argsCounter++
	}

	if user.PhoneNumber != "" {
		query += fmt.Sprintf("phone_number = $%d, ", argsCounter)
		args = append(args, user.PhoneNumber)
		argsCounter++
	}

	if user.Email != "" {
		query += fmt.Sprintf("email = $%d, ", argsCounter)
		args = append(args, user.Email)
		argsCounter++
	}

	if user.PasswordHash != nil {
		query += fmt.Sprintf("password_hash = $%d, ", argsCounter)
		args = append(args, user.PasswordHash)
		argsCounter++
	}

	query = strings.Trim(query, ", ")

	query += fmt.Sprintf(" WHERE id = $%d", argsCounter)
	args = append(args, id)

	_, err := p.pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

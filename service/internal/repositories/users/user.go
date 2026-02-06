package users

import (
	"context"
	"errors"
	"fmt"
	"github.com/Parnishkaspb/ozon_posts/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	const query = `
		SELECT id, name, surname FROM users
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Query: %w", err)
	}
	defer rows.Close()

	users, err := r.collectRows(rows)

	if err != nil {
		return nil, fmt.Errorf("CollectRows: %w", err)
	}

	return users, nil
}

func (r *Repo) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	const query = `SELECT id, name, surname FROM users WHERE id = $1`

	u := new(models.User)
	err := r.pool.QueryRow(ctx, query, userID).Scan(&u.ID, &u.Name, &u.Surname)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("QueryRow Scan: %w", err)
	}
	return u, nil
}

func (r *Repo) collectRows(rows pgx.Rows) ([]*models.User, error) {
	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.User, error) {
		p := new(models.User)
		return p, row.Scan(&p.ID, &p.Name, &p.Surname)
	})

	return users, err
}

func (r *Repo) GetUserByLoginPassword(ctx context.Context, login, password string) (*models.User, error) {
	var user models.User
	const query = `SELECT id, name, surname FROM users WHERE login = $1 AND password = $2`

	err := r.pool.QueryRow(ctx, query, login, password).Scan(
		&user.ID,
		&user.Name,
		&user.Surname,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, errors.New("QueryRow Scan")
	}

	return &user, nil
}

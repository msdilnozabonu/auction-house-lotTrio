package user

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uniqueViolation = "23505"

type Repo interface {
	Create(ctx context.Context, login, password string, role string) error
	GetUserByLogin(ctx context.Context, login string) (User, error)
	GetByID(ctx context.Context, id int64) (model.User, error)
}

type repo struct {
	repo *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (Repo, error) {
	return &repo{repo: pool}, nil
}

type User struct {
	ID        int64
	Login     string
	Password  string
	Role      string
	CreatedAt time.Time
}

func (r *repo) Create(ctx context.Context, login, password string, role string) error {
	_, err := r.repo.Exec(ctx, `INSERT INTO users (login, password_hash, role) values ($1, $2, $3)`,
		login, password, role)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
			return model.ErrUserAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *repo) GetUserByLogin(ctx context.Context, login string) (User, error) {
	var user User
	err := r.repo.QueryRow(ctx, `SELECT id, login, password_hash, role, created_at FROM users WHERE login = $1`, login).
		Scan(&user.ID, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, model.ErrUserNotFound
		}
		return user, fmt.Errorf("get user by login: %w", err)
	}
	return user, nil
}

func (r *repo) GetByID(ctx context.Context, id int64) (model.User, error) {
	var user model.User
	err := r.repo.QueryRow(ctx,
		`SELECT id, login, role, created_at
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Login, &user.Role, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, model.ErrUserNotFound
		}
		return user, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

package user

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo interface {
	Create(ctx context.Context, login, password string, role string) error
	GetUserByLogin(ctx context.Context, login string) (User, error)
}

type repo struct {
	repo *pgxpool.Pool
}

func New(ctx context.Context, pool *pgxpool.Pool) (Repo, error) {
	return &repo{repo: pool}, nil
}

type User struct {
	ID        int
	Login     string
	Password  string
	Role      string
	CreatedAt time.Time
}

func (r *repo) Create(ctx context.Context, login, password string, role string) error {
	_, err := r.repo.Exec(ctx, `INSERT INTO users (login, password_hash, role) values ($1, $2, $3)`,
		login, password, role)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) GetUserByLogin(ctx context.Context, login string) (User, error) {
	var user User
	err := r.repo.QueryRow(ctx, `SELECT id, login, password_hash, role, created_at FROM users WHERE login = $1`, login).
		Scan(&user.ID, &user.Login, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, model.ErrIncorrectPassword
		}
		return user, err
	}
	return user, nil
}

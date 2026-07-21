package session

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	ID               int64
	UserID           int64
	RefreshTokenHash string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ExpiresAt        time.Time
}

type Repo interface {
	CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (int64, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	RotateSessionToken(ctx context.Context, sessionID int64, newTokenHash string, newExpiresAt time.Time) error
	DeleteSession(ctx context.Context, tokenHash string) error
	GetUserRole(ctx context.Context, userID int64) (string, error)
}

type repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repo {
	return &repo{pool: pool}
}

func (r *repo) CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sessions (user_id, refresh_token, expires_at) VALUES ($1, $2, $3) RETURNING id`,
		userID, tokenHash, expiresAt).Scan(&id)
	return id, err
}

func (r *repo) GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	var s Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, refresh_token, created_at, updated_at, expires_at
		 FROM sessions WHERE refresh_token = $1`, tokenHash).
		Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.CreatedAt, &s.UpdatedAt, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return s, model.ErrSessionNotFound
		}
		return s, err
	}
	return s, nil
}

func (r *repo) RotateSessionToken(ctx context.Context, sessionID int64,
	newTokenHash string, newExpiresAt time.Time) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE sessions
     SET refresh_token = $1, expires_at = $2, updated_at = now()
     WHERE id = $3`,
		newTokenHash, newExpiresAt, sessionID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return model.ErrSessionNotFound
	}
	return nil
}

func (r *repo) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE refresh_token = $1`, tokenHash)
	return err
}

func (r *repo) GetUserRole(ctx context.Context, userID int64) (string, error) {
	var role string
	err := r.pool.QueryRow(ctx, `SELECT role FROM users WHERE id = $1`, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", model.ErrUserNotFound
		}
		return "", err
	}
	return role, nil
}

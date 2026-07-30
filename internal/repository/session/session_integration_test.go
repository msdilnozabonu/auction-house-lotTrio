//go:build integration

package session

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)

	return pool
}

func createUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool) int64 {
	t.Helper()
	login := fmt.Sprintf("user_%d", time.Now().UnixNano())

	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO users(login, password_hash, role) VALUES ($1, 'hash', 'bidder') RETURNING id`,
		login).Scan(&id)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func TestCreateSession_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	userID := createUser(t, ctx, pool)
	expires := time.Now().Add(24 * time.Hour).UTC()
	token := fmt.Sprintf("token_%d", time.Now().UnixNano())
	sessionID, err := repo.CreateSession(ctx, userID, token, expires)

	require.NoError(t, err)
	require.NotZero(t, sessionID)

	var dbToken string
	err = pool.QueryRow(ctx,
		`SELECT refresh_token FROM sessions WHERE id = $1`,
		sessionID,
	).Scan(&dbToken)

	require.NoError(t, err)
	require.Equal(t, token, dbToken)
}

func TestGetSessionByTokenHash_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	userID := createUser(t, ctx, pool)
	token := fmt.Sprintf("token_%d", time.Now().UnixNano())
	expires := time.Now().Add(time.Hour)
	sessionID, err := repo.CreateSession(ctx, userID, token, expires)
	require.NoError(t, err)

	session, err := repo.GetSessionByTokenHash(ctx, token)

	require.NoError(t, err)
	require.Equal(t, sessionID, session.ID)
	require.Equal(t, userID, session.UserID)
	require.Equal(t, token, session.RefreshTokenHash)
}

func TestGetSessionByTokenHash_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	_, err := repo.GetSessionByTokenHash(ctx, "does-not-exist")
	require.ErrorIs(t, err, model.ErrSessionNotFound)
}

func TestRotateSessionToken_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	userID := createUser(t, ctx, pool)
	oldToken := fmt.Sprintf("old_%d", time.Now().UnixNano())
	id, err := repo.CreateSession(ctx, userID, oldToken, time.Now().Add(time.Hour))
	require.NoError(t, err)

	newExpires := time.Now().Add(48 * time.Hour)
	newToken := fmt.Sprintf("new_%d", time.Now().UnixNano())
	err = repo.RotateSessionToken(ctx, id, newToken, newExpires)
	require.NoError(t, err)

	session, err := repo.GetSessionByTokenHash(ctx, newToken)

	require.NoError(t, err)
	require.Equal(t, id, session.ID)
	require.Equal(t, newToken, session.RefreshTokenHash)
}

func TestRotateSessionToken_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	err := repo.RotateSessionToken(ctx, 999999,"token", time.Now())
	require.ErrorIs(t, err, model.ErrSessionNotFound)
}

func TestDeleteSession_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	userID := createUser(t, ctx, pool)
	deleteToken := fmt.Sprintf("delete_%d", time.Now().UnixNano())
	_, err := repo.CreateSession(ctx, userID, deleteToken, time.Now().Add(time.Hour))
	require.NoError(t, err)

	err = repo.DeleteSession(ctx, deleteToken)
	require.NoError(t, err)

	_, err = repo.GetSessionByTokenHash(ctx, deleteToken)
	require.ErrorIs(t, err, model.ErrSessionNotFound)
}

func TestGetUserRole_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	userID := createUser(t, ctx, pool)
	role, err := repo.GetUserRole(ctx, userID)

	require.NoError(t, err)
	require.Equal(t, "bidder", role)
}

func TestGetUserRole_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo := New(pool)
	role, err := repo.GetUserRole(ctx, -1)

	require.ErrorIs(t, err, model.ErrUserNotFound)
	require.Empty(t, role)
}

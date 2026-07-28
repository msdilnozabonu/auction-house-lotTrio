//go:build integration
package user

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"auction-house-lotTrio/internal/model"

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

func TestCreate_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	login := fmt.Sprintf("user_%d", time.Now().UnixNano())
	err = repo.Create(ctx, login, "hash", "bidder")
	require.NoError(t, err)

	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE login = $1`,
		login,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE login = $1`, login)
	})
}

func TestCreate_Duplicate(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	login := fmt.Sprintf("user_%d", time.Now().UnixNano())
	err = repo.Create(ctx, login, "hash", "bidder")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE login = $1`, login)
	})
	err = repo.Create(ctx, login, "hash", "bidder")
	require.ErrorIs(t, err, model.ErrUserAlreadyExists)
}

func TestGetUserByLogin_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	login := fmt.Sprintf("user_%d", time.Now().UnixNano())
	err = repo.Create(ctx, login, "hash", "bidder")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE login = $1`, login)
	})

	user, err := repo.GetUserByLogin(ctx, login)
	require.NoError(t, err)
	require.Equal(t, login, user.Login)
	require.Equal(t, "hash", user.Password)
	require.Equal(t, "bidder", user.Role)
	require.NotZero(t, user.ID)
}

func TestGetUserByLogin_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	_, err = repo.GetUserByLogin(ctx, "does-not-exist")
	require.ErrorIs(t, err, model.ErrUserNotFound)
}

func TestGetByID_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	login := fmt.Sprintf("user_%d", time.Now().UnixNano())
	err = repo.Create(ctx, login, "hash", "bidder")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE login = $1`, login)
	})

	dbUser, err := repo.GetUserByLogin(ctx, login)
	require.NoError(t, err)

	user, err := repo.GetByID(ctx, dbUser.ID)
	require.NoError(t, err)
	require.Equal(t, dbUser.ID, user.ID)
	require.Equal(t, login, user.Login)
	require.Equal(t, "bidder", user.Role)
}

func TestGetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, -1)
	require.ErrorIs(t, err, model.ErrUserNotFound)
}


//go:build integration

package watchlist

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

func createTestData(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (userID, sellerID, lotID int64) {
	t.Helper()
	suffix := time.Now().UnixNano()
	err := pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`,
		fmt.Sprintf("user_%d", suffix), "x").Scan(&userID)
	require.NoError(t, err)

	err = pool.QueryRow(ctx, `INSERT INTO users (login, password_hash) VALUES ($1, $2)
		 RETURNING id`, fmt.Sprintf("seller_%d", suffix), "x",
	).Scan(&sellerID)
	require.NoError(t, err)

	err = pool.QueryRow(ctx, `INSERT INTO lots (title, description, category, start_price, current_price,
			status,	seller_id)
		VALUES ($1, 'Description', 'electronics', 100, 100, 'live', $2) RETURNING id`,
		"Watch Lot", sellerID).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM watchlist WHERE lot_id = $1`, lotID)
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id = $1`, lotID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1, $2)`, userID, sellerID)
	})
	return
}

func TestAdd_Integration(t *testing.T) {
	ctx := context.Background()

	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	userID, _, lotID := createTestData(t, ctx, pool)

	err = repo.Add(ctx, userID, lotID)
	require.NoError(t, err)

	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM watchlist WHERE user_id = $1 AND lot_id = $2`, userID, lotID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestAdd_Duplicate(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	userID, _, lotID := createTestData(t, ctx, pool)

	err = repo.Add(ctx, userID, lotID)
	require.NoError(t, err)

	err = repo.Add(ctx, userID, lotID)
	require.ErrorIs(t, err, model.ErrAlreadyWatching)
}

func TestDelete_Integration(t *testing.T) {
	ctx := context.Background()

	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	userID, _, lotID := createTestData(t, ctx, pool)

	err = repo.Add(ctx, userID, lotID)
	require.NoError(t, err)

	err = repo.Delete(ctx, userID, lotID)
	require.NoError(t, err)

	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM watchlist WHERE user_id = $1 AND lot_id = $2`, userID, lotID).Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestGetWatchlist_Integration(t *testing.T) {
	ctx := context.Background()

	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	userID, _, lotID := createTestData(t, ctx, pool)

	err = repo.Add(ctx, userID, lotID)
	require.NoError(t, err)

	items, total, err := repo.GetWatchlist(ctx, userID, 1, 10)
	require.NoError(t, err)

	require.Equal(t, 1, total)
	require.Len(t, items, 1)

	require.Equal(t, lotID, items[0].LotID)
	require.Equal(t, "Watch Lot", items[0].LotTitle)
	require.Equal(t, 100.0, items[0].CurrentPrice)
	require.Equal(t, "live", items[0].Status)
}

func TestGetWatchlist_Empty(t *testing.T) {
	ctx := context.Background()

	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	items, total, err := repo.GetWatchlist(ctx, -1, 1, 10)
	require.NoError(t, err)

	require.Zero(t, total)
	require.Empty(t, items)
}

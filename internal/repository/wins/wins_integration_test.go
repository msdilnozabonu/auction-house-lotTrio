//go:build integration
package wins

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func testPool(t *testing.T) *pgxpool.Pool {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err)
	return pool
}

func TestGetWins_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	suffix := time.Now().UnixNano()

	sellerLogin := fmt.Sprintf("seller_%d", suffix)
	bidderLogin := fmt.Sprintf("bidder_%d", suffix)

	var sellerID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO users(login, password_hash, role) VALUES ($1, 'hash', 'seller') RETURNING id`,
		sellerLogin).Scan(&sellerID)
	require.NoError(t, err)

	var bidderID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO users(login, password_hash, role) VALUES ($1, 'hash', 'bidder') RETURNING id`,
		bidderLogin).Scan(&bidderID)
	require.NoError(t, err)

	endsAt := time.Now().UTC().Truncate(time.Second)

	var lotID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO lots (seller_id, title,	description, category, start_price, current_price,
			current_winner_id, status, ends_at)
		VALUES ($1, 'MacBook', 'Laptop', 'electronics', 1000000, 1504000, $2, 'closed', $3)
		RETURNING id`, sellerID, bidderID, endsAt).Scan(&lotID)
	require.NoError(t, err)

	wins, err := repo.GetWins(ctx, bidderID)
	require.NoError(t, err)
	require.Len(t, wins, 1)
	require.Equal(t, lotID, wins[0].LotID)
	require.Equal(t, "MacBook", wins[0].LotTitle)
	require.Equal(t, 1504000.0, wins[0].WinningBid)
	require.WithinDuration(t, endsAt, wins[0].ClosedAt, time.Second)
}

func TestGetWins_Empty(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	wins, err := repo.GetWins(ctx, -1)
	require.NoError(t, err)
	require.Empty(t, wins)
}

func TestGetWins_QueryError(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)

	repo, err := New(pool)
	require.NoError(t, err)

	pool.Close()

	_, err = repo.GetWins(ctx, 1)
	require.Error(t, err)
	require.ErrorContains(t, err, "get wins")
}

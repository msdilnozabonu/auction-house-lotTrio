//go:build integration
package bid

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"auction-house-lotTrio/internal/model"

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

func TestPlaceBid_ConcurrentBids_ExactlyOneWins(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := New(pool)
	require.NoError(t, err)

	var seller, bidder1, bidder2, lotID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1,$2) RETURNING id`,
		"seller_"+t.Name(), "x").Scan(&seller)
	require.NoError(t, err)

	err = pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1,$2) RETURNING id`,
		"bidder1_"+t.Name(), "x").Scan(&bidder1)
	require.NoError(t, err)

	err = pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1,$2) RETURNING id`,
		"bidder2_"+t.Name(), "x").Scan(&bidder2)
	require.NoError(t, err)

	err = pool.QueryRow(ctx,
		`INSERT INTO lots (title, start_price, current_price, status, seller_id)
		 VALUES ($1,$2,$2,'live',$3) RETURNING id`,
		"race-lot-"+t.Name(), 100.0, seller).Scan(&lotID)
	require.NoError(t, err)
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM bids WHERE lot_id = $1`, lotID)
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id = $1`, lotID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, seller, bidder1, bidder2)
	}()

	bidders := []int64{bidder1, bidder2}
	errs := make([]error, 2)
	var (
		wg    sync.WaitGroup
		start = make(chan struct{})
	)
	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			errs[i] = repo.PlaceBid(ctx, lotID, bidders[i], 150.0)
		}(i)
	}
	close(start)
	wg.Wait()

	var successCount, outbidCount int
	for _, e := range errs {
		switch {
		case e == nil:
			successCount++
		case errors.Is(e, model.ErrOutbid):
			outbidCount++
		default:
			t.Fatalf("unexpected error: %v", e)
		}
	}
	require.Equal(t, 1, successCount, "exactly one concurrent bid must be accepted")
	require.Equal(t, 1, outbidCount, "the other concurrent bid must be rejected as outbid")

	var finalPrice float64
	var winnerID int64
	err = pool.QueryRow(ctx,
		`SELECT current_price, current_winner_id FROM lots WHERE id = $1`, lotID).
		Scan(&finalPrice, &winnerID)
	require.NoError(t, err)
	require.InDelta(t, 150.0, finalPrice, 0.001)
	require.Contains(t,
		[]int64{bidder1, bidder2},
		winnerID,
	)

	var bidCount int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bids WHERE lot_id = $1`,
		lotID,
	).Scan(&bidCount)

	require.NoError(t, err)
	require.Equal(t, 1, bidCount)
}
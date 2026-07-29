//go:build integration
package bid

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
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

func uniqueSuffix(t *testing.T) string {
	t.Helper()
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000_000))
	require.NoError(t, err)
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), n.Int64())
}

func createUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, prefix string) int64 {
	t.Helper()
	login := fmt.Sprintf("%s_%s", prefix, uniqueSuffix(t))

	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO users (login, password_hash) VALUES ($1,$2) RETURNING id`,
		login, "x",
	).Scan(&id)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

func TestPlaceBid_ConcurrentBids_ExactlyOneWins(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo, err := New(pool)
	require.NoError(t, err)

	seller := createUser(t, ctx, pool, "seller")
	bidder1 := createUser(t, ctx, pool, "bidder1")
	bidder2 := createUser(t, ctx, pool, "bidder2")

	var lotID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO lots (title, start_price, current_price, status, seller_id)
     VALUES ($1,$2,$2,'live',$3) RETURNING id`,
		fmt.Sprintf("race-lot_%s", uniqueSuffix(t)),
		100.0,
		seller,
	).Scan(&lotID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM bids WHERE lot_id = $1`, lotID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM lots WHERE id = $1`, lotID)
	})

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
	require.Contains(t, []int64{bidder1, bidder2}, winnerID)

	var bidCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM bids WHERE lot_id = $1`, lotID).Scan(&bidCount)
	require.NoError(t, err)
	require.Equal(t, 1, bidCount)
}

func TestGetBidderBids_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createUser(t, ctx, pool, "seller")
	bidderID := createUser(t, ctx, pool, "bidder")

	var lotID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO lots (title, start_price, current_price, status, seller_id)
		VALUES ($1,$2,$2,'closed',$3) RETURNING id`, "Test Lot", 100.0, sellerID).Scan(&lotID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM bids WHERE lot_id=$1`, lotID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM lots WHERE id=$1`, lotID)
	})

	_, err = pool.Exec(ctx,
		`INSERT INTO bids(lot_id, bidder_id, amount) VALUES ($1,$2,$3)`, lotID, bidderID, 150.0)
	require.NoError(t, err)

	bids, err := repo.GetBidderBids(ctx, bidderID)
	require.NoError(t, err)
	require.Len(t, bids, 1)

	require.Equal(t, lotID, bids[0].LotID)
	require.Equal(t, "Test Lot", bids[0].LotTitle)
	require.Equal(t, 150.0, bids[0].Amount)
}

func TestGetBidderBids_Empty(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	bids, err := repo.GetBidderBids(ctx, -1)
	require.NoError(t, err)
	require.Empty(t, bids)
}

func TestGetBidsByLotIDForSeller_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createUser(t, ctx, pool, "seller")
	bidderID := createUser(t, ctx, pool, "bidder")

	var lotID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO lots (title, start_price, current_price, status, seller_id)
		VALUES ($1,$2,$2,'live',$3) RETURNING id`,"Seller Lot", 100.0, sellerID).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM bids WHERE lot_id=$1`, lotID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM lots WHERE id=$1`, lotID)
	})

	_, err = pool.Exec(ctx,
		`INSERT INTO bids (lot_id, bidder_id, amount)
			VALUES ($1,$2,$3), ($1,$2,$4)`, lotID, bidderID, 150.0, 200.0)
	require.NoError(t, err)

	bids, err := repo.GetBidsByLotIDForSeller(ctx, lotID)
	require.NoError(t, err)
	require.Len(t, bids, 2)
	require.Equal(t, lotID, bids[0].LotID)
	require.Equal(t, "Seller Lot", bids[0].LotTitle)
	require.Equal(t, bidderID, bids[0].BidderID)
}

func TestGetBidsByLotIDForBidder_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createUser(t, ctx, pool, "seller")
	bidderID := createUser(t, ctx, pool, "bidder")

	var lotID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO lots (title, start_price, current_price, status, seller_id)
		VALUES ($1,$2,$2,'live',$3) RETURNING id`, "Bidder Lot", 100.0, sellerID).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM bids WHERE lot_id=$1`, lotID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM lots WHERE id=$1`, lotID)
	})

	_, err = pool.Exec(ctx,
		`INSERT INTO bids (lot_id, bidder_id, amount)
		VALUES ($1,$2,100), ($1,$2,110), ($1,$2,120)`, lotID, bidderID)
	require.NoError(t, err)

	filter := model.BidFilter{
		LotID:  lotID,
		Limit:  2,
		Offset: 0,
	}

	bids, err := repo.GetBidsByLotIDForBidder(ctx, filter)
	require.NoError(t, err)
	require.Len(t, bids, 2)
	require.Equal(t, lotID, bids[0].LotID)
	require.Equal(t, "Bidder Lot", bids[0].LotTitle)
	require.Equal(t, bidderID, bids[0].BidderID)
	require.Equal(t, 120.0, bids[0].Amount)
	require.Equal(t, 110.0, bids[1].Amount)

	filter.Offset = 2
	bids, err = repo.GetBidsByLotIDForBidder(ctx, filter)
	require.NoError(t, err)
	require.Len(t, bids, 1)
	require.Equal(t, 100.0, bids[0].Amount)
}
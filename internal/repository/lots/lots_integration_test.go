//go:build integration

package lots

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

func createSeller(t *testing.T, ctx context.Context, pool *pgxpool.Pool) int64 {
	t.Helper()
	login := fmt.Sprintf("seller_%d", time.Now().UnixNano())

	var sellerID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO users(login, password_hash, role) VALUES ($1,'hash','seller') RETURNING id`,
		login).Scan(&sellerID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, sellerID)
	})
	return sellerID
}

func TestFindExpiredLot_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("expired_%d", time.Now().UnixNano())
	endsAt := time.Now().Add(-time.Hour)
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "", endsAt,
		"live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx,
		`SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})

	ids, err := repo.FindExpiredLot(ctx, time.Now())
	require.NoError(t, err)
	require.Contains(t, ids, lotID)
}

func TestCloseLot_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)

	var bidderID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO users(login,password_hash,role) VALUES($1,'hash','bidder') RETURNING id`,
		fmt.Sprintf("bidder_%d", time.Now().UnixNano())).Scan(&bidderID)
	require.NoError(t, err)

	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(-time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `INSERT INTO bids(lot_id,bidder_id,amount) VALUES($1,$2,$3)`, lotID, bidderID, 150)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM bids WHERE lot_id=$1`, lotID)
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, bidderID)
	})

	ok, err := repo.CloseLot(ctx, lotID)
	require.NoError(t, err)
	require.True(t, ok)

	lot, err := repo.GetById(ctx, lotID)
	require.NoError(t, err)
	require.Equal(t, "closed", lot.Status)
	require.Equal(t, bidderID, lot.WinnerID)
}

func TestCreateLot_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(
		ctx, title, "Description", "electronics", 100, "", time.Now().Add(time.Hour),
		"live", sellerID, 100)
	require.NoError(t, err)

	var count int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM lots WHERE title=$1`, title).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx,
			`DELETE FROM lots WHERE title=$1`,
			title,
		)
	})
}

func TestGetById_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	_, err = repo.GetById(ctx, -1)
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestGetByIdForBid_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})

	lot, err := repo.GetByIdForBid(ctx, lotID)
	require.NoError(t, err)
	require.Equal(t, lotID, lot.ID)
	require.Equal(t, title, lot.Title)
}

func TestGetAll_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})

	filter := model.LotsFilter{
		Search: title,
		Page:   1,
		Limit:  10,
	}

	lots, total, err := repo.GetAll(ctx, filter)
	require.NoError(t, err)
	require.NotZero(t, total)
	require.NotEmpty(t, lots)

	found := false
	for _, lot := range lots {
		if lot.ID == lotID {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestFindLotsAdmin_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("admin_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx,
		`SELECT id FROM lots WHERE title=$1`,
		title,
	).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})

	filter := model.LotsFilter{
		Status:   "live",
		SellerID: sellerID,
		Search:   title,
		Page:     1,
		Limit:    10,
	}

	lots, total, err := repo.FindLotsAdmin(ctx, filter)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, lots, 1)
	require.Equal(t, lotID, lots[0].ID)
	require.Equal(t, sellerID, lots[0].SellerID)
	require.Equal(t, title, lots[0].Title)
	require.Equal(t, "live", lots[0].Status)
}

func TestFindLotsAdmin_Empty(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	lots, total, err := repo.FindLotsAdmin(ctx, model.LotsFilter{
		Search: "this_lot_should_never_exist_123456789",
		Page:   1,
		Limit:  10,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, 0)
	require.Zero(t, total)
	require.Empty(t, lots)
}

func TestUploadPhoto_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})

	err = repo.UploadPhoto(ctx, lotID, "photo.jpg")
	require.NoError(t, err)

	photo, err := repo.GetPhoto(ctx, lotID)
	require.NoError(t, err)
	require.Equal(t, "photo.jpg", photo)
}

func TestGetPhoto_NotFound(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	_, err = repo.GetPhoto(ctx, -1)
	require.Error(t, err)
}

func TestUpdateLot_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})

	lot, err := repo.GetById(ctx, lotID)
	require.NoError(t, err)

	lot.Title = "Updated"
	lot.Description = "Updated description"
	lot.Category = "cars"
	lot.StartPrice = 500

	err = repo.UpdateLot(ctx, *lot)
	require.NoError(t, err)

	updated, err := repo.GetById(ctx, lotID)
	require.NoError(t, err)
	require.Equal(t, "Updated", updated.Title)
	require.Equal(t, "Updated description", updated.Description)
	require.Equal(t, "cars", updated.Category)
	require.Equal(t, 500.0, updated.StartPrice)
}

func TestUpdateStatus_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "closed", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title = $1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id = $1`, lotID)
	})

	ok, err := repo.UpdateStatus(ctx, lotID, "closed")
	require.NoError(t, err)
	require.True(t, ok)

	lot, err := repo.GetById(ctx, lotID)
	require.NoError(t, err)
	require.Equal(t, "closed", lot.Status)
}

func TestDeleteLots_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title = $1`, title).Scan(&lotID)
	require.NoError(t, err)

	err = repo.DeleteLots(ctx, lotID)
	require.NoError(t, err)

	_, err = repo.GetById(ctx, lotID)
	require.ErrorIs(t, err, model.ErrNotFound)
}

func TestModerateLot_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("moderate_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title = $1`, title).Scan(&lotID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `UPDATE lots SET moderation_status = 'pending' WHERE id = $1`, lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id = $1`, lotID)
	})

	ok, err := repo.ModerateLot(ctx, lotID, "approved", "")
	require.NoError(t, err)
	require.True(t, ok)

	var moderationStatus string
	err = pool.QueryRow(ctx,
		`SELECT moderation_status
		 FROM lots
		 WHERE id = $1`,
		lotID,
	).Scan(&moderationStatus)
	require.NoError(t, err)
	require.Equal(t, "approved", moderationStatus)
}

func TestGetPlatformStats_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	stats, err := repo.GetPlatformStats(ctx)
	require.NoError(t, err)
	require.NotNil(t, stats.LotsByStatus)
	require.NotNil(t, stats.TopCategories)
	require.GreaterOrEqual(t, stats.TotalRevenue, 0.0)
	require.GreaterOrEqual(t, stats.AverageCheck, 0.0)
}

func TestGetMineLots_Integration(t *testing.T) {
	ctx := context.Background()

	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "Description", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title = $1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id = $1`, lotID)
	})

	filter := model.LotsFilter{
		Page:  1,
		Limit: 10,
	}

	lots, total, err := repo.GetMineLots(ctx, sellerID, filter)
	require.NoError(t, err)
	require.Zero(t, total)
	require.Len(t, lots, 1)
	require.Equal(t, lotID, lots[0].ID)
	require.Equal(t, sellerID, lots[0].SellerID)
	require.Equal(t, title, lots[0].Title)
	require.Equal(t, "Description", lots[0].Description)
	require.Equal(t, "electronics", lots[0].Category)
	require.Equal(t, "live", lots[0].Status)
}
func TestCancelLot_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)
	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "desc", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM lots WHERE id=$1`, lotID)
	})
	ok, err := repo.CancelLot(ctx, lotID, "duplicate listing")
	require.NoError(t, err)
	require.True(t, ok)

	var status, reason string
	err = pool.QueryRow(ctx, `SELECT status, cancellation_reason FROM lots WHERE id=$1`, lotID).
		Scan(&status, &reason)
	require.NoError(t, err)
	require.Equal(t, "cancelled", status)
	require.Equal(t, "duplicate listing", reason)
}

func TestCreateReportAndGetListOfReports_Integration(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	defer pool.Close()

	repo, err := New(pool)
	require.NoError(t, err)

	sellerID := createSeller(t, ctx, pool)
	title := fmt.Sprintf("lot_%d", time.Now().UnixNano())
	err = repo.CreateLot(ctx, title, "desc", "electronics", 100, "",
		time.Now().Add(time.Hour), "live", sellerID, 100)
	require.NoError(t, err)

	var lotID int64
	err = pool.QueryRow(ctx, `SELECT id FROM lots WHERE title=$1`, title).Scan(&lotID)
	require.NoError(t, err)

	reporterID := createSeller(t, ctx, pool)
	err = repo.CreateReport(ctx, lotID, reporterID, "fraudulent activity")
	require.NoError(t, err)

	reports, err := repo.GetListOfReports(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, reports)

	found := false
	for _, rep := range reports {
		if rep.LotID == lotID && rep.ReporterID == reporterID {
			found = true
			require.Equal(t, "fraudulent activity", rep.Reason)
		}
	}
	require.True(t, found, "expected report not found")
}

package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	lotExpired = `SELECT id FROM lots WHERE status = 'live' AND ends_at < $1`
	closeLot   = `UPDATE lots 
	SET status = 'closed',
	    current_winner_id = (SELECT bidder_id FROM bids WHERE lot_id = $1
		ORDER BY amount DESC, created_at ASC LIMIT 1)
	WHERE id = $1 AND status = 'live'`
	selectByID = `SELECT id, title, description, category, start_price, current_price, COALESCE(current_winner_id, 0), 
       status, starts_at, ends_at, photo_path, seller_id FROM lots WHERE id = $1`
	selectByIDForBid = `SELECT id, title, description, category, start_price, current_price, 
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, photo_path, seller_id FROM lots WHERE id = $1 
       AND status = 'live'`
	selectAll = `SELECT id, seller_id, title, coalesce(description,''), category, start_price, current_price, 
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, coalesce(photo_path, '') 
		FROM lots WHERE status = 'live' 
      	AND ($1 = '' OR title ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
      	AND ($2 = '' OR category = $2) AND ($3 = 0 OR current_price >= $3)
      	AND ($4 = 0 OR current_price <= $4) ORDER BY id LIMIT $5 OFFSET $6`
	countAll = `SELECT COUNT(*) FROM lots WHERE status = 'live' 
        AND ($1 = '' OR title ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
        AND ($2 = '' OR category = $2) AND ($3 = 0 OR current_price >= $3) AND ($4 = 0 OR current_price <= $4)`
	baseQuery = `SELECT id, seller_id, title, coalesce(description, ''), start_price, current_price, 
       coalesce(current_winner_id, 0), status, starts_at, ends_at, coalesce(photo_path, '') FROM lots WHERE 1=1`
	baseCountQuery = `SELECT COUNT(*) FROM lots WHERE 1=1`
	moderateLot    = `UPDATE lots SET moderation_status = $1, rejection_reason = $2 
            WHERE id = $3 AND moderation_status = 'pending'`
	lotsByStatus = `SELECT status, COUNT(*) FROM lots GROUP BY status`
	revenue      = `SELECT COALESCE(SUM(current_price), 0), COALESCE(AVG(current_price), 0)
		FROM lots WHERE status = 'closed' AND current_winner_id IS NOT NULL`
	topCategories = `SELECT category, COUNT(*), COALESCE(SUM(current_price), 0) FROM lots WHERE status = 'closed'
		AND current_winner_id IS NOT NULL GROUP BY category ORDER BY SUM(current_price) DESC LIMIT 5`
	lotExport = `SELECT id, title, category, status, seller_id, current_price, ends_at 
	FROM lots WHERE %s BETWEEN $1 AND $2 ORDER BY id`
	selectMine = `Select id, seller_id, title, description,category, start_price, current_price,
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, photo_path FROM lots WHERE seller_id = $1 
      	AND ($2 = '' OR title ILIKE '%'||$2||'%' OR description ILIKE '%'||$2||'%')
      	AND ($3 = '' OR category = $3) AND ($4 = 0 OR current_price >= $4)
      	AND ($5 = 0 OR current_price <= $5) ORDER BY id LIMIT $6 OFFSET $7`
	cancelLot = `UPDATE lots SET status = 'cancelled', cancellation_reason = $1
            WHERE id = $2 AND status IN('draft', 'live')`
	createReport= `INSERT INTO reports (lot_id, reporter_id, reason)
				VALUES ($1, $2, $3)`
	getListOfReports = `SELECT id, lot_id, reporter_id, reason, created_at 
		FROM reports WHERE status = 'pending' ORDER BY created_at`
)

type Repo interface {
	FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error)
	CloseLot(ctx context.Context, id int64) (bool, error)
	CreateLot(ctx context.Context, title, description, category string, startPrice float64, photo string,
		endsAt time.Time, status string, sellerID int64, currentPrice float64) error
	GetAll(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error)
	GetById(ctx context.Context, id int64) (*model.Lots, error)
	GetByIdForBid(ctx context.Context, id int64) (*model.Lots, error)
	UpdateLot(ctx context.Context, l model.Lots) error
	UpdateStatus(ctx context.Context, id int64, status string) (bool, error)
	DeleteLots(ctx context.Context, id int64) error
	FindLotsAdmin(ctx context.Context, lots model.LotsFilter) ([]model.Lots, int, error)
	UploadPhoto(ctx context.Context, id int64, photo string) error
	GetPhoto(ctx context.Context, id int64) (string, error)
	ModerateLot(ctx context.Context, id int64, status string, reason string) (bool, error)
	GetPlatformStats(ctx context.Context) (model.PlatformStats, error)
	ExportLots(ctx context.Context, dateField string, dateFrom, dateTo time.Time) (
		[]model.LotsExport, error)
	GetMineLots(ctx context.Context, sellerID int64, filter model.LotsFilter) ([]model.Lots, int, error)
	CancelLot(ctx context.Context, id int64, reason string) (bool, error)
	CreateReport(ctx context.Context, lotId, reporterId int64, reason string) error
	GetListOfReports(ctx context.Context) ([]model.ReportLot, error)
}

type repo struct {
	repo *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (Repo, error) {
	return &repo{repo: pool}, nil
}

func (r *repo) FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error) {
	rows, err := r.repo.Query(ctx, lotExpired, now)
	if err != nil {
		return nil, fmt.Errorf("find expired lot: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan lot id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get lot ids: %w", err)
	}
	return ids, nil
}

func (r *repo) CloseLot(ctx context.Context, id int64) (bool, error) {
	tx, err := r.repo.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin transaction: %w", err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}(tx, ctx)
	tags, err := tx.Exec(ctx, closeLot, id)
	if err != nil {
		return false, fmt.Errorf("close lot: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit transaction: %w", err)
	}
	return tags.RowsAffected() > 0, nil
}
func (r *repo) CreateLot(ctx context.Context, title, description, category string, startPrice float64,
	photo string, endsAt time.Time, status string, sellerID int64, currentPrice float64) error {
	_, err := r.repo.Exec(ctx, `INSERT INTO lots (title, description, category, start_price, photo_path, 
                  ends_at, status, seller_id, current_price) 
     VALUES ($1,$2, $3,$4, $5, $6, $7, $8, $9)`,
		title, description, category, startPrice, photo, endsAt, status, sellerID, currentPrice)
	if err != nil {
		return model.ErrDatabase
	}
	return nil
}

func (r *repo) GetAll(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 15
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int
	err := r.repo.QueryRow(ctx, countAll, filter.Search, filter.Category, filter.MinPrice, filter.MaxPrice).
		Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count lots: %w", err)
	}

	rows, err := r.repo.Query(ctx, selectAll, filter.Search, filter.Category, filter.MinPrice, filter.MaxPrice,
		limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get lots: %w", err)
	}
	defer rows.Close()

	var out []model.Lots
	for rows.Next() {
		var p model.Lots
		if err = rows.Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &p.Category, &p.StartPrice, &p.CurrentPrice,
			&p.WinnerID, &p.Status, &p.StartAt, &p.EndAt, &p.Photo); err != nil {
			return nil, 0, fmt.Errorf("get lots: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("get lots: %w", err)
	}
	return out, total, nil
}

func (r *repo) GetById(ctx context.Context, id int64) (*model.Lots, error) {
	return r.getByID(ctx, selectByID, id)
}

func (r *repo) GetByIdForBid(ctx context.Context, id int64) (*model.Lots, error) {
	return r.getByID(ctx, selectByIDForBid, id)
}

func (r *repo) UpdateLot(ctx context.Context, l model.Lots) error {
	err := r.repo.QueryRow(ctx, `UPDATE lots SET title = $1, description = $2, category = $3, start_price = $4, 
                photo_path = $5, ends_at = $6 WHERE id = $7 RETURNING id, seller_id, title, 
                description, category, start_price, current_price, status, photo_path, ends_at`,
		l.Title, l.Description, l.Category, l.StartPrice, l.Photo, l.EndAt, l.ID).
		Scan(&l.ID, &l.SellerID, &l.Title, &l.Description, &l.Category, &l.StartPrice,
			&l.CurrentPrice, &l.Status, &l.Photo, &l.EndAt)
	if err != nil {
		slog.Error("update lots by id", "err", err)
		return fmt.Errorf("update lots: %w", err)
	}
	return nil
}

func (r *repo) UpdateStatus(ctx context.Context, id int64, status string) (bool, error) {
	var returnedID int64
	err := r.repo.QueryRow(ctx, `UPDATE lots
    SET status = $1
    WHERE id = $2 AND ($1 != 'live' OR moderation_status = 'approved')
    RETURNING id`, status, id).
		Scan(&returnedID)
	if err != nil {
		slog.Error("update lots by id", "err", err)
		return false, fmt.Errorf("update lots: %w", err)
	}
	return true, nil
}

func (r *repo) DeleteLots(ctx context.Context, id int64) error {
	_, err := r.repo.Exec(ctx, `DELETE FROM lots WHERE id = $1`, id)
	if err != nil {
		slog.Error("delete lots by id", "err", err)
		return fmt.Errorf("delete lots: %w", err)
	}
	return nil
}

//nolint:funlen
func (r *repo) FindLotsAdmin(ctx context.Context, lots model.LotsFilter) ([]model.Lots, int, error) { //nolint:cyclop
	baseQ := baseQuery
	baseCountQ := baseCountQuery
	var args []any
	argN := 1
	if lots.Status != "" {
		baseQ += fmt.Sprintf(" AND status = $%d", argN)
		baseCountQ += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, lots.Status)
		argN++
	}
	if lots.SellerID != 0 {
		baseQ += fmt.Sprintf(" AND seller_id = $%d", argN)
		baseCountQ += fmt.Sprintf(" AND seller_id = $%d", argN)
		args = append(args, lots.SellerID)
		argN++
	}
	if lots.Search != "" {
		baseQ += fmt.Sprintf(" AND (title ILIKE $%d)", argN)
		baseCountQ += fmt.Sprintf(" AND (title ILIKE $%d)", argN)
		args = append(args, "%"+lots.Search+"%")
		argN++
	}
	if lots.DateFrom != nil {
		baseQ += fmt.Sprintf(" AND starts_at >= $%d", argN)
		baseCountQ += fmt.Sprintf(" AND starts_at >= $%d", argN)
		args = append(args, *lots.DateFrom)
		argN++
	}
	if lots.DateTo != nil {
		baseQ += fmt.Sprintf(" AND ends_at <= $%d", argN)
		baseCountQ += fmt.Sprintf(" AND ends_at <= $%d", argN)
		args = append(args, *lots.DateTo)
		argN++
	}

	var total int
	err := r.repo.QueryRow(ctx, baseCountQ, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count lots: %w", err)
	}

	baseQ += fmt.Sprintf(" ORDER BY id ASC LIMIT $%d OFFSET $%d", argN, argN+1)
	args = append(args, lots.Limit, (lots.Page-1)*lots.Limit)

	rows, err := r.repo.Query(ctx, baseQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("find lots: %w", err)
	}
	defer rows.Close()

	var result []model.Lots
	for rows.Next() {
		var l model.Lots
		if err = rows.Scan(&l.ID, &l.SellerID, &l.Title, &l.Description, &l.StartPrice, &l.CurrentPrice,
			&l.WinnerID, &l.Status, &l.StartAt, &l.EndAt, &l.Photo); err != nil {
			return nil, 0, fmt.Errorf("find lots: %w", err)
		}
		result = append(result, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("find lots: %w", err)
	}
	return result, total, nil
}

func (r *repo) getByID(ctx context.Context, sqlQuery string, id int64) (*model.Lots, error) {
	var lot model.Lots
	err := r.repo.QueryRow(ctx, sqlQuery, id).Scan(&lot.ID, &lot.Title, &lot.Description, &lot.Category,
		&lot.StartPrice, &lot.CurrentPrice, &lot.WinnerID, &lot.Status, &lot.StartAt, &lot.EndAt,
		&lot.Photo, &lot.SellerID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrNotFound
	}
	if err != nil {
		slog.Error("get lot by id", "err", err)
		return nil, fmt.Errorf("get lot: %w", err)
	}

	return &lot, nil
}

func (r *repo) UploadPhoto(ctx context.Context, id int64, photo string) error {
	_, err := r.repo.Exec(ctx, `UPDATE lots SET photo_path = $1 WHERE id = $2`, photo, id)
	if err != nil {
		slog.Error("upload photo by id", "err", err)
		return fmt.Errorf("upload photo by id: %w", err)
	}
	return nil
}

func (r *repo) GetPhoto(ctx context.Context, id int64) (string, error) {
	var photo string
	err := r.repo.QueryRow(ctx, `SELECT photo_path FROM lots WHERE id = $1`, id).
		Scan(&photo)
	if err != nil {
		slog.Error("get photo by id", "err", err)
		return "", fmt.Errorf("get photo by id: %w", err)
	}
	return photo, nil
}

func (r *repo) ModerateLot(ctx context.Context, id int64, status string, reason string) (bool, error) {
	updated, err := r.repo.Exec(ctx, moderateLot, status, reason, id)
	if err != nil {
		return false, fmt.Errorf("moderate lot: %w", err)
	}
	return updated.RowsAffected() > 0, nil
}

func (r *repo) GetPlatformStats(ctx context.Context) (model.PlatformStats, error) {
	var stats model.PlatformStats
	stats.LotsByStatus = make(map[string]int)

	rows, err := r.repo.Query(ctx, lotsByStatus)
	if err != nil {
		return stats, fmt.Errorf("get lots by status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		err = rows.Scan(&status, &count)
		if err != nil {
			return stats, fmt.Errorf("scan status count: %w", err)
		}
		stats.LotsByStatus[status] = count
	}
	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("get lots by status: %w", err)
	}
	err = r.repo.QueryRow(ctx, revenue).Scan(&stats.TotalRevenue, &stats.AverageCheck)
	if err != nil {
		return stats, fmt.Errorf("get lots by revenue: %w", err)
	}
	topCatRows, err := r.repo.Query(ctx, topCategories)
	if err != nil {
		return stats, fmt.Errorf("get top categories: %w", err)
	}
	defer topCatRows.Close()
	for topCatRows.Next() {
		var catStats model.CategoryStat
		err = topCatRows.Scan(&catStats.Category, &catStats.Count, &catStats.Revenue)
		if err != nil {
			return stats, fmt.Errorf("scan category stats: %w", err)
		}
		stats.TopCategories = append(stats.TopCategories, catStats)
	}
	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("get top categories: %w", err)
	}
	return stats, nil
}

func (r *repo) ExportLots(ctx context.Context, dateField string, dateFrom, dateTo time.Time) (
	[]model.LotsExport, error) {
	query := fmt.Sprintf(lotExport, dateField)
	rows, err := r.repo.Query(ctx, query, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("export lots: %w", err)
	}
	defer rows.Close()
	var result []model.LotsExport
	for rows.Next() {
		var l model.LotsExport
		if err := rows.Scan(&l.ID, &l.Title, &l.Category, &l.Status, &l.SellerID, &l.Price, &l.EndsAt); err != nil {
			return nil, fmt.Errorf("scan lots export: %w", err)
		}
		result = append(result, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("export lots: %w", err)
	}
	return result, nil
}

func (r *repo) GetMineLots(ctx context.Context, sellerID int64, filter model.LotsFilter) ([]model.Lots, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 15
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	rows, err := r.repo.Query(ctx, selectMine, sellerID, filter.Search, filter.Category,
		filter.MinPrice, filter.MaxPrice, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get lots: %w", err)
	}
	defer rows.Close()

	var out []model.Lots
	for rows.Next() {
		var p model.Lots
		if err := rows.Scan(&p.ID, &p.SellerID, &p.Title, &p.Description, &p.Category,
			&p.StartPrice, &p.CurrentPrice, &p.WinnerID, &p.Status, &p.StartAt, &p.EndAt, &p.Photo); err != nil {
			return nil, 0, fmt.Errorf("get mine lots: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("get mine lots: %w", err)
	}
	return out, 0, nil
}

func (r *repo) CancelLot(ctx context.Context, id int64, reason string) (bool, error) {
	row, err := r.repo.Exec(ctx, cancelLot, reason, id)
	if err != nil {
		return false, fmt.Errorf("cancel lot: %w", err)
	}
	affected := row.RowsAffected()
	if affected == 0 {
		return false, model.ErrStatusNotChanged
	}
	return true, nil
}

func (r *repo) CreateReport(ctx context.Context, lotId, reporterId int64, reason string) error{
	_, err := r.repo.Exec(ctx, createReport, lotId, reporterId, reason)
		if err != nil {
			return fmt.Errorf("create report: %w", err)
		}
		return nil
}

func (r *repo) GetListOfReports(ctx context.Context) ([]model.ReportLot, error){
	rows, err := r.repo.Query(ctx, getListOfReports)
	if err != nil {
		return nil, fmt.Errorf("get list of reports: %w", err)
	}
	defer rows.Close()
	var result []model.ReportLot
	for rows.Next(){
		var rep model.ReportLot
		if err := rows.Scan(&rep.ID, &rep.LotID,
			&rep.ReporterID, &rep.Reason, &rep.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan report lot: %w", err)
		}
		result = append(result, rep)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get list of reports: %w", err)
	}
	return result, nil
}
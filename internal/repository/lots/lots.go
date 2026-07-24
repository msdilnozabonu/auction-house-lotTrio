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
	selectAll = `SELECT id, seller_id, title, description, category, start_price, current_price, 
       COALESCE(current_winner_id, 0), status, starts_at, ends_at, photo_path FROM lots WHERE status = 'live' 
      	AND ($1 = '' OR title ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
      	AND ($2 = '' OR category = $2) AND ($3 = 0 OR current_price >= $3)
      	AND ($4 = 0 OR current_price <= $4) ORDER BY id LIMIT $5 OFFSET $6`
	countAll = `SELECT COUNT(*) FROM lots WHERE status = 'live' 
        AND ($1 = '' OR title ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
        AND ($2 = '' OR category = $2) AND ($3 = 0 OR current_price >= $3) AND ($4 = 0 OR current_price <= $4)`
	baseQuery = `SELECT id, seller_id, title, description, start_price, current_price, current_winner_id, status, 
       starts_at, ends_at, photo_path FROM lots WHERE 1=1`
	baseCountQuery = `SELECT COUNT(*) FROM lots WHERE 1=1`
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
	UpdateStatus(ctx context.Context, id int64, status string) error
	DeleteLots(ctx context.Context, id int64) error
	FindLotsAdmin(ctx context.Context, lots model.LotsFilter) ([]model.Lots, int, error)
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

func (r *repo) UpdateStatus(ctx context.Context, id int64, status string) error {
	err := r.repo.QueryRow(ctx, `UPDATE lots SET status = $1 where id = $2 returning status, id`, status, id).
		Scan(&status, &id)
	if err != nil {
		slog.Error("update lots by id", "err", err)
		return fmt.Errorf("update lots: %w", err)
	}
	return nil
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
	args = append(args, lots.PageSize, (lots.Page-1)*lots.PageSize)

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
  
func (r *repo) getByID(ctx context.Context, sqlQuery  string, id int64) (*model.Lots, error) {
	var lot model.Lots
	err := r.repo.QueryRow(ctx, sqlQuery , id).Scan(&lot.ID, &lot.Title, &lot.Description, &lot.Category,
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

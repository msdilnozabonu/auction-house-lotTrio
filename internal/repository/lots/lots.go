package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	lotExpired = `SELECT id FROM lots WHERE status = 'live' AND ends_at < $1`
	closeLot   = `UPDATE lots SET status = 'closed' WHERE id = $1 AND status = 'live'`
	selectByID = `SELECT id, title, description, start_price, current_price, current_winner_id, 
       status, starts_at, ends_at, photo_path FROM lots WHERE id = $1`
	selectAll  = `SELECT id, title, description, start_price, current_price, status, starts_at, 
       ends_at, photo_path FROM lots WHERE status = 'live' ORDER BY id`
)

type Repo interface {
	FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error)
	CloseLot(ctx context.Context, id int64) (bool, error)
	CreateLot(ctx context.Context, title, description string, startPrice float64, photo string,
		endsAt time.Time, status string, sellerID int64, currentPrice float64) error
	GetAll(ctx context.Context) ([]model.Lots, error)
	UpdateLot(ctx context.Context, id int64, status string, currentPrice int64) error
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
	tags, err := r.repo.Exec(ctx, closeLot, id)
	if err != nil {
		return false, fmt.Errorf("close lot: %w", err)
	}
	return tags.RowsAffected() > 0, nil
}
func (r *repo) CreateLot(ctx context.Context, title, description string, startPrice float64,
	photo string, endsAt time.Time, status string, sellerID int64, currentPrice float64) error {
	_, err := r.repo.Exec(ctx, `INSERT INTO lots (title, description, start_price, photo_path, 
                  ends_at, status, seller_id, current_price) 
     VALUES ($1,$2, $3,$4, $5, $6, $7, $8)`,
		title, description, startPrice, photo, endsAt, status, sellerID, currentPrice)
	if err != nil {
		return model.ErrDatabase
	}
	return nil
}

func (r *repo) GetAll(ctx context.Context) ([]model.Lots, error) {
	rows, err := r.repo.Query(ctx, selectAll)
	if err != nil {
		return nil, fmt.Errorf("get lots: %w", err)
	}
	defer rows.Close()

	var out []model.Lots
	for rows.Next() {
		var p model.Lots
		if err = rows.Scan(&p.ID, &p.Title, &p.Description, &p.StartPrice, &p.CurrentPrice,
			&p.Status, &p.StartAt, &p.EndAt, &p.Photo); err != nil {
			return nil, fmt.Errorf("get lots: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get lots: %w", err)
	}
	return out, nil
}

// UpdateLot в процессе.
func (r *repo) UpdateLot(ctx context.Context, id int64, status string, currentPrice int64) error {
	err := r.repo.QueryRow(ctx, `UPDATE lots SET status = $1, current_price = $2 WHERE id = $3`,
		status, currentPrice, id).Scan(&id, &status, &currentPrice)
	if err != nil {
		return fmt.Errorf("%w: %v", model.ErrDatabase, err)
	}
	return nil
}

func (r *repo) DeleteLots(ctx context.Context, id int64) {}

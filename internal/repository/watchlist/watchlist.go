package watchlist

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	addWatch       = `INSERT INTO watchlist (user_id, lot_id) VALUES ($1, $2)`
	deleteWatch    = `DELETE FROM watchlist WHERE user_id = $1 AND lot_id = $2`
	countWatchlist = `SELECT COUNT(*) FROM watchlist WHERE user_id = $1`

	getWatchlist = `SELECT l.id, l.title, l.current_price, l.status, w.created_at FROM watchlist w
    JOIN lots l ON l.id = w.lot_id WHERE w.user_id = $1 ORDER BY w.created_at DESC LIMIT $2 OFFSET $3
`
)

type Repo interface {
	Add(ctx context.Context, userID, lotID int64) error
	Delete(ctx context.Context, userID, lotID int64) error
	GetWatchlist(ctx context.Context, userID int64, page, limit int) ([]model.WatchItem, int, error)
}

type repo struct {
	repo *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (Repo, error) {
	return &repo{
		repo: pool,
	}, nil
}

func (r *repo) Add(ctx context.Context, userID, lotID int64) error {
	_, err := r.repo.Exec(ctx, addWatch, userID, lotID)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return model.ErrAlreadyWatching
		}

		return fmt.Errorf("add watch: %w", err)
	}

	return nil
}

func (r *repo) Delete(ctx context.Context, userID, lotID int64) error {
	_, err := r.repo.Exec(ctx, deleteWatch, userID, lotID)
	if err != nil {
		return fmt.Errorf("delete watch: %w", err)
	}

	return nil
}

func (r *repo) GetWatchlist(ctx context.Context, userID int64, page, limit int) ([]model.WatchItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 15
	}

	offset := (page - 1) * limit

	var total int
	err := r.repo.QueryRow(ctx, countWatchlist, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count watchlist: %w", err)
	}
	rows, err := r.repo.Query(ctx, getWatchlist, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get watchlist: %w", err)
	}
	defer rows.Close()
	items := make([]model.WatchItem, 0)

	for rows.Next() {
		var item model.WatchItem

		if err := rows.Scan(
			&item.LotID,
			&item.LotTitle,
			&item.CurrentPrice,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan watchlist: %w", err)
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate watchlist: %w", err)
	}
	return items, total, nil
}

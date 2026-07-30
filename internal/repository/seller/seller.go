package seller

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	countStats = `SELECT COUNT(*), COALESCE(SUM(current_price), 0), COALESCE(AVG(current_price), 0) 
		FROM lots WHERE status = 'closed' AND seller_id = $1`
	sellerStats = `SELECT id, title, current_price FROM lots WHERE status = 'closed' AND seller_id = $1 
                    ORDER BY current_price DESC LIMIT 3`
)

type sellerRepo struct {
	pool *pgxpool.Pool


}

type Repo interface {
	GetSellerStats(ctx context.Context, sellerID int64) (model.SellerStats, error)
	CanceledLot(ctx context.Context, id, sellerID int64 ) error
}

func New(pool *pgxpool.Pool) (Repo, error) {
	return &sellerRepo{
		pool: pool,

	}, nil
}

func (r *sellerRepo) GetSellerStats(ctx context.Context, sellerID int64) (model.SellerStats, error) {
	var s model.SellerStats
	err := r.pool.QueryRow(ctx, countStats, sellerID).
		Scan(&s.CountSelle, &s.SumCurrentPrice, &s.AVGCurrentPrice)
	if err != nil {
		return s, fmt.Errorf("seller stats: %w", err)
	}

	rows, err := r.pool.Query(ctx, sellerStats, sellerID)
	if err != nil {
		return s, fmt.Errorf("top lots: %w", err)
	}
	defer rows.Close()

	var out []model.TopLot
	for rows.Next() {
		var p model.TopLot
		if err = rows.Scan(&p.ID, &p.Title, &p.CurrentPrice); err != nil {
			return s, fmt.Errorf("top lots: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return s, fmt.Errorf("top lots: %w", err)
	}

	s.TopLots = out
	return s, nil
}

func (r *sellerRepo) CanceledLot(ctx context.Context, id, sellerID int64 ) error {
	_, err := r.pool.Exec(ctx, `UPDATE lots SET status = 'cancelled' WHERE id = $1 AND seller_id = $2`, id, sellerID)
	if err != nil {
		slog.Error("update canceled lot", "err", err)
		return fmt.Errorf("cancel lot: %w", err)
	}
	return nil
}

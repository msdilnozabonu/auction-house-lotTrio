package lots

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	lotExpired = `SELECT id FROM lots WHERE status = 'live' AND ends_at < $1`
	closeLot   = `UPDATE lots SET status = 'closed' WHERE id = $1 AND status = 'live'`
)

type Repo interface {
	FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error)
	CloseLot(ctx context.Context, id int64) (bool, error)
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

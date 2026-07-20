package lots

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	LotExpired = `SELECT * FROM lots WHERE status = 'live' AND end_date < $1`
)

type Repo interface {
}

type repo struct {
	repo *pgxpool.Pool
}

func New(ctx context.Context, pool *pgxpool.Pool) (Repo, error) {
	return &repo{repo: pool}, nil
}

func (r *repo) FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error) {
	rows, err := r.repo.Query(ctx, LotExpired, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

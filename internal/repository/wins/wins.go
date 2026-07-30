package wins

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	getBidWins = `SELECT id, title, current_price, ends_at FROM lots WHERE current_winner_id = $1 
                AND status = 'closed' ORDER BY ends_at DESC`
)

type Repo interface {
	GetWins(ctx context.Context, bidderID int64) ([]model.Win, error)
}

type repo struct {
	repo *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (Repo, error) {
	return &repo{
		repo: pool,
	}, nil
}

func (r *repo) GetWins(ctx context.Context, bidderID int64) ([]model.Win, error) {
	rows, err := r.repo.Query(ctx, getBidWins, bidderID)
	if err != nil {
		return nil, fmt.Errorf("get wins: %w", err)
	}
	defer rows.Close()

	wins := make([]model.Win, 0)
	for rows.Next() {
		var win model.Win

		if err := rows.Scan(
			&win.LotID, &win.LotTitle, &win.WinningBid, &win.ClosedAt); err != nil {
			return nil, fmt.Errorf("scan win: %w", err)
		}
		wins = append(wins, win)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate wins: %w", err)
	}

	return wins, nil
}

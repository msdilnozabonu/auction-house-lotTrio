package bid

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	updateBid = `UPDATE lots
        SET current_price = $1,
            current_winner_id = $2
        WHERE id = $3
          AND status = 'live'
          AND $1 > current_price`

	insertBid = `INSERT INTO bids (lot_id, bidder_id, amount)
        VALUES ($1, $2, $3)`
)

type Repo interface {
	PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error
}

type repo struct {
	repo *pgxpool.Pool
}

func New(pool *pgxpool.Pool) (Repo, error) {
	return &repo{
		repo: pool,
	}, nil
}

func (r *repo) PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error {
	tx, err := r.repo.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	tag, err := tx.Exec(ctx, updateBid, amount, bidderID, lotID)
	if err != nil {
		return fmt.Errorf("update lot bid: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return model.ErrOutbid
	}

	_, err = tx.Exec(ctx, insertBid, lotID, bidderID, amount)
	if err != nil {
		return fmt.Errorf("insert bid: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

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
          AND status = 'live' AND ((current_winner_id IS NULL AND $1 >= current_price) OR
          (current_winner_id IS NOT NULL AND $1 > current_price))`

	insertBid = `INSERT INTO bids (lot_id, bidder_id, amount)
        VALUES ($1, $2, $3)`
	getBidderBids = `SELECT b.lot_id, l.title, b.amount, b.created_at FROM bids b
		JOIN lots l ON l.id = b.lot_id WHERE b.bidder_id = $1 ORDER BY b.created_at DESC`
	getBidsBylotIDForSeller = `Select b.lot_id, l.title, b.bidder_id, b.amount, b.created_at FROM bids b 
    join lots l on l.id = b.lot_id where b.lot_id = $1`
	getBidsByLotIDForBidder = `SELECT b.lot_id, l.title, b.bidder_id, b.amount, b.created_at FROM bids b 
        JOIN lots l ON l.id = b.lot_id WHERE b.lot_id = $1 ORDER BY b.created_at DESC LIMIT $2 OFFSET $3`
)

type Repo interface {
	PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error
	GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error)
	GetBidsByLotIDForSeller(ctx context.Context, lotID int64) ([]model.Bid, error)
	GetBidsByLotIDForBidder(ctx context.Context, filter model.BidFilter) ([]model.Bid, error)
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

func (r *repo) GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error) {
	rows, err := r.repo.Query(ctx, getBidderBids, bidderID)
	if err != nil {
		return nil, fmt.Errorf("get bidder bids: %w", err)
	}
	defer rows.Close()

	bids := make([]model.Bid, 0)
	for rows.Next() {
		var bid model.Bid
		if err := rows.Scan(&bid.LotID, &bid.LotTitle, &bid.Amount, &bid.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan bid: %w", err)
		}

		bids = append(bids, bid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bids: %w", err)
	}

	return bids, nil
}

func (r *repo) GetBidsByLotIDForSeller(ctx context.Context, lotID int64) ([]model.Bid, error) {
	rows, err := r.repo.Query(ctx, getBidsBylotIDForSeller, lotID)
	if err != nil {
		return nil, fmt.Errorf("get bids: %w", err)
	}
	defer rows.Close()

	bids := make([]model.Bid, 0)
	for rows.Next() {
		var bid model.Bid
		if err := rows.Scan(&bid.LotID, &bid.LotTitle, &bid.BidderID, &bid.Amount, &bid.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan bid: %w", err)
		}
		bids = append(bids, bid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bids: %w", err)
	}

	return bids, nil
}

func (r *repo) GetBidsByLotIDForBidder(ctx context.Context, filter model.BidFilter) ([]model.Bid, error) {
	rows, err := r.repo.Query(ctx, getBidsByLotIDForBidder, filter.LotID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("get bids: %w", err)
	}
	defer rows.Close()

	bids := make([]model.Bid, 0)
	for rows.Next() {
		var bid model.Bid
		if err := rows.Scan(&bid.LotID, &bid.LotTitle, &bid.BidderID, &bid.Amount, &bid.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan bid: %w", err)
		}
		bids = append(bids, bid)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bids: %w", err)
	}

	return bids, nil
}

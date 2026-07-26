package bid

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/bid"
	"context"
	"fmt"
	"log/slog"
)

type Service interface {
	PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error
	GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error)
}

type service struct {
	bidsRepo bid.Repo
	logger   *slog.Logger
}

func NewService(bidsRepo bid.Repo) Service {
	return &service{
		bidsRepo: bidsRepo,
		logger:   slog.With("module", "bid"),
	}
}

func (s *service) PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error {
	if err := s.bidsRepo.PlaceBid(ctx, lotID, bidderID, amount); err != nil {
		return fmt.Errorf("place bid: %w", err)
	}
	return nil
}

func (s *service) GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error) {
	bids, err := s.bidsRepo.GetBidderBids(ctx, bidderID)
	if err != nil {
		return nil, fmt.Errorf("get my bids: %w", err)
	}

	return bids, nil
}
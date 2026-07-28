package bid

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/bid"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"fmt"
	"log/slog"
)

type Service interface {
	PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error
	GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error)
	GetBidsByLotID(ctx context.Context, lotID, sellerID int64) ([]model.Bid, error)
}

type service struct {
	bidsRepo bid.Repo
	logger   *slog.Logger
	lotRepo  lots.Repo
}

func NewService(bidsRepo bid.Repo, lotRepo lots.Repo) Service {
	return &service{
		bidsRepo: bidsRepo,
		logger:   slog.With("module", "bid"),
		lotRepo:  lotRepo,
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

func (s *service) GetBidsByLotID(ctx context.Context, lotID, sellerID int64) ([]model.Bid, error) {
	lotById, err := s.lotRepo.GetById(ctx, lotID)
	if err != nil {
		return nil, fmt.Errorf("get lot: %w", err)
	}
	if lotById.SellerID != sellerID {
		return nil, model.ErrForbidden
	}

	bids, err := s.bidsRepo.GetBidsByLotID(ctx, lotID)
	if err != nil {
		return nil, fmt.Errorf("get bids: %w", err)
	}
	return bids, nil
}

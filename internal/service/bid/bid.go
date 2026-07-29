package bid

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/bid"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"fmt"
	"log/slog"
	"math"
)

const (
	minStepPercent = 5
	statusLive     = "live"
	defaultLimit   = 15
	wholePercent = 100
)

type Service interface {
	PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error
	GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error)
	GetBidsByLotIDForSeller(ctx context.Context, lotID, sellerID int64) ([]model.Bid, error)
	GetBidsByLotIDForBidder(ctx context.Context, lotID int64, page, limit int) ([]model.Bid, error)
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
	if amount != math.Trunc(amount) {
		return model.ErrInvalid
	}
	lot, err := s.lotRepo.GetById(ctx, lotID)
	if err != nil {
		return fmt.Errorf("get lot: %w", err)
	}
	if lot.Status != statusLive {
		return model.ErrLotNotLive
	}

	if lot.WinnerID == 0 {
		if amount < lot.StartPrice {
			return model.ErrBidTooLow
		}
	} else {
		minStep := lot.CurrentPrice * minStepPercent / wholePercent
		requiredBid := lot.CurrentPrice + minStep
		if amount < requiredBid {
			return model.ErrBidTooLow
		}
	}

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

func (s *service) GetBidsByLotIDForSeller(ctx context.Context, lotID, sellerID int64) ([]model.Bid, error) {
	lotById, err := s.lotRepo.GetById(ctx, lotID)
	if err != nil {
		return nil, fmt.Errorf("get lot: %w", err)
	}
	if lotById.SellerID != sellerID {
		return nil, model.ErrForbidden
	}

	bids, err := s.bidsRepo.GetBidsByLotIDForSeller(ctx, lotID)
	if err != nil {
		return nil, fmt.Errorf("get bids: %w", err)
	}
	return bids, nil
}

func (s *service) GetBidsByLotIDForBidder(ctx context.Context, lotID int64, page, limit int) ([]model.Bid, error) {
	if _, err := s.lotRepo.GetById(ctx, lotID); err != nil {
		return nil, fmt.Errorf("get lot: %w", err)
	}
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	filter := model.BidFilter{LotID:  lotID, Limit:  limit, Offset: (page - 1) * limit}
	bids, err := s.bidsRepo.GetBidsByLotIDForBidder(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("get bids: %w", err)
	}
	return bids, nil
}

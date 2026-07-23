package bid

import (
	"auction-house-lotTrio/internal/repository/bid"
	"context"
	"fmt"
	"log/slog"
)

type Service interface {
	PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error
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

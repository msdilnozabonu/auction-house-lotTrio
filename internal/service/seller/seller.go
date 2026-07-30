package seller

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/bid"
	"auction-house-lotTrio/internal/repository/lots"
	"auction-house-lotTrio/internal/repository/seller"
	"context"
	"fmt"
	"log/slog"
)

type sellerService struct {
	sellerRepo seller.Repo
	logger     *slog.Logger
	lotRepo    lots.Repo
	bidRepo    bid.Repo
}

type Service interface {
	GetSellerStats(ctx context.Context, sellerID int64) (model.SellerStats, error)
	CancelLot(ctx context.Context, id, sellerID int64) error
}

func NewService(repo seller.Repo, bidRepo  bid.Repo, lotRepo lots.Repo) Service {
	return &sellerService{
		sellerRepo: repo,
		logger:     slog.With("module", "seller"),
		lotRepo: lotRepo,
		bidRepo:    bidRepo,
	}
}

func (s *sellerService) GetSellerStats(ctx context.Context, sellerID int64) (model.SellerStats, error) {
	stats, err := s.sellerRepo.GetSellerStats(ctx, sellerID)
	if err != nil {
		s.logger.Error("get seller stats", "err", err)
		return model.SellerStats{}, fmt.Errorf("get seller stats: %w", err)
	}

	return stats, nil
}

func (s *sellerService) CancelLot(ctx context.Context, id, sellerID int64) error {
	lotID, err := s.lotRepo.GetById(ctx, id)
	if err != nil {
		s.logger.Error("get lot", "err", err)
		return fmt.Errorf("get lot: %w", err)
	}
	if sellerID != lotID.SellerID {
		return model.ErrForbidden
	}

	if lotID.Status != "live" && lotID.Status != "draft" {
		return model.ErrStatusNotChanged
	}

	bids, err := s.bidRepo.GetBidsByLotIDForSeller(ctx, id)
	if err != nil {
		return fmt.Errorf("get bids: %w", err)
	}

	if len(bids) > 0 {
		return model.ErrHasBids
	}

	err = s.sellerRepo.CanceledLot(ctx, id, sellerID)
	if err != nil {
		s.logger.Error("canceled-lot", "err", err)
		return fmt.Errorf("canceled-lot: %w", err)
	}
	return nil
}
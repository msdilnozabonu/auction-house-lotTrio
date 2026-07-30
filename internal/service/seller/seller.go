package seller

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/seller"
	"context"
	"fmt"
	"log/slog"
)

type sellerService struct {
	sellerRepo seller.Repo
	logger     *slog.Logger
}

type Service interface {
	GetSellerStats(ctx context.Context, sellerID int64) (model.SellerStats, error)
}

func NewService(repo seller.Repo) Service {
	return &sellerService{
		sellerRepo: repo,
		logger:     slog.With("module", "seller"),
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

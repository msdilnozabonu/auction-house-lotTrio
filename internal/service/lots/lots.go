package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type LotService interface {
	CreateLot(ctx context.Context, title, description string, startPrice float64,
		photo string, endsAt time.Time, status string, sellerID int64) error
	GetAll(ctx context.Context) ([]model.Lots, error)
}

type lotsService struct {
	lotsRepo lots.LotRepo
}

func NewLotsService(lotsRepo lots.LotRepo) LotService {
	return &lotsService{lotsRepo: lotsRepo}
}

func (s *lotsService) CreateLot(ctx context.Context, title, description string,
	startPrice float64, photo string,
	endsAt time.Time, status string, sellerID int64) error {
	if startPrice <= 0 {
		return model.ErrStartPrice
	}

	if !endsAt.After(time.Now()) {
		return model.ErrClosed
	}

	currentPrice := startPrice
	err := s.lotsRepo.CreateLot(ctx, title, description, startPrice, photo, endsAt, status, sellerID, currentPrice)
	if err != nil {
		return fmt.Errorf("create lot: %w", err)
	}
	return nil
}

func (s *lotsService) GetAll(ctx context.Context) ([]model.Lots, error) {
	items, err := s.lotsRepo.GetAll(ctx)
	if err != nil {
		slog.Error("get all lots", "err", err)
		return nil, fmt.Errorf("get all lots: %w", err)
	}
	return items, nil
}

package wins

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/wins"
	"context"
	"fmt"
	"log/slog"
)

type Service interface {
	GetWins(ctx context.Context, bidderID int64) ([]model.Win, error)
}

type service struct {
	winsRepo wins.Repo
	logger   *slog.Logger
}

func NewService(winsRepo wins.Repo) Service {
	return &service{
		winsRepo: winsRepo,
		logger:   slog.With("module", "wins"),
	}
}

func (s *service) GetWins(ctx context.Context, bidderID int64) ([]model.Win, error) {
	bidderWins, err := s.winsRepo.GetWins(ctx, bidderID)
	if err != nil {
		return nil, fmt.Errorf("get wins: %w", err)
	}

	return bidderWins, nil
}

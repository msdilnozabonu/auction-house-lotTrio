package watchlist

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/watchlist"
	"context"
	"fmt"
	"log/slog"
)

type Service interface {
	Add(ctx context.Context, userID, lotID int64) error
	Delete(ctx context.Context, userID, lotID int64) error
	GetWatchlist(ctx context.Context, userID int64, page, limit int) ([]model.WatchItem, int, error)
}

type service struct {
	watchRepo watchlist.Repo
	logger    *slog.Logger
}

func NewService(watchRepo watchlist.Repo) Service {
	return &service{
		watchRepo: watchRepo,
		logger:    slog.With("module", "watchlist"),
	}
}

func (s *service) Add(ctx context.Context, userID, lotID int64) error {
	if err := s.watchRepo.Add(ctx, userID, lotID); err != nil {
		return fmt.Errorf("add watch: %w", err)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, userID, lotID int64) error {
	if err := s.watchRepo.Delete(ctx, userID, lotID); err != nil {
		return fmt.Errorf("delete watch: %w", err)
	}

	return nil
}

func (s *service) GetWatchlist(
	ctx context.Context,
	userID int64,
	page, limit int,
) ([]model.WatchItem, int, error) {
	items, total, err := s.watchRepo.GetWatchlist(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("get watchlist: %w", err)
	}

	return items, total, nil
}

package lots

import (
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Service interface {
	CloseExpiredLot(ctx context.Context) (int, error)
	StartScheduler(ctx context.Context, interval time.Duration)
}

type service struct {
	lotsRepo lots.Repo
	logger   *slog.Logger
}

func NewService(lotsRepo lots.Repo) Service {
	return &service{
		lotsRepo: lotsRepo,
		logger:   slog.With("module", "lots"),
	}
}

func (s *service) CloseExpiredLot(ctx context.Context) (int, error) {
	ids, err := s.lotsRepo.FindExpiredLot(ctx, time.Now())
	if err != nil {
		s.logger.Error("failed to find expired lot: ", "error: ", err)
		return 0, fmt.Errorf("find expired lot: %w", err)
	}
	closed := 0
	for _, id := range ids {
		tags, err := s.lotsRepo.CloseLot(ctx, id)
		if err != nil {
			s.logger.Error("failed to close lot: ", "error: ", err)
			return 0, fmt.Errorf("close lot: %w", err)
		}
		if tags {
			s.logger.Info("lot closed: ", "id: ", id)
			closed++
		}
	}
	s.logger.Info("closed lots: ", "count: ", closed)
	return closed, nil
}

func (s *service) StartScheduler(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_, err := s.CloseExpiredLot(ctx)
				if err != nil {
					s.logger.Error("failed to close expired lot: ", "error: ", err)
					return
				}
			case <-ctx.Done():
				s.logger.Info("scheduler stopped")
				return
			}
		}
	}()
}

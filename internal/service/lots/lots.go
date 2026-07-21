package lots

import (
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"time"
)

type Service interface {
	CloseExpiredLot(ctx context.Context) (int, error)
	StartScheduler(ctx context.Context, interval time.Duration)
}

type service struct {
	lotsRepo lots.Repo
}

func NewService(lotsRepo lots.Repo) Service {
	return &service{lotsRepo: lotsRepo}
}

func (s *service) CloseExpiredLot(ctx context.Context) (int, error) {
	ids, err := s.lotsRepo.FindExpiredLot(ctx, time.Now())
	if err != nil {
		return 0, err
	}
	closed := 0
	for _, id := range ids {
		tags, err := s.lotsRepo.CloseLot(ctx, id)
		if err != nil {
			return 0, err
		}
		if tags {
			closed++
		}
	}
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
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

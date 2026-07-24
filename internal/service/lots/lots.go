package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Service interface {
	CloseExpiredLot(ctx context.Context) (int, error)
	StartScheduler(ctx context.Context, interval time.Duration)
	CreateLot(ctx context.Context, title, description, category string, startPrice float64,
		photo string, endsAt time.Time, status string, sellerID int64) error
	GetAll(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error)
	GetByID(ctx context.Context, p model.Lots) (*model.Lots, error)
	UpdateById(ctx context.Context, p model.Lots) error
	DeleteLots(ctx context.Context, p model.Lots) error
	FindLotsForAdmin(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error)
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

func (s *service) CreateLot(ctx context.Context, title, description, category string,
	startPrice float64, photo string,
	endsAt time.Time, status string, sellerID int64) error {
	if startPrice <= 0 {
		return model.ErrStartPrice
	}

	if !endsAt.After(time.Now()) {
		return model.ErrClosed
	}

	currentPrice := startPrice
	err := s.lotsRepo.CreateLot(ctx, title, description, category,
		startPrice, photo, endsAt, status, sellerID, currentPrice)
	if err != nil {
		s.logger.Error("create lot", "err", err)
		return fmt.Errorf("create lot: %w", err)
	}
	return nil
}

func (s *service) GetAll(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error) {
	items, total, err := s.lotsRepo.GetAll(ctx, filter)
	if err != nil {
		s.logger.Error("get all lots", "err", err)
		return nil, 0, fmt.Errorf("get all lots: %w", err)
	}
	return items, total, nil
}

func (s *service) GetByID(ctx context.Context, p model.Lots) (*model.Lots, error) {
	lot, err := s.lotsRepo.GetById(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("get lot by id: %w", err)
	}

	return lot, nil
}

func (s *service) UpdateById(ctx context.Context, p model.Lots) error {
	lot, err := s.lotsRepo.GetById(ctx, p.ID)
	if err != nil {
		s.logger.Error("get lot by id", "err", err)
		return fmt.Errorf("get lot by id: %w", err)
	}

	if p.SellerID != lot.SellerID {
		return model.ErrForbidden
	}

	if p.StartPrice <= 0 {
		return model.ErrStartPrice
	}

	if !p.EndAt.After(time.Now()) {
		return model.ErrClosed
	}

	err = s.lotsRepo.UpdateLot(ctx, p)
	if err != nil {
		s.logger.Error("update lot", "err", err)
		return fmt.Errorf("update lot: %w", err)
	}

	return nil
}

func (s *service) DeleteLots(ctx context.Context, p model.Lots) error {
	lot, err := s.lotsRepo.GetById(ctx, p.ID)
	if err != nil {
		s.logger.Error("get lot by id", "err", err)
		return fmt.Errorf("get lot by id: %w", err)
	}

	if p.SellerID != lot.SellerID {
		return model.ErrForbidden
	}

	err = s.lotsRepo.DeleteLots(ctx, p.ID)
	if err != nil {
		s.logger.Error("delete lots", "err", err)
		return fmt.Errorf("delete lots: %w", err)
	}

	return nil
}

func (s *service) FindLotsForAdmin(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize <1 || filter.PageSize >100{
		filter.PageSize = 50
	}
	items, total, err := s.lotsRepo.FindLotsAdmin(ctx, filter)
	if err != nil {
		s.logger.Error("find lots for admin", "err", err)
		return nil, 0, fmt.Errorf("find lots for admin: %w", err)
	}
	return items, total, nil
}
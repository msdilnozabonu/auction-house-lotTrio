package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"errors"
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
	GetByIDForBid(ctx context.Context, p model.Lots) (*model.Lots, error)
	UpdateById(ctx context.Context, p model.Lots) error
	UpdateStatus(ctx context.Context, sellerID, id int64, status string) error
	DeleteLots(ctx context.Context, p model.Lots) error
	FindLotsForAdmin(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error)
	UploadPhotoById(ctx context.Context, sellerID, id int64, filename string, size int64, contentType string) error
	GetPhoto(ctx context.Context, id, sellerID int64) (string, error)
	ModerateALot(ctx context.Context, id int64, approve bool, reason string) error
	GetPlatformStats(ctx context.Context) (model.PlatformStats, error)
	ExportLots(ctx context.Context, from, to time.Time) ([]model.LotsExport, error)
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

func (s *service) GetByIDForBid(ctx context.Context, p model.Lots) (*model.Lots, error) {
	lot, err := s.lotsRepo.GetByIdForBid(ctx, p.ID)
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

func (s *service) UpdateStatus(ctx context.Context, sellerID, id int64, status string) error {
	lot, err := s.lotsRepo.GetById(ctx, id)
	if err != nil {
		s.logger.Error("get lot by id", "err", err)
		return fmt.Errorf("get lot by id: %w", err)
	}

	if sellerID != lot.SellerID {
		return model.ErrForbidden
	}

	changStatus := map[string]string{
		"draft":  "live",
		"live":   "closed",
		"closed": "cancelled",
	}

	next, ok := changStatus[lot.Status]
	if !ok || next != status {
		return model.ErrStatusNotChanged
	}

	ok, err = s.lotsRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		s.logger.Error("update lot", "err", err)
		return fmt.Errorf("update lot: %w", err)
	}
	if !ok{
		return model.ErrStatusNotChanged
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
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 50
	}
	items, total, err := s.lotsRepo.FindLotsAdmin(ctx, filter)
	if err != nil {
		s.logger.Error("find lots for admin", "err", err)
		return nil, 0, fmt.Errorf("find lots for admin: %w", err)
	}
	return items, total, nil
}
func (s *service) UploadPhotoById(ctx context.Context, sellerID, id int64, filename string,
	size int64, contentType string) error {
	lot, err := s.lotsRepo.GetById(ctx, id)
	if err != nil {
		s.logger.Error("get lot by id", "err", err)
		return fmt.Errorf("get lot by id: %w", err)
	}

	if sellerID != lot.SellerID {
		return model.ErrForbidden
	}

	const maxBite = 5 * 1024 * 1024
	if size > maxBite {
		s.logger.Error("upload photo", "size", size)
		return model.ErrFileTooLarge
	}

	if contentType != "image/jpeg" && contentType != "image/png" {
		return model.ErrInvalidFileType
	}

	err = s.lotsRepo.UploadPhoto(ctx, id, filename)
	if err != nil {
		s.logger.Error("upload photo", "err", err)
		return fmt.Errorf("upload photo: %w", err)
	}
	return nil
}

func (s *service) GetPhoto(ctx context.Context, id, sellerID int64) (string, error) {
	lot, err := s.lotsRepo.GetById(ctx, id)
	if err != nil {
		s.logger.Error("get lot by id", "err", err)
		return "", fmt.Errorf("get lot by id: %w", err)
	}

	if sellerID != lot.SellerID {
		return "", model.ErrForbidden
	}

	photo, err := s.lotsRepo.GetPhoto(ctx, id)
	if err != nil {
		s.logger.Error("get photo", "err", err)
		return "", model.ErrPhotoNotFound
	}

	return photo, nil
}

func (s *service) ModerateALot(ctx context.Context, id int64, approve bool, reason string) error {
	status := "approved"
	if !approve{
		if reason == "" {
			return errors.New("reason is required")
		}
		status = "rejected"
	}
	moderate, err:= s.lotsRepo.ModerateLot(ctx, id, status, reason)
	if err!=nil {
		s.logger.Error("moderate lot", "err", err)
		return fmt.Errorf("moderate lot: %w", err)
	}
	if !moderate {
		return model.ErrStatusNotChanged
	}
	return nil
}

func (s *service) GetPlatformStats(ctx context.Context) (model.PlatformStats, error) {
	stats, err := s.lotsRepo.GetPlatformStats(ctx)
	if err != nil {
		s.logger.Error("get platform stats", "err", err)
		return model.PlatformStats{}, fmt.Errorf("get platform stats: %w", err)
	}
	return stats, nil
}

func (s *service) ExportLots(ctx context.Context, from, to time.Time) ([]model.LotsExport, error)  {
	rows, err := s.lotsRepo.ExportLots(ctx, "ends_at", from, to)
	if err != nil {
		s.logger.Error("export lots", "err", err)
		return nil, fmt.Errorf("export lots: %w", err)
	}
	return rows, nil
}
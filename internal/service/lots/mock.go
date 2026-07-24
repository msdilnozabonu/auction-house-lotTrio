package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type Mock struct{
	mock.Mock
}

func (m *Mock) CloseExpiredLot(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *Mock) StartScheduler(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

func (m *Mock) CreateLot(ctx context.Context, title, description, category string,
	startPrice float64, photo string, endsAt time.Time, status string, sellerID int64) error {
	args := m.Called(ctx, title, description, category, startPrice, photo, endsAt, status, sellerID)
	return args.Error(0)
}

func (m *Mock) GetAll(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]model.Lots), args.Int(1), args.Error(2)
}

func (m *Mock) GetByID(ctx context.Context, p model.Lots) (*model.Lots, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(*model.Lots), args.Error(1)
}

func (m *Mock) UpdateById(ctx context.Context, p model.Lots) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *Mock) DeleteLots(ctx context.Context, p model.Lots) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *Mock) FindLotsForAdmin(ctx context.Context, filter model.LotsFilter) ([]model.Lots, int, error) {
	args := m.Called(ctx, filter)
		return args.Get(0).([]model.Lots), args.Int(1), args.Error(2)
}

var _ Service = (*Mock)(nil)

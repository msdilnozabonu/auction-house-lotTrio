// nolint:wrapcheck
package lots

import (
	"auction-house-lotTrio/internal/model"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

const (
	errorIndex = 2
)

type Mock struct {
	mock.Mock
}

func (m *Mock) CancelLot(ctx context.Context, id int64, reason string) error {
	args := m.Called(ctx, id, reason)
		return args.Error(0)
}

func (m *Mock) ReportLot(ctx context.Context, lotId, reporterId int64, reason string) error {
	args := m.Called(ctx, lotId, reporterId, reason)
		return args.Error(0)
}

func (m *Mock) GetListOfReports(ctx context.Context) ([]model.ReportLot, error) {
	args := m.Called(ctx)
		return args.Get(0).([]model.ReportLot), args.Error(1)
}

func (m *Mock) ExportLots(ctx context.Context, from, to time.Time) ([]model.LotsExport, error) {
	args := m.Called(ctx, from, to)
		return args.Get(0).([]model.LotsExport), args.Error(1)
}

func (m *Mock) GetPlatformStats(ctx context.Context) (model.PlatformStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(model.PlatformStats), args.Error(1)
}

func (m *Mock) GetMineLots(ctx context.Context, sellerID int64, filter model.LotsFilter) ([]model.Lots, int, error) {
	m.Called(ctx, sellerID, filter)
	return nil, 0, nil
}

func (m *Mock) ModerateALot(ctx context.Context, id int64, approve bool, reason string) error {
	args := m.Called(ctx, id, approve, reason)
	return args.Error(0)
}

func (m *Mock) UploadPhotoById(ctx context.Context, sellerID, id int64, filename string,
	size int64, contentType string) error {
	args := m.Called(ctx, sellerID, id, filename, size, contentType)
	return args.Error(0)
}

func (m *Mock) GetPhoto(ctx context.Context, id, sellerID int64) (string, error) {
	args := m.Called(ctx, id, sellerID)
	return args.String(0), args.Error(1)
}

func (m *Mock) GetByIDForBid(ctx context.Context, p model.Lots) (*model.Lots, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(*model.Lots), args.Error(1)
}

func (m *Mock) UpdateStatus(ctx context.Context, sellerID, id int64, status string) error {
	args := m.Called(ctx, sellerID, id, status)
	return args.Error(0)
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
	return args.Get(0).([]model.Lots), args.Int(1), args.Error(errorIndex)
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
	return args.Get(0).([]model.Lots), args.Int(1), args.Error(errorIndex)
}

var _ Service = (*Mock)(nil)

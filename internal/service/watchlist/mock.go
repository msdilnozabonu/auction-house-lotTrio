//nolint:wrapcheck
package watchlist

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

const (
	errorIndex = 2
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Add(ctx context.Context, userID, lotID int64) error {
	args := m.Called(ctx, userID, lotID)
	return args.Error(0)
}

func (m *MockService) Delete(ctx context.Context, userID, lotID int64) error {
	args := m.Called(ctx, userID, lotID)
	return args.Error(0)
}

func (m *MockService) GetWatchlist(ctx context.Context, userID int64, page, limit int) ([]model.WatchItem, int, error) {
	args := m.Called(ctx, userID, page, limit)
	return args.Get(0).([]model.WatchItem), args.Int(1), args.Error(errorIndex)
}

var _ Service = (*MockService)(nil)

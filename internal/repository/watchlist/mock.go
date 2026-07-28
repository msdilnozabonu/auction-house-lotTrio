//nolint:wrapcheck
package watchlist

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

const (
	resultIndex = 0
	totalIndex  = 1
	errorIndex  = 2
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Add(ctx context.Context, userID, lotID int64) error {
	args := m.Called(ctx, userID, lotID)
	return args.Error(resultIndex)
}

func (m *MockRepo) Delete(ctx context.Context, userID, lotID int64) error {
	args := m.Called(ctx, userID, lotID)
	return args.Error(resultIndex)
}

func (m *MockRepo) GetWatchlist(
	ctx context.Context,
	userID int64,
	page, limit int,
) ([]model.WatchItem, int, error) {
	args := m.Called(ctx, userID, page, limit)

	return args.Get(resultIndex).([]model.WatchItem),
		args.Int(totalIndex),
		args.Error(errorIndex)
}

var _ Repo = (*MockRepo)(nil)

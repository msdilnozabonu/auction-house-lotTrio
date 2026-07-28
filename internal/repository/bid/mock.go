//nolint:wrapcheck
package bid

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetBidsByLotID(ctx context.Context, lotID int64) ([]model.Bid, error) {
	args := m.Called(ctx, lotID)
	return args.Get(0).([]model.Bid), args.Error(1)
}

func (m *MockRepo) PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error {
	args := m.Called(ctx, lotID, bidderID, amount)
	return args.Error(0)
}

func (m *MockRepo) GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error) {
	args := m.Called(ctx, bidderID)
	bids, _ := args.Get(0).([]model.Bid)
	return bids, args.Error(1)
}

var _ Repo = (*MockRepo)(nil)
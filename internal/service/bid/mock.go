package bid

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error {
	args := m.Called(ctx, lotID, bidderID, amount)
	return args.Error(0)
}

func (m *MockService) GetBidderBids(ctx context.Context, bidderID int64) ([]model.Bid, error) {
	args := m.Called(ctx, bidderID)
	bids, _ := args.Get(0).([]model.Bid)
	 return bids, args.Error(1)
}

func (m *MockService) GetBidsByLotID(ctx context.Context, lotID, sellerID int64) ([]model.Bid, error) {
	args := m.Called(ctx, lotID, sellerID)
	 bids, _ := args.Get(0).([]model.Bid)
	 return bids, args.Error(1)
}

var _ Service = (*MockService)(nil)
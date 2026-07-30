package bid

import (
	"auction-house-lotTrio/internal/model"
	repoBid "auction-house-lotTrio/internal/repository/bid"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_PlaceBid(t *testing.T) {

	t.Run("success", func(t *testing.T) {
		m := new(repoBid.MockRepo)
		n := new(lots.MockRepo)

		n.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:           1,
			Status:       "live",
			StartPrice:   50,
			CurrentPrice: 50,
			WinnerID:     0,
		}, nil)

		m.On("PlaceBid",
			mock.Anything,
			int64(1),
			int64(2),
			100.0,
		).Return(nil)

		svc := NewService(m, n)
		err := svc.PlaceBid(context.Background(), 1, 2, 100)
		require.NoError(t, err)
		m.AssertExpectations(t)
		n.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		m := new(repoBid.MockRepo)
		n := new(lots.MockRepo)

		n.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:           1,
			Status:       "live",
			StartPrice:   50,
			CurrentPrice: 50,
			WinnerID:     0,
		}, nil)

		m.On("PlaceBid",
			mock.Anything,
			int64(1),
			int64(2),
			100.0,
		).Return(errors.New("db error"))

		svc := NewService(m, n)
		err := svc.PlaceBid(context.Background(), 1, 2, 100)
		require.ErrorContains(t, err, "place bid")
		m.AssertExpectations(t)
		n.AssertExpectations(t)
	})
}

func TestService_GetBidderBids(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)
		expected := []model.Bid{
			{
				LotID:    1,
				LotTitle: "Test lot",
				Amount:   150,
			},
		}

		bidsRepo.On("GetBidderBids", mock.Anything, int64(2)).Return(expected, nil)

		svc := NewService(bidsRepo, lotRepo)
		actual, err := svc.GetBidderBids(context.Background(), 2)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
		bidsRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		bidsRepo.On("GetBidderBids", mock.Anything, int64(2)).
			Return([]model.Bid(nil), errors.New("db error"))

		svc := NewService(bidsRepo, lotRepo)
		actual, err := svc.GetBidderBids(context.Background(), 2)
		require.Nil(t, actual)
		require.Error(t, err)
		require.ErrorContains(t, err, "get my bids")
		require.ErrorContains(t, err, "db error")
		bidsRepo.AssertExpectations(t)
	})
}

func TestService_GetBidsByLotIDForSeller(t *testing.T) {
	t.Run("lot not found", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).
			Return((*model.Lots)(nil), errors.New("record not found"))

		svc := NewService(bidsRepo, lotRepo)
		bids, err := svc.GetBidsByLotIDForSeller(context.Background(), 1, 10)
		require.Nil(t, bids)
		require.ErrorContains(t, err, "get lot")
		lotRepo.AssertExpectations(t)
	})

	t.Run("forbidden", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:       1,
			SellerID: 100,
		}, nil)

		svc := NewService(bidsRepo, lotRepo)
		bids, err := svc.GetBidsByLotIDForSeller(context.Background(), 1, 200)
		require.Nil(t, bids)
		require.ErrorIs(t, err, model.ErrForbidden)
		lotRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)
		expected := []model.Bid{
			{
				LotID:    1,
				LotTitle: "House",
				Amount:   100,
			},
		}

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:       1,
			SellerID: 10,
		}, nil)

		bidsRepo.On("GetBidsByLotIDForSeller", mock.Anything, int64(1)).Return(expected, nil)

		svc := NewService(bidsRepo, lotRepo)
		actual, err := svc.GetBidsByLotIDForSeller(context.Background(), 1, 10)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
		lotRepo.AssertExpectations(t)
		bidsRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:       1,
			SellerID: 10,
		}, nil)

		bidsRepo.On("GetBidsByLotIDForSeller", mock.Anything, int64(1)).
			Return([]model.Bid(nil), errors.New("db error"))

		svc := NewService(bidsRepo, lotRepo)
		actual, err := svc.GetBidsByLotIDForSeller(context.Background(), 1, 10)
		require.Nil(t, actual)
		require.ErrorContains(t, err, "get bids")
		require.ErrorContains(t, err, "db error")
		lotRepo.AssertExpectations(t)
		bidsRepo.AssertExpectations(t)
	})
}

func TestService_GetBidsByLotIDForBidder(t *testing.T) {
	t.Run("lot not found", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).
			Return((*model.Lots)(nil), errors.New("record not found"))

		svc := NewService(bidsRepo, lotRepo)
		bids, err := svc.GetBidsByLotIDForBidder(context.Background(), 1, 1, 10)
		require.Nil(t, bids)
		require.ErrorContains(t, err, "get lot")
		lotRepo.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		expected := []model.Bid{
			{
				LotID:    1,
				LotTitle: "House",
				Amount:   100,
			},
		}

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{ID: 1}, nil)

		bidsRepo.On("GetBidsByLotIDForBidder", mock.Anything,
			model.BidFilter{
				LotID:  1,
				Limit:  10,
				Offset: 0,
			}).Return(expected, nil)

		svc := NewService(bidsRepo, lotRepo)
		actual, err := svc.GetBidsByLotIDForBidder(context.Background(), 1, 1, 10)
		require.NoError(t, err)
		require.Equal(t, expected, actual)
		lotRepo.AssertExpectations(t)
		bidsRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{ID: 1}, nil)

		bidsRepo.On("GetBidsByLotIDForBidder", mock.Anything,
			model.BidFilter{
				LotID:  1,
				Limit:  10,
				Offset: 0,
			}).Return([]model.Bid(nil), errors.New("db error"))

		svc := NewService(bidsRepo, lotRepo)
		actual, err := svc.GetBidsByLotIDForBidder(context.Background(), 1, 1, 10)
		require.Nil(t, actual)
		require.ErrorContains(t, err, "get bids")
		require.ErrorContains(t, err, "db error")
		lotRepo.AssertExpectations(t)
		bidsRepo.AssertExpectations(t)
	})

	t.Run("invalid page uses default", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{ID: 1}, nil)

		bidsRepo.On("GetBidsByLotIDForBidder", mock.Anything,
			model.BidFilter{
				LotID:  1,
				Limit:  10,
				Offset: 0,
			},
		).
			Return([]model.Bid{}, nil)

		svc := NewService(bidsRepo, lotRepo)
		_, err := svc.GetBidsByLotIDForBidder(context.Background(), 1, 0, 10)
		require.NoError(t, err)
		lotRepo.AssertExpectations(t)
		bidsRepo.AssertExpectations(t)
	})

	t.Run("invalid limit uses default", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{ID: 1}, nil)

		bidsRepo.On("GetBidsByLotIDForBidder", mock.Anything,
			model.BidFilter{
				LotID:  1,
				Limit:  15,
				Offset: 0,
			}).Return([]model.Bid{}, nil)

		svc := NewService(bidsRepo, lotRepo)
		_, err := svc.GetBidsByLotIDForBidder(context.Background(), 1, 1, 0)
		require.NoError(t, err)
		lotRepo.AssertExpectations(t)
		bidsRepo.AssertExpectations(t)
	})
}

func TestService_PlaceBidValidation(t *testing.T) {
	t.Run("invalid amount", func(t *testing.T) {
		svc := NewService(new(repoBid.MockRepo), new(lots.MockRepo))
		err := svc.PlaceBid(context.Background(), 1, 2, 100.5)
		require.ErrorIs(t, err, model.ErrInvalid)
	})

	t.Run("get lot error", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).
			Return((*model.Lots)(nil), errors.New("db error"))

		svc := NewService(bidsRepo, lotRepo)
		err := svc.PlaceBid(context.Background(), 1, 2, 100)
		require.ErrorContains(t, err, "get lot")
		lotRepo.AssertExpectations(t)
	})

	t.Run("lot not live", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:     1,
			Status: "closed",
		}, nil)

		svc := NewService(bidsRepo, lotRepo)
		err := svc.PlaceBid(context.Background(), 1, 2, 100)
		require.ErrorIs(t, err, model.ErrLotNotLive)
		lotRepo.AssertExpectations(t)
	})

	t.Run("first bid too low", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)

		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:         1,
			Status:     "live",
			StartPrice: 100,
			WinnerID:   0,
		}, nil)

		svc := NewService(bidsRepo, lotRepo)
		err := svc.PlaceBid(context.Background(), 1, 2, 90)
		require.ErrorIs(t, err, model.ErrBidTooLow)
		lotRepo.AssertExpectations(t)
	})

	t.Run("bid below minimum step", func(t *testing.T) {
		bidsRepo := new(repoBid.MockRepo)
		lotRepo := new(lots.MockRepo)
		lotRepo.On("GetById", mock.Anything, int64(1)).Return(&model.Lots{
			ID:           1,
			Status:       "live",
			WinnerID:     10,
			CurrentPrice: 100,
		}, nil)

		svc := NewService(bidsRepo, lotRepo)
		err := svc.PlaceBid(context.Background(), 1, 2, 104)
		require.ErrorIs(t, err, model.ErrBidTooLow)
		lotRepo.AssertExpectations(t)
	})
}

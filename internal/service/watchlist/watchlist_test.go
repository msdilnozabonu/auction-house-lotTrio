package watchlist

import (
	"auction-house-lotTrio/internal/model"
	repo "auction-house-lotTrio/internal/repository/watchlist"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Add(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)

		mockRepo.On("Add", mock.Anything, int64(1), int64(2)).Return(nil)
		err := service.Add(context.Background(), 1, 2)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)

		mockRepo.On("Add", mock.Anything, int64(1), int64(2)).Return(errors.New("db error"))
		err := service.Add(context.Background(), 1, 2)
		require.Error(t, err)
		require.ErrorContains(t, err, "add watch")
		require.ErrorContains(t, err, "db error")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)

		mockRepo.On("Delete", mock.Anything, int64(1), int64(2)).Return(nil)
		err := service.Delete(context.Background(), 1, 2)
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)
		mockRepo.On("Delete", mock.Anything, int64(1), int64(2)).Return(errors.New("db error"))

		err := service.Delete(context.Background(), 1, 2)
		require.Error(t, err)
		require.ErrorContains(t, err, "delete watch")
		require.ErrorContains(t, err, "db error")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetWatchlist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)

		expected := []model.WatchItem{
			{
				LotID:        1,
				LotTitle:     "MacBook",
				CurrentPrice: 1500,
				Status:       "live",
				CreatedAt:    time.Now(),
			},
		}
		mockRepo.On("GetWatchlist", mock.Anything, int64(1), 1, 10).Return(expected, 1, nil)

		items, total, err := service.GetWatchlist(context.Background(), 1, 1, 10)
		require.NoError(t, err)
		require.Equal(t, expected, items)
		require.Equal(t, 1, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)
		mockRepo.On("GetWatchlist", mock.Anything, int64(1), 1, 10).
			Return([]model.WatchItem{}, 0, errors.New("db error"))
		items, total, err := service.GetWatchlist(context.Background(), 1, 1, 10)
		require.Nil(t, items)
		require.Zero(t, total)
		require.Error(t, err)
		require.ErrorContains(t, err, "get watchlist")
		require.ErrorContains(t, err, "db error")
		mockRepo.AssertExpectations(t)
	})
}

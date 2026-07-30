package wins

import (
	"auction-house-lotTrio/internal/model"
	repo "auction-house-lotTrio/internal/repository/wins"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetWins(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)
		expected := []model.Win{
			{
				LotID:      1,
				LotTitle:   "MacBook",
				WinningBid: 1500,
				ClosedAt:   time.Now(),
			},
		}
		mockRepo.On("GetWins", mock.Anything, int64(10)).Return(expected, nil)
		wins, err := service.GetWins(context.Background(), 10)

		require.NoError(t, err)
		require.Equal(t, expected, wins)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(repo.MockRepo)
		service := NewService(mockRepo)
		mockRepo.On("GetWins", mock.Anything, int64(10)).Return([]model.Win{}, errors.New("db error"))
		wins, err := service.GetWins(context.Background(), 10)
		require.Nil(t, wins)
		require.Error(t, err)
		require.ErrorContains(t, err, "get wins")
		require.ErrorContains(t, err, "db error")
		mockRepo.AssertExpectations(t)
	})
}
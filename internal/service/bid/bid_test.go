package bid

import (
	repoBid "auction-house-lotTrio/internal/repository/bid"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_PlaceBid(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := new(repoBid.MockRepo)

		m.On("PlaceBid",
			mock.Anything,
			int64(1),
			int64(2),
			100.0,
		).Return(nil)

		svc := NewService(m)
		err := svc.PlaceBid(context.Background(), 1, 2, 100)
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		m := new(repoBid.MockRepo)

		m.On("PlaceBid",
			mock.Anything,
			int64(1),
			int64(2),
			100.0,
		).Return(errors.New("db error"))

		svc := NewService(m)
		err := svc.PlaceBid(context.Background(), 1, 2, 100)
		require.ErrorContains(t, err, "place bid")
		m.AssertExpectations(t)
	})
}
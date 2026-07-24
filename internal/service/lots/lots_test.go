package lots

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/lots"
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CloseExpiredLot_Success(t *testing.T) {
	m := new(lots.MockRepo)

	expiredIds := []int64{1, 2, 3}
	m.On("FindExpiredLot", mock.Anything, mock.AnythingOfType("time.Time")).Return(expiredIds, nil)
	m.On("CloseLot", mock.Anything, int64(1)).Return(true, nil)
	m.On("CloseLot", mock.Anything, int64(2)).Return(true, nil)
	m.On("CloseLot", mock.Anything, int64(3)).Return(true, nil)

	s := NewService(m)
	lot, err := s.CloseExpiredLot(context.Background())

	require.NoError(t, err)
	require.Equal(t, 3, lot)
	m.AssertExpectations(t)
}

func TestService_CloseExpiredLot_SkipAlreadyClosed(t *testing.T) {
	m := new(lots.MockRepo)

	expiredIds := []int64{1, 2}
	m.On("FindExpiredLot", mock.Anything, mock.AnythingOfType("time.Time")).Return(expiredIds, nil)
	m.On("CloseLot", mock.Anything, int64(1)).Return(true, nil)
	m.On("CloseLot", mock.Anything, int64(2)).Return(false, nil)

	s := NewService(m)
	lot, err := s.CloseExpiredLot(context.Background())

	require.NoError(t, err)
	require.Equal(t, 1, lot)
	m.AssertExpectations(t)
}
func TestService_CloseExpiredLot_FakeError(t *testing.T) {
	m := new(lots.MockRepo)

	m.On("FindExpiredLot", mock.Anything, mock.AnythingOfType("time.Time")).Return([]int64(nil), errors.New("database error"))

	s := NewService(m)
	closed, err := s.CloseExpiredLot(context.Background())
	require.Error(t, err)
	require.Equal(t, 0, closed)
	m.AssertExpectations(t)
}

func TestService_CloseExpiredLot_CloseLotError(t *testing.T) {
	m := new(lots.MockRepo)

	m.On("FindExpiredLot", mock.Anything, mock.AnythingOfType("time.Time")).Return([]int64{1}, nil)
	m.On("CloseLot", mock.Anything, int64(1)).Return(false, errors.New("database error"))

	s := NewService(m)
	closed, err := s.CloseExpiredLot(context.Background())
	require.Error(t, err)
	require.Equal(t, 0, closed)
	m.AssertExpectations(t)
}

func TestService_StartScheduler_Success(t *testing.T) {
	m := new(lots.MockRepo)

	m.On("FindExpiredLot", mock.Anything, mock.AnythingOfType("time.Time")).Return([]int64{1}, nil)
	m.On("CloseLot", mock.Anything, int64(1)).Return(true, nil)

	s := NewService(m)

	sigCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s.StartScheduler(sigCtx, 10*time.Millisecond)

	time.Sleep(50 * time.Millisecond)
	cancel()

	m.AssertExpectations(t)
}

func TestLotsService_CreateLot(t *testing.T) {
	m := new(lots.MockRepo)
	t.Run("success", func(t *testing.T) {
		m.On("CreateLot", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		svc := NewService(m)
		err := svc.CreateLot(t.Context(), "Phone", "good phone", "electronics", 12000,
			"https://i.pinimg.com/474x/bd/a8/0e/bda80e9324bd6d5c83b84b6eac5a1e5d.jpg", time.Now().Add(24*time.Hour), "live", int64(1))
		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("invalid price", func(t *testing.T) {
		svc := NewService(m)
		err := svc.CreateLot(t.Context(), "Phone", "good phone", "electronics", 0,
			"https://i.pinimg.com/474x/bd/a8/0e/bda80e9324bd6d5c83b84b6eac5a1e5d.jpg", time.Now().Add(24*time.Hour), "live", int64(1))
		require.Error(t, err)
		require.Equal(t, err, model.ErrStartPrice)
	})

	t.Run("ends at in the past", func(t *testing.T) {
		svc := NewService(m)
		err := svc.CreateLot(t.Context(), "Phone", "good phone", "electronics", 12000,
			"https://i.pinimg.com/474x/bd/a8/0e/bda80e9324bd6d5c83b84b6eac5a1e5d.jpg", time.Now().Add(-24*time.Hour), "live", int64(1))
		require.Error(t, err)
		require.Equal(t, err, model.ErrClosed)
	})
}

func TestLotsService_GetAll(t *testing.T) {
	m := new(lots.MockRepo)
	m.On("GetAll", mock.Anything, mock.Anything).
		Return([]model.Lots{
			{ID: 1},
			{Title: "test1"},
		}, 2, nil)
	svc := NewService(m)
	items, total, err := svc.GetAll(t.Context(), model.LotsFilter{})
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, 2, total)
}

func TestService_FindLotsForAdmin(t *testing.T) {
	m := new(lots.MockRepo)
	svc := NewService(m)

	t.Run("successful", func(t *testing.T) {
		filter := model.LotsFilter{Page: 1, Limit: 10, PageSize: 10}
		expectedLots := []model.Lots{{ID: 1, Title: "Admin Lot"}}
		m.On("FindLotsAdmin", mock.Anything, filter).Return(expectedLots, 1, nil).Once()

		items, total, err := svc.FindLotsForAdmin(t.Context(), filter)
		require.NoError(t, err)
		require.Equal(t, expectedLots, items)
		require.Equal(t, 1, total)
	})

	t.Run("normalization", func(t *testing.T) {
		filter := model.LotsFilter{}
		expectedFilter := model.LotsFilter{Page: 1, Limit: 50, PageSize: 50}
		m.On("FindLotsAdmin", mock.Anything, expectedFilter).Return([]model.Lots{}, 0, nil).Once()

		_, _, err := svc.FindLotsForAdmin(t.Context(), filter)
		require.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		filter := model.LotsFilter{Page: 1, Limit: 10, PageSize: 10}
		m.On("FindLotsAdmin", mock.Anything, filter).Return([]model.Lots{}, 0, errors.New("db error")).
			Once()

		_, _, err := svc.FindLotsForAdmin(t.Context(), filter)
		require.Error(t, err)
	})

	m.AssertExpectations(t)
}

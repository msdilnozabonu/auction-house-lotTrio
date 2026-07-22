package lots

import (
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

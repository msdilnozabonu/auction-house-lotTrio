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

func TestService_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := new(lots.MockRepo)
		expected := &model.Lots{ID: 1, SellerID: 10}

		m.On("GetById", mock.Anything, int64(1)).
			Return(expected, nil)

		svc := NewService(m /*, logger, если нужен */)

		got, err := svc.GetByID(t.Context(), model.Lots{ID: 1})

		require.NoError(t, err)
		require.Equal(t, expected, got)
		m.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		m := new(lots.MockRepo)
		repoErr := errors.New("not found")

		m.On("GetById", mock.Anything, int64(2)).
			Return((*model.Lots)(nil), repoErr)

		svc := NewService(m)

		got, err := svc.GetByID(t.Context(), model.Lots{ID: 2})

		require.Nil(t, got)
		require.ErrorContains(t, err, "get lot by id")
		m.AssertExpectations(t)
	})
}

func TestService_UpdateById(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10}
		update := model.Lots{
			ID:         1,
			SellerID:   10,
			StartPrice: 100,
			EndAt:      time.Now().Add(24 * time.Hour),
		}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("UpdateLot", mock.Anything, update).Return(nil)

		svc := NewService(m)

		err := svc.UpdateById(t.Context(), update)

		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("forbidden: different seller", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 999}
		update := model.Lots{
			ID:         1,
			SellerID:   10,
			StartPrice: 100,
			EndAt:      time.Now().Add(24 * time.Hour),
		}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.UpdateById(t.Context(), update)

		require.ErrorIs(t, err, model.ErrForbidden)
		m.AssertExpectations(t)
		m.AssertNotCalled(t, "UpdateLot", mock.Anything, mock.Anything)
	})

	t.Run("invalid start price", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10}
		update := model.Lots{
			ID:         1,
			SellerID:   10,
			StartPrice: 0,
			EndAt:      time.Now().Add(24 * time.Hour),
		}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.UpdateById(t.Context(), update)

		require.ErrorIs(t, err, model.ErrStartPrice)
		m.AssertNotCalled(t, "UpdateLot", mock.Anything, mock.Anything)
	})

	t.Run("lot already closed", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10}
		update := model.Lots{
			ID:         1,
			SellerID:   10,
			StartPrice: 100,
			EndAt:      time.Now().Add(-time.Hour), // уже прошло
		}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.UpdateById(t.Context(), update)

		require.ErrorIs(t, err, model.ErrClosed)
		m.AssertNotCalled(t, "UpdateLot", mock.Anything, mock.Anything)
	})

	t.Run("get by id error", func(t *testing.T) {
		m := new(lots.MockRepo)
		repoErr := errors.New("db error")

		m.On("GetById", mock.Anything, int64(1)).
			Return((*model.Lots)(nil), repoErr)

		svc := NewService(m)

		err := svc.UpdateById(t.Context(), model.Lots{ID: 1})

		require.ErrorContains(t, err, "get lot by id")
		m.AssertNotCalled(t, "UpdateLot", mock.Anything, mock.Anything)
	})

	t.Run("update lot error", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10}
		update := model.Lots{
			ID:         1,
			SellerID:   10,
			StartPrice: 100,
			EndAt:      time.Now().Add(24 * time.Hour),
		}
		repoErr := errors.New("db write error")

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("UpdateLot", mock.Anything, update).Return(repoErr)

		svc := NewService(m)

		err := svc.UpdateById(t.Context(), update)

		require.ErrorContains(t, err, "update lot")
		m.AssertExpectations(t)
	})
}

func TestService_UpdateStatus(t *testing.T) {
	t.Run("success: draft -> live", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10, Status: "draft"}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("UpdateStatus", mock.Anything, int64(1), "live").Return(nil)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "live")

		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("success: live -> closed", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10, Status: "live"}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("UpdateStatus", mock.Anything, int64(1), "closed").Return(nil)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "closed")

		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("forbidden: different seller", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 999, Status: "draft"}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "live")

		require.ErrorIs(t, err, model.ErrForbidden)
		m.AssertExpectations(t)
		m.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("status not changed: invalid transition", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10, Status: "draft"}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "closed")

		require.ErrorIs(t, err, model.ErrStatusNotChanged)
		m.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("status not changed: unknown current status", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10, Status: "cancelled"}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "live")

		require.ErrorIs(t, err, model.ErrStatusNotChanged)
		m.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("get by id error", func(t *testing.T) {
		m := new(lots.MockRepo)
		repoErr := errors.New("db error")

		m.On("GetById", mock.Anything, int64(1)).
			Return((*model.Lots)(nil), repoErr)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "live")

		require.ErrorContains(t, err, "get lot by id")
		m.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("update status error", func(t *testing.T) {
		m := new(lots.MockRepo)
		repoErr := errors.New("db write error")

		existing := &model.Lots{ID: 1, SellerID: 10, Status: "draft"}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("UpdateStatus", mock.Anything, int64(1), "live").Return(repoErr)

		svc := NewService(m)

		err := svc.UpdateStatus(t.Context(), 10, 1, "live")

		require.ErrorContains(t, err, "update lot")
		m.AssertExpectations(t)
	})
}

func TestService_DeleteLots(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 10}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("DeleteLots", mock.Anything, int64(1)).Return(nil)

		svc := NewService(m)

		err := svc.DeleteLots(t.Context(), model.Lots{ID: 1, SellerID: 10})

		require.NoError(t, err)
		m.AssertExpectations(t)
	})

	t.Run("forbidden: different seller", func(t *testing.T) {
		m := new(lots.MockRepo)

		existing := &model.Lots{ID: 1, SellerID: 999}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)

		svc := NewService(m)

		err := svc.DeleteLots(t.Context(), model.Lots{ID: 1, SellerID: 10})

		require.ErrorIs(t, err, model.ErrForbidden)
		m.AssertExpectations(t)
		m.AssertNotCalled(t, "DeleteLots", mock.Anything, mock.Anything)
	})

	t.Run("get by id error", func(t *testing.T) {
		m := new(lots.MockRepo)
		repoErr := errors.New("db error")

		m.On("GetById", mock.Anything, int64(1)).
			Return((*model.Lots)(nil), repoErr)

		svc := NewService(m)

		err := svc.DeleteLots(t.Context(), model.Lots{ID: 1})

		require.ErrorContains(t, err, "get lot by id")
		m.AssertNotCalled(t, "DeleteLots", mock.Anything, mock.Anything)
	})

	t.Run("delete lots error", func(t *testing.T) {
		m := new(lots.MockRepo)
		repoErr := errors.New("db delete error")

		existing := &model.Lots{ID: 1, SellerID: 10}

		m.On("GetById", mock.Anything, int64(1)).Return(existing, nil)
		m.On("DeleteLots", mock.Anything, int64(1)).Return(repoErr)

		svc := NewService(m)

		err := svc.DeleteLots(t.Context(), model.Lots{ID: 1, SellerID: 10})

		require.ErrorContains(t, err, "delete lots")
		m.AssertExpectations(t)
	})
}
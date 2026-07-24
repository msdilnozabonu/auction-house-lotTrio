package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/session"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Refresh_Success(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)
	refresh, hash, err := newRefreshToken()
	require.NoError(t, err)

	sessionRepo.
		On("GetSessionByTokenHash", mock.Anything, hash).
		Return(session.Session{
			ID:        1,
			UserID:    2,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

	sessionRepo.
		On("GetUserRole", mock.Anything, int64(2)).
		Return("admin", nil)

	sessionRepo.
		On("RotateSessionToken", mock.Anything, int64(1), mock.Anything, mock.Anything).
		Return(nil)

	svc := NewService(userRepo, sessionRepo)
	tokens, err := svc.Refresh(t.Context(), refresh)
	require.NoError(t, err)
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)
	sessionRepo.AssertExpectations(t)
}

func TestService_Refresh_SessionNotFound(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)

	refresh, hash, err := newRefreshToken()
	require.NoError(t, err)

	sessionRepo.
		On("GetSessionByTokenHash", mock.Anything, hash).
		Return(session.Session{}, model.ErrSessionNotFound)

	svc := NewService(userRepo, sessionRepo)
	_, err = svc.Refresh(t.Context(), refresh)
	require.ErrorIs(t, err, model.ErrInvalidToken)
	sessionRepo.AssertExpectations(t)
}

func TestService_Refresh_ExpiredSession(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)
	refresh, hash, err := newRefreshToken()
	require.NoError(t, err)
	sessionRepo.
		On("GetSessionByTokenHash", mock.Anything, hash).
		Return(session.Session{
			ID:        1,
			UserID:    2,
			ExpiresAt: time.Now().Add(-time.Hour),
		}, nil)

	svc := NewService(userRepo, sessionRepo)
	_, err = svc.Refresh(t.Context(), refresh)
	require.ErrorIs(t, err, model.ErrInvalidToken)
	sessionRepo.AssertExpectations(t)
}

func TestService_Refresh_GetUserRoleError(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)
	refresh, hash, err := newRefreshToken()
	require.NoError(t, err)

	sessionRepo.
		On("GetSessionByTokenHash", mock.Anything, hash).
		Return(session.Session{
			ID:        1,
			UserID:    2,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

	sessionRepo.
		On("GetUserRole", mock.Anything, int64(2)).
		Return("", errors.New("db error"))
	svc := NewService(userRepo, sessionRepo)
	_, err = svc.Refresh(t.Context(), refresh)
	require.ErrorContains(t, err, "get user role")
	sessionRepo.AssertExpectations(t)
}

func TestService_Refresh_RotateSessionError(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)
	refresh, hash, err := newRefreshToken()
	require.NoError(t, err)

	sessionRepo.
		On("GetSessionByTokenHash", mock.Anything, hash).
		Return(session.Session{
			ID:        1,
			UserID:    2,
			ExpiresAt: time.Now().Add(time.Hour),
		}, nil)

	sessionRepo.
		On("GetUserRole", mock.Anything, int64(2)).
		Return("admin", nil)

	sessionRepo.
		On("RotateSessionToken", mock.Anything, int64(1), mock.Anything, mock.Anything).
		Return(errors.New("db error"))
	svc := NewService(userRepo, sessionRepo)
	_, err = svc.Refresh(t.Context(), refresh)
	require.ErrorContains(t, err, "rotate session token")
	sessionRepo.AssertExpectations(t)
}

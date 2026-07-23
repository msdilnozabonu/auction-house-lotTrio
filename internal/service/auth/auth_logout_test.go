package auth

import (
	"auction-house-lotTrio/internal/repository/session"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Logout_Success(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)
	refreshToken := "refresh-token"

	sessionRepo.
		On("DeleteSession", mock.Anything, hashToken(refreshToken)).
		Return(nil)

	svc := NewService(userRepo, sessionRepo)
	err := svc.Logout(t.Context(), refreshToken)
	require.NoError(t, err)
	sessionRepo.AssertExpectations(t)
}

func TestService_Logout_DeleteSessionError(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)
	refreshToken := "refresh-token"

	sessionRepo.
		On("DeleteSession", mock.Anything, hashToken(refreshToken)).
		Return(errors.New("db error"))

	svc := NewService(userRepo, sessionRepo)
	err := svc.Logout(t.Context(), refreshToken)
	require.ErrorContains(t, err, "delete session")
	sessionRepo.AssertExpectations(t)
}
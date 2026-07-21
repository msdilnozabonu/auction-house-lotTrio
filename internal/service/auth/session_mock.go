//nolint:wrapcheck
package auth

import (
	"auction-house-lotTrio/internal/repository/session"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type SessionMock struct {
	mock.Mock
}

func (m *SessionMock) CreateSession(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
) (int64, error) {
	args := m.Called(ctx, userID, tokenHash, expiresAt)
	return args.Get(0).(int64), args.Error(1)
}

func (m *SessionMock) GetSessionByTokenHash(
	ctx context.Context,
	tokenHash string,
) (session.Session, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *SessionMock) RotateSessionToken(
	ctx context.Context,
	sessionID int64,
	newTokenHash string,
	newExpiresAt time.Time,
) error {
	args := m.Called(ctx, sessionID, newTokenHash, newExpiresAt)
	return args.Error(0)
}

func (m *SessionMock) DeleteSession(
	ctx context.Context,
	tokenHash string,
) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *SessionMock) GetUserRole(
	ctx context.Context,
	userID int64,
) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

var _ session.Repo = (*SessionMock)(nil)

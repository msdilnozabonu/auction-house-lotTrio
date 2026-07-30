//nolint:wrapcheck
package auth

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRegister struct {
	mock.Mock
}

func (m *MockRegister) Register(ctx context.Context, login, password string, role string) error {
	args := m.Called(ctx, login, password, role)
	return args.Error(0)
}

func (m *MockRegister) Login(ctx context.Context, login, password string) (TokenPair, error) {
	args := m.Called(ctx, login, password)
	return args.Get(0).(TokenPair), args.Error(1)
}

func (m *MockRegister) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	return args.Get(0).(TokenPair), args.Error(1)
}

func (m *MockRegister) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockRegister) ValidateAccessToken(token string) (int64, string, error) {
	return 0, "", nil
}

var _ Service = (*MockRegister)(nil)

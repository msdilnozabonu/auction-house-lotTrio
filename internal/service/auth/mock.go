package auth

import (
	"auction-house-lotTrio/internal/repository/user"
	"context"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) Create(ctx context.Context, login, password string, role string) error {
	args := m.Called(ctx, login, password, role)
	return args.Error(0)
}

func (m *Mock) GetUserByLogin(ctx context.Context, login string) (user.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(user.User), args.Error(1)
}

var _ user.Repo = (*Mock)(nil)

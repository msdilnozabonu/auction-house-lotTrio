//nolint:wrapcheck
package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/user"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Create(ctx context.Context, login, password string, role string) error {
	args := m.Called(ctx, login, password, role)
	return args.Error(0)
}

func (m *MockService) GetUserByLogin(ctx context.Context, login string) (user.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(user.User), args.Error(1)
}

func (m *MockService) GetByID(ctx context.Context, id int64) (model.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.User), args.Error(1)
}

var _ user.Repo = (*MockService)(nil)

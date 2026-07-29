//nolint:wrapcheck
package user

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Me(ctx context.Context, userID int64) (model.User, error) {
	args := m.Called(ctx, userID)
	 return args.Get(0).(model.User), args.Error(1)
}

var _ Service = (*MockService)(nil)
package lots

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) FindExpiredLot(ctx context.Context, now time.Time) ([]int64, error) {
	args := m.Called(ctx, now)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRepo) CloseLot(ctx context.Context, id int64) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

var _ Repo = (*MockRepo)(nil)

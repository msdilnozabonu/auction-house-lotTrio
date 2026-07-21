package lots

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

func (m *Mock) CloseExpiredLot(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *Mock) StartScheduler(ctx context.Context, interval time.Duration) {
	m.Called(ctx, interval)
}

var _ Service = (*Mock)(nil)

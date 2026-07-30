//nolint:wrapcheck
package wins

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) GetWins(ctx context.Context, bidderID int64) ([]model.Win, error) {
	args := m.Called(ctx, bidderID)
	return args.Get(0).([]model.Win), args.Error(1)
}

var _ Service = (*MockService)(nil)

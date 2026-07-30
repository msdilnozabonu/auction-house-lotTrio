//nolint:wrapcheck
package wins

import (
	"auction-house-lotTrio/internal/model"
	"context"

	"github.com/stretchr/testify/mock"
)

const (
	resultIndex = 0
	errorIndex  = 1
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetWins(ctx context.Context, bidderID int64) ([]model.Win, error) {
	args := m.Called(ctx, bidderID)
	return args.Get(resultIndex).([]model.Win), args.Error(errorIndex)
}

var _ Repo = (*MockRepo)(nil)

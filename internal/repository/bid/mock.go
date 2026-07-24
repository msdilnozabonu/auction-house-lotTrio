//nolint:wrapcheck
package bid

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) PlaceBid(ctx context.Context, lotID, bidderID int64, amount float64) error {
	args := m.Called(ctx, lotID, bidderID, amount)
	return args.Error(0)
}

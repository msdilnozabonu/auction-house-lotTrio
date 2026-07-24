package user

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/service/auth"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Me_Success(t *testing.T) {
	repo := new(auth.Mock)
	repo.On("GetByID", mock.Anything, int64(1)).
		Return(model.User{ID: 1, Login: "login"}, nil)

	svc := New(repo)
	u, err := svc.Me(t.Context(), 1)

	require.NoError(t, err)
	require.Equal(t, int64(1), u.ID)
	repo.AssertExpectations(t)
}

func TestService_Me_NotFound(t *testing.T) {
	repo := new(auth.Mock)
	repo.On("GetByID", mock.Anything, int64(1)).
		Return(model.User{}, errors.New("not found"))

	svc := New(repo)
	_, err := svc.Me(t.Context(), 1)

	require.ErrorContains(t, err, "get user by id")
	repo.AssertExpectations(t)
}

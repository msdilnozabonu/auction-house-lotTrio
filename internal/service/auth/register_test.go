package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/user"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	m := new(Mock)
	m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	svc := NewService(m)
	err := svc.Register(t.Context(), "login", "passpass", "admin")

	require.NoError(t, err)
	m.AssertExpectations(t)
}

func TestService_IncorrectLenLogin(t *testing.T) {
	m := new(Mock)
	svc := NewService(m)
	err := svc.Register(t.Context(), "lo", "passpass", "admin")

	require.ErrorIs(t, err, model.ErrLenLogin)
}

func TestService_IncorrectLenPassword(t *testing.T) {
	m := new(Mock)
	svc := NewService(m)
	err := svc.Register(t.Context(), "login", "pasass", "admin")

	require.ErrorIs(t, err, model.ErrLenPass)
}

func TestService_RepoError(t *testing.T) {
	m := new(Mock)
	m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("repo error"))
	svc := NewService(m)
	err := svc.Register(t.Context(), "login", "passpass", "admin")
	require.Error(t, err)
	m.AssertExpectations(t)
}

func TestService_GetUserNotFound(t *testing.T) {
	m := new(Mock)
	m.On("GetUserByLogin", mock.Anything, mock.Anything).
		Return(user.User{}, errors.New("user not found"))
	svc := NewService(m)
	_, err := svc.Login(t.Context(), "login", "passpass")

	require.Error(t, err)
	m.AssertExpectations(t)
}

func TestService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(" correct password"), bcrypt.MinCost)

	m := new(Mock)
	m.On("GetUserByLogin", mock.Anything, mock.Anything).
		Return(user.User{Password: string(hash)}, nil)

	svc := NewService(m)
	_, err := svc.Login(t.Context(), "login", "passpass")

	require.ErrorIs(t, err, model.ErrIncorrectPassword)
	m.AssertExpectations(t)
}

func TestService_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)

	m := new(Mock)
	m.On("GetUserByLogin", mock.Anything, mock.Anything).
		Return(user.User{Login: "login", Password: string(hash), Role: "admin"}, nil)

	svc := NewService(m)
	token, err := svc.Login(t.Context(), "login", "correctpass")

	require.NoError(t, err)
	require.NotEmpty(t, token)
	m.AssertExpectations(t)
}

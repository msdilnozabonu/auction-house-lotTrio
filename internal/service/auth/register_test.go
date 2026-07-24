package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/session"
	"auction-house-lotTrio/internal/repository/user"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestService_Register(t *testing.T) {
	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	svc := NewService(m, sessionRepo)
	err := svc.Register(t.Context(), "login", "passpass", "admin")

	require.NoError(t, err)
	m.AssertExpectations(t)
}

func TestService_Register_DefaultRole(t *testing.T) {
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)

	userRepo.
		On("Create", mock.Anything, "login", mock.Anything, "bidder").
		Return(nil)

	svc := NewService(userRepo, sessionRepo)
	err := svc.Register(t.Context(), "login", "password", "")
	require.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestService_IncorrectLenLogin(t *testing.T) {
	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	svc := NewService(m, sessionRepo)
	err := svc.Register(t.Context(), "lo", "passpass", "admin")

	require.ErrorIs(t, err, model.ErrLenLogin)
}

func TestService_IncorrectLenPassword(t *testing.T) {
	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	svc := NewService(m, sessionRepo)
	err := svc.Register(t.Context(), "login", "pasass", "admin")

	require.ErrorIs(t, err, model.ErrLenPass)
}

func TestService_RepoError(t *testing.T) {
	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	m.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("repo error"))
	svc := NewService(m, sessionRepo)
	err := svc.Register(t.Context(), "login", "passpass", "admin")
	require.Error(t, err)
	m.AssertExpectations(t)
}

func TestService_GetUserNotFound(t *testing.T) {
	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	m.On("GetUserByLogin", mock.Anything, mock.Anything).
		Return(user.User{}, errors.New("user not found"))
	svc := NewService(m, sessionRepo)
	_, err := svc.Login(t.Context(), "login", "passpass")

	require.Error(t, err)
	m.AssertExpectations(t)
}

func TestService_Login_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(" correct password"), bcrypt.MinCost)

	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	m.On("GetUserByLogin", mock.Anything, mock.Anything).
		Return(user.User{Password: string(hash)}, nil)

	svc := NewService(m, sessionRepo)
	_, err := svc.Login(t.Context(), "login", "passpass")

	require.ErrorIs(t, err, model.ErrIncorrectPassword)
	m.AssertExpectations(t)
}

func TestService_Login_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)

	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	m.On("GetUserByLogin", mock.Anything, mock.Anything).
		Return(user.User{Login: "login", Password: string(hash), Role: "admin"}, nil)
	sessionRepo.
		On("CreateSession", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(int64(1), nil)

	svc := NewService(m, sessionRepo)
	token, err := svc.Login(t.Context(), "login", "correctpass")

	require.NoError(t, err)
	require.NotEmpty(t, token)
	m.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestService_Login_CreateSessionError(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	userRepo := new(Mock)
	sessionRepo := new(session.SessionMock)

	userRepo.
		On("GetUserByLogin", mock.Anything, "login").
		Return(user.User{
			ID:       1,
			Login:    "login",
			Password: string(hash),
			Role:     "admin",
		}, nil)

	sessionRepo.
		On("CreateSession", mock.Anything, int64(1), mock.Anything, mock.Anything).
		Return(int64(0), errors.New("db error"))

	svc := NewService(userRepo, sessionRepo)
	_, err := svc.Login(t.Context(), "login", "correctpass")
	require.ErrorContains(t, err, "create session")
	userRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestService_Register_PasswordTooLong(t *testing.T) {
	m := new(Mock)
	sessionRepo := new(session.SessionMock)
	svc := NewService(m, sessionRepo)
	err := svc.Register(t.Context(), "login", strings.Repeat("a", 73), "admin")
	require.ErrorContains(t, err, "generate password hash")
}

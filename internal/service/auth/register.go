package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/user"
	"context"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	userRepo user.Repo
	tokens   map[string]user.User
	mu       sync.Mutex
}

type Service interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) (string, error)
}

func NewService(userRepo user.Repo) Service {
	return &service{
		userRepo: userRepo,
		tokens:   make(map[string]user.User),
		mu:       sync.Mutex{},
	}
}

func (s *service) Register(ctx context.Context, login, password string) error {
	if len(login) < 3 {
		return model.ErrLenLogin
	}

	if len(password) < 8 {
		return model.ErrLenPass
	}

	pasHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	err = s.userRepo.Create(ctx, login, string(pasHash))
	if err != nil {
		return err
	}

	return nil
}

func (s *service) Login(ctx context.Context, login, password string) (string, error) {
	userInfo, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(password))
	if err != nil {
		return "", model.ErrIncorrectPassword
	}

	token := s.newToken(userInfo)
	return token, nil
}

func (s *service) newToken(userInfo user.User) string {
	claims := jwt.MapClaims{
		"uid":  userInfo.ID,
		"role": userInfo.Role,
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return tokenString
}

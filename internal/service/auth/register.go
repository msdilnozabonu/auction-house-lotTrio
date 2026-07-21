package auth

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/repository/session"
	"auction-house-lotTrio/internal/repository/user"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = 15 * time.Minute
	refreshTokenDuration = 240 * time.Hour

	minLoginLen       = 3
	minPasswordLen    = 8
	refreshTokenBytes = 32
)

type service struct {
	userRepo    user.Repo
	sessionRepo session.Repo
	jwtSecret   []byte
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Service interface {
	Register(ctx context.Context, login, password string, role string) error
	Login(ctx context.Context, login, password string) (TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateAccessToken(token string) (int64, string, error)
}

func NewService(userRepo user.Repo, sessionRepo session.Repo) Service {
	return &service{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtSecret:   []byte(os.Getenv("JWT_SECRET")),
	}
}

func (s *service) Register(ctx context.Context, login, password string, role string) error {
	if len(login) < minLoginLen {
		return model.ErrLenLogin
	}

	if len(password) < minPasswordLen {
		return model.ErrLenPass
	}

	pasHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return fmt.Errorf("generate password hash: %w", err)
	}

	err = s.userRepo.Create(ctx, login, string(pasHash), role)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *service) Login(ctx context.Context, login, password string) (TokenPair, error) {
	userInfo, err := s.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return TokenPair{}, fmt.Errorf("get user by login: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(userInfo.Password), []byte(password))
	if err != nil {
		return TokenPair{}, model.ErrIncorrectPassword
	}

	access, err := s.newAccessToken(userInfo.ID, userInfo.Role)
	if err != nil {
		return TokenPair{}, fmt.Errorf("create access token: %w", err)
	}
	refresh, hash, err := newRefreshToken()
	if err != nil {
		return TokenPair{}, fmt.Errorf("generate refresh token: %w", err)
	}

	_, err = s.sessionRepo.CreateSession(ctx, userInfo.ID, hash, time.Now().Add(refreshTokenDuration))
	if err != nil {
		return TokenPair{}, fmt.Errorf("create session: %w", err)
	}

	return TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *service) newAccessToken(userID int64, role string) (string, error) {
	claims := jwt.MapClaims{
		"uid":  userID,
		"role": role,
		"exp":  time.Now().Add(accessTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return tokenString, nil
}

func newRefreshToken() (string, string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate random refresh token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	hash := hashToken(raw)
	return raw, hash, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
	hash := hashToken(refreshToken)

	sess, err := s.sessionRepo.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return TokenPair{}, model.ErrInvalidToken
	}

	if time.Now().After(sess.ExpiresAt) {
		return TokenPair{}, model.ErrInvalidToken
	}

	role, err := s.sessionRepo.GetUserRole(ctx, sess.UserID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("get user role: %w", err)
	}

	newRefresh, newHash, err := newRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	err = s.sessionRepo.RotateSessionToken(ctx, sess.ID, newHash, time.Now().Add(refreshTokenDuration))
	if err != nil {
		return TokenPair{}, fmt.Errorf("rotate session token: %w", err)
	}

	access, err := s.newAccessToken(sess.UserID, role)
	if err != nil {
		return TokenPair{}, err
	}
	return TokenPair{AccessToken: access, RefreshToken: newRefresh}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if err := s.sessionRepo.DeleteSession(ctx, hashToken(refreshToken)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *service) ValidateAccessToken(token string) (int64, string, error) {
	tok, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil || !tok.Valid {
		return 0, "", model.ErrInvalidToken
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", model.ErrInvalidToken
	}

	uid, ok := claims["uid"].(float64)
	if !ok {
		return 0, "", model.ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok {
		return 0, "", model.ErrInvalidToken
	}

	return int64(uid), role, nil
}

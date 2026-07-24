package auth

import (
	"auction-house-lotTrio/internal/model"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestService_ValidateAccessToken_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	svc := NewService(new(Mock), nil)
	token, err := svc.(*service).newAccessToken(1, "admin")
	require.NoError(t, err)

	uid, role, err := svc.ValidateAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, int64(1), uid)
	require.Equal(t, "admin", role)
}

func TestService_ValidateAccessToken_InvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")

	svc := NewService(new(Mock), nil)
	_, _, err := svc.ValidateAccessToken("invalid-token")
	require.ErrorIs(t, err, model.ErrInvalidToken)
}

func TestService_ValidateAccessToken_WrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret1")
	svc := NewService(new(Mock), nil)
	token, err := svc.(*service).newAccessToken(10, "seller")
	require.NoError(t, err)

	os.Setenv("JWT_SECRET", "secret2")

	svc2 := NewService(new(Mock), nil)
	_, _, err = svc2.ValidateAccessToken(token)
	require.ErrorIs(t, err, model.ErrInvalidToken)
}

func TestService_ValidateAccessToken_MissingUID(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	svc := NewService(new(Mock), nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "admin",
	})
	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, _, err = svc.ValidateAccessToken(tokenString)
	require.ErrorIs(t, err, model.ErrInvalidToken)
}

func TestService_ValidateAccessToken_UnexpectedSigningMethod(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	svc := NewService(new(Mock), nil)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"uid":  1,
		"role": "admin",
	})

	tokenString, _ := token.SigningString()
	_, _, err := svc.ValidateAccessToken(tokenString + ".")
	require.ErrorIs(t, err, model.ErrInvalidToken)
}

func TestService_ValidateAccessToken_MissingRole(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")
	svc := NewService(new(Mock), nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": float64(1),
	})
	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	_, _, err = svc.ValidateAccessToken(tokenString)
	require.ErrorIs(t, err, model.ErrInvalidToken)
}

package response

import (
	"auction-house-lotTrio/internal/model"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestRespondError_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrNotFound)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.JSONEq(t,`{"error":"`+model.ErrNotFound.Error()+`"}`, w.Body.String())
}

func TestRespondError_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrInvalid)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRespondError_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrUserNotFound)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t,`{"error":"invalid login or password"}`, w.Body.String())
}

func TestRespondError_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrForbidden)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRespondError_Conflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrAlreadyWatching)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestRespondError_Internal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, errors.New("database error"))
	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRespondJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondJSON(c, http.StatusCreated, gin.H{"message": "ok"})
	require.Equal(t, http.StatusCreated, w.Code)
	require.JSONEq(t, `{"message":"ok"}`, w.Body.String())
}

func TestRespondError_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrInvalidToken)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRespondError_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrUnauthorized)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRespondError_Outbid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrOutbid)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestRespondError_Closed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrClosed)
	require.Equal(t, http.StatusConflict, w.Code)
}

func TestRespondError_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := newContext()
	RespondError(c, model.ErrInvalidID)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

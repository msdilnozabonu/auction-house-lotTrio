// nolint
package response

import (
	"auction-house-lotTrio/internal/model"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const errorKey = "error"

func RespondError(c *gin.Context, err error) {
	switch {
	// 404 Not Found
	case errors.Is(err, model.ErrNotFound):
		RespondJSON(c, http.StatusNotFound, gin.H{errorKey: err.Error()})
	// 400 Bad Request
	case errors.Is(err, model.ErrInvalid),
		errors.Is(err, model.ErrLenLogin),
		errors.Is(err, model.ErrLenPass):
		RespondJSON(c, http.StatusBadRequest, gin.H{errorKey: err.Error()})
	// 401 Unauthorized
	case errors.Is(err, model.ErrInvalidToken),
		errors.Is(err, model.ErrInvalidAuth):
		RespondJSON(c, http.StatusUnauthorized, gin.H{errorKey: err.Error()})
	// 401 Unauthorized (hide authentication details)
	case errors.Is(err, model.ErrIncorrectPassword),
		errors.Is(err, model.ErrUserNotFound):
		RespondJSON(c, http.StatusUnauthorized, gin.H{errorKey: "invalid login or password"})
	case errors.Is(err, model.ErrUnauthorized):
		RespondJSON(c, http.StatusUnauthorized, gin.H{errorKey: err.Error()})
	// 403 Forbidden
	case errors.Is(err, model.ErrForbidden):
		RespondJSON(c, http.StatusForbidden, gin.H{errorKey: err.Error()})
	// 409 Conflict
	case errors.Is(err, model.ErrOutbid):
		RespondJSON(c, http.StatusConflict, gin.H{errorKey: err.Error()})
	case errors.Is(err, model.ErrClosed):
		RespondJSON(c, http.StatusConflict, gin.H{errorKey: err.Error()})
	case errors.Is(err, model.ErrUserAlreadyExists):
		RespondJSON(c, http.StatusConflict, gin.H{errorKey: err.Error()})
	// 500 Internal Server Error
	default:
		RespondJSON(c, http.StatusInternalServerError, gin.H{errorKey: err.Error()})
	}
}
func RespondJSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

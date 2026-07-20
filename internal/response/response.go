package response

import (
	"auction-house-lotTrio/internal/model"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RespondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		RespondJSON(c, http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrInvalid):
		RespondJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrForbidden):
		RespondJSON(c, http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrOutbid):
		RespondJSON(c, http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrClosed):
		RespondJSON(c, http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrUserAlreadyExists):
		RespondJSON(c, http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrLenLogin):
		RespondJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrLenPass):
		RespondJSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrInvalidToken):
		RespondJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrIncorrectPassword):
		RespondJSON(c, http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
	case errors.Is(err, model.ErrInvalidAuth):
		RespondJSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, model.ErrUserNotFound):
		RespondJSON(c, http.StatusUnauthorized, gin.H{"error": "invalid login or password"})
	default:
		RespondJSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
func RespondJSON(c *gin.Context, status int, body gin.H) {
	c.JSON(status, body)
}

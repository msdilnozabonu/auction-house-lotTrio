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
	default:
		RespondJSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
func RespondJSON(c *gin.Context, status int, body gin.H) {
	c.JSON(status, body)
}

package seller

import (
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/seller"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type handler struct {
	sellerService seller.Service
}

type Handler interface {
	GetSellerStats(c *gin.Context)
	CanceledLot(c *gin.Context)
}

func NewHandler(sellerService seller.Service) Handler {
	return &handler{sellerService: sellerService}
}

func (h *handler) GetSellerStats(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, errors.New("no user_id found"))
		return
	}

	sellerID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, errors.New("no user_id found"))
		return
	}

	s, err := h.sellerService.GetSellerStats(c.Request.Context(), sellerID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	response.RespondJSON(c, http.StatusOK, s)
}

func (h *handler) CanceledLot(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, errors.New("no user_id found"))
		return
	}
	sellerID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, errors.New("no user_id found"))
		return
	}

	idPar := c.Param("id")
	id, err := strconv.ParseInt(idPar, 10, 64)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	err = h.sellerService.CancelLot(c.Request.Context(), id, sellerID)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "canceled"})
}
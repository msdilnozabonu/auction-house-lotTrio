package bid

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/bid"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	messageKey = "message"
)

type Handler interface {
	PlaceBid(c *gin.Context)
	GetBiddersBids(c *gin.Context)
	GetBidsByLotID(c *gin.Context)
}

type handler struct {
	bidsService bid.Service
	logger      *slog.Logger
}

func NewHandler(bidsService bid.Service) Handler {
	return &handler{
		bidsService: bidsService,
		logger:      slog.With("module", "bids"),
	}
}

// PlaceBid godoc
//
//	@Summary		Сделать ставку
//	@Description	Размещает ставку на активный лот
//	@Tags			bids
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int					true	"ID лота"
//	@Param			input	body	PlaceBidRequest		true	"Сумма ставки"
//	@Success		201
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		409
//	@Security		BearerAuth
//	@Router			/lots/:id/bid [post]
func (h *handler) PlaceBid(c *gin.Context) {
	var req PlaceBidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.RespondError(c, err)
		return
	}

	lotID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}

	bidderID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}

	err = h.bidsService.PlaceBid(c.Request.Context(), lotID, bidderID, req.Amount)
	if err != nil {
		h.logger.Error("place bid", "err", err)
		response.RespondError(c, err)
		return
	}

	response.RespondJSON(c, http.StatusCreated, gin.H{
		messageKey: "bid placed successfully!",
	})
}

// GetBiddersBids godoc
//
//	@Summary		Получить свои ставки
//	@Description	Возвращает все ставки авторизованного участника
//	@Tags			bids
//	@Produce		json
//	@Success		200		{array}		model.Bid
//	@Failure		401
//	@Failure		403
//	@Failure		500
//	@Security		BearerAuth
//	@Router			/bids [get]
func (h *handler) GetBiddersBids(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}

	bidderID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}

	bids, err := h.bidsService.GetBidderBids(c.Request.Context(), bidderID)
	if err != nil {
		h.logger.Error("get my bids", "err", err)
		response.RespondError(c, err)
		return
	}

	response.RespondJSON(c, http.StatusOK, bids)
}

func (h *handler) GetBidsByLotID(c *gin.Context) {
	lotID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.RespondError(c, err)
		return
	}

	userID, ok := c.Get("user_id")
	if !ok {
		response.RespondError(c, model.ErrUnauthorized)
		return
	}

	sellerID, ok := userID.(int64)
	if !ok {
		response.RespondError(c, model.ErrForbidden)
		return
	}

	bids, err := h.bidsService.GetBidsByLotID(c.Request.Context(), lotID, sellerID)
	if err != nil {
		h.logger.Error("get my bids", "err", err)
		response.RespondError(c, err)
		return
	}
	response.RespondJSON(c, http.StatusOK, bids)
}

type PlaceBidRequest struct {
	Amount float64 `binding:"required" json:"amount"`
}

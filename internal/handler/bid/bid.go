package bid

import (
	"auction-house-lotTrio/internal/model"
	"auction-house-lotTrio/internal/response"
	"auction-house-lotTrio/internal/service/bid"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	messageKey = "message"
	defaultPage      = 1
	defaultPageLimit = 15
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
//	@Description	Размещает ставку на активный лот. Сумма ставки должна соответствовать минимальному шагу ставки.
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
		response.RespondError(c, model.ErrInvalidID)
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

// GetBidsByLotID godoc
//
// @Summary      Получить ставки по лоту
// @Description  Возвращает историю ставок по указанному лоту.
// @Description  Участник получает историю с пагинацией.
// @Description  Продавец получает ставки по своему лоту.
// @Tags         bids
// @Produce      json
// @Param        id      path    int true  "ID лота"
// @Param        page    query   int false "Номер страницы (для участника)"
// @Param        limit   query   int false "Количество элементов на странице (для участника)"
// @Success      200     {array} model.Bid
// @Failure      400
// @Failure      401
// @Failure      403
// @Failure      404
// @Failure      500
// @Security     BearerAuth
// @Router       /lots/{id}/bids [get]
func (h *handler) GetBidsByLotID(c *gin.Context) {
	lotID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.RespondError(c, model.ErrInvalidID)
		return
	}

	role := c.GetString("role")
	var bids []model.Bid
	switch role {
	case "bidder":
		bids, err = h.getBidderBidsByLot(c, lotID)
	case "seller":
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

		bids, err = h.bidsService.GetBidsByLotIDForSeller(c.Request.Context(), lotID, sellerID)

	default:
		response.RespondError(c, model.ErrForbidden)
		return
	}
	if err != nil {
		h.logger.Error("get my bids by lot", "err", err)
		response.RespondError(c, err)
		return
	}
	response.RespondJSON(c, http.StatusOK, bids)
}

func (h *handler) getBidderBidsByLot(c *gin.Context, lotID int64) ([]model.Bid, error) {
	page := defaultPage
	if p, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage))); err == nil && p > 0 {
		page = p
	}

	limit := defaultPageLimit
	if l, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultPageLimit))); err == nil && l > 0 {
		limit = l
	}

	bids, err := h.bidsService.GetBidsByLotIDForBidder(c.Request.Context(), lotID, page, limit)
	if err != nil {
		return nil, fmt.Errorf("get bids by lot for bidder: %w", err)
	}
	return bids, nil
}

type PlaceBidRequest struct {
	Amount float64 `binding:"required" json:"amount"`
}

package handler

import (
	"net/http"
	"strconv"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type PriceHandler struct {
	repo *repository.PriceRepository
}

func NewPriceHandler(repo *repository.PriceRepository) *PriceHandler {
	return &PriceHandler{repo: repo}
}

func (handler *PriceHandler) GetPricesByOfferID(ctx *gin.Context) {
	offerID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	var prices []models.Price
	paymentType := ctx.Query("payment_type")
	if paymentType != "" {
		prices, err = handler.repo.GetByOfferID(ctx.Request.Context(), offerID, models.PaymentType(paymentType))
	} else {
		prices, err = handler.repo.GetByOfferID(ctx.Request.Context(), offerID)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prices"})
		return
	}

	ctx.JSON(http.StatusOK, prices)
}

func (handler *PriceHandler) GetLatestPrice(ctx *gin.Context) {
	offerID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	paymentType := ctx.Query("payment_type")
	var price *models.Price
	if paymentType != "" {
		price, err = handler.repo.GetLatestByOfferID(ctx.Request.Context(), offerID, models.PaymentType(paymentType))
	} else {
		price, err = handler.repo.GetLatestByOfferID(ctx.Request.Context(), offerID)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no prices found for this offer"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get latest price"})
		return
	}

	ctx.JSON(http.StatusOK, price)
}

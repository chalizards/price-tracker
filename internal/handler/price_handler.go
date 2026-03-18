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

func (handler *PriceHandler) GetPricesByProductID(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var prices []models.Price
	paymentType := ctx.Query("payment_type")
	if paymentType != "" {
		prices, err = handler.repo.GetByProductID(ctx.Request.Context(), productID, models.PaymentType(paymentType))
	} else {
		prices, err = handler.repo.GetByProductID(ctx.Request.Context(), productID)
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prices"})
		return
	}

	ctx.JSON(http.StatusOK, prices)
}

func (handler *PriceHandler) GetLatestPrice(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	paymentType := ctx.Query("payment_type")
	var price *models.Price
	if paymentType != "" {
		price, err = handler.repo.GetLatestByProductID(ctx.Request.Context(), productID, models.PaymentType(paymentType))
	} else {
		price, err = handler.repo.GetLatestByProductID(ctx.Request.Context(), productID)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "no prices found for this product"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get latest price"})
		return
	}

	ctx.JSON(http.StatusOK, price)
}

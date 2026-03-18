package handler

import (
	"net/http"
	"strconv"

	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ScrapeHandler struct {
	productRepo     *repository.ProductRepository
	priceRepo       *repository.PriceRepository
	trackingService *service.PriceTrackingService
}

func NewScrapeHandler(productRepo *repository.ProductRepository, priceRepo *repository.PriceRepository, trackingService *service.PriceTrackingService) *ScrapeHandler {
	return &ScrapeHandler{productRepo: productRepo, priceRepo: priceRepo, trackingService: trackingService}
}

func (handler *ScrapeHandler) ScrapeProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := handler.productRepo.GetByID(ctx.Request.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get product"})
		return
	}

	if err := handler.trackingService.ScrapeProduct(ctx.Request.Context(), product); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prices, err := handler.priceRepo.GetByProductID(ctx.Request.Context(), product.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "scraped but failed to fetch prices"})
		return
	}

	// Return latest prices from this scrape (pix + credit)
	limit := 2
	if len(prices) < limit {
		limit = len(prices)
	}

	ctx.JSON(http.StatusOK, prices[:limit])
}

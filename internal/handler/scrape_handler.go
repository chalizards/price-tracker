package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type ScrapeHandler struct {
	storeRepo       *repository.StoreRepository
	priceRepo       *repository.PriceRepository
	trackingService *service.PriceTrackingService
}

func NewScrapeHandler(storeRepo *repository.StoreRepository, priceRepo *repository.PriceRepository, trackingService *service.PriceTrackingService) *ScrapeHandler {
	return &ScrapeHandler{storeRepo: storeRepo, priceRepo: priceRepo, trackingService: trackingService}
}

func (handler *ScrapeHandler) ScrapeStore(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	store, err := handler.storeRepo.GetByID(ctx.Request.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get store"})
		return
	}

	// Use a detached context so the scrape isn't canceled if the HTTP client disconnects.
	scrapeCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := handler.trackingService.ScrapeStore(scrapeCtx, store); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prices, err := handler.priceRepo.GetByStoreID(ctx.Request.Context(), store.ID)
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

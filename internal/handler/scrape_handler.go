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
	offerRepo       *repository.OfferRepository
	priceRepo       *repository.PriceRepository
	trackingService *service.PriceTrackingService
}

func NewScrapeHandler(offerRepo *repository.OfferRepository, priceRepo *repository.PriceRepository, trackingService *service.PriceTrackingService) *ScrapeHandler {
	return &ScrapeHandler{offerRepo: offerRepo, priceRepo: priceRepo, trackingService: trackingService}
}

func (handler *ScrapeHandler) ScrapeOffer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	offer, err := handler.offerRepo.GetByID(ctx.Request.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "offer not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get offer"})
		return
	}

	// Use a detached context so the scrape isn't canceled if the HTTP client disconnects.
	scrapeCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := handler.trackingService.ScrapeOffer(scrapeCtx, offer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	prices, err := handler.priceRepo.GetByOfferID(ctx.Request.Context(), offer.ID)
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

package handler

import (
	"net/http"
	"strconv"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type OfferHandler struct {
	offerRepo   *repository.OfferRepository
	productRepo *repository.ProductRepository
}

func NewOfferHandler(offerRepo *repository.OfferRepository, productRepo *repository.ProductRepository) *OfferHandler {
	return &OfferHandler{offerRepo: offerRepo, productRepo: productRepo}
}

func (handler *OfferHandler) CreateOffer(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if _, err := handler.productRepo.GetByID(ctx.Request.Context(), productID); err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get product"})
		return
	}

	var offer models.Offer
	if err := ctx.ShouldBindJSON(&offer); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offer.ProductID = productID
	if err := handler.offerRepo.Create(ctx.Request.Context(), &offer); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create offer"})
		return
	}

	ctx.JSON(http.StatusCreated, offer)
}

func (handler *OfferHandler) GetOffersByProductID(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	offers, err := handler.offerRepo.GetByProductID(ctx.Request.Context(), productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get offers"})
		return
	}

	ctx.JSON(http.StatusOK, offers)
}

func (handler *OfferHandler) UpdateOffer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	var offer models.Offer
	if err := ctx.ShouldBindJSON(&offer); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	offer.ID = id
	updated, err := handler.offerRepo.Update(ctx.Request.Context(), &offer)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "offer not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update offer"})
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (handler *OfferHandler) DeleteOffer(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offer id"})
		return
	}

	if err := handler.offerRepo.Delete(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "offer not found"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

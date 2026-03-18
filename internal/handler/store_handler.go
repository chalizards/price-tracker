package handler

import (
	"net/http"
	"strconv"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type StoreHandler struct {
	storeRepo   *repository.StoreRepository
	productRepo *repository.ProductRepository
}

func NewStoreHandler(storeRepo *repository.StoreRepository, productRepo *repository.ProductRepository) *StoreHandler {
	return &StoreHandler{storeRepo: storeRepo, productRepo: productRepo}
}

func (handler *StoreHandler) CreateStore(ctx *gin.Context) {
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

	var store models.Store
	if err := ctx.ShouldBindJSON(&store); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store.ProductID = productID
	if err := handler.storeRepo.Create(ctx.Request.Context(), &store); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create store"})
		return
	}

	ctx.JSON(http.StatusCreated, store)
}

func (handler *StoreHandler) GetStoresByProductID(ctx *gin.Context) {
	productID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	stores, err := handler.storeRepo.GetByProductID(ctx.Request.Context(), productID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stores"})
		return
	}

	ctx.JSON(http.StatusOK, stores)
}

func (handler *StoreHandler) UpdateStore(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	var store models.Store
	if err := ctx.ShouldBindJSON(&store); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	store.ID = id
	updated, err := handler.storeRepo.Update(ctx.Request.Context(), &store)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update store"})
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (handler *StoreHandler) DeleteStore(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	if err := handler.storeRepo.Delete(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

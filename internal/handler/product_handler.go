package handler

import (
	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
)

type ProductHandler struct {
	repo *repository.ProductRepository
}

func NewProductHandler(repo *repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (handler *ProductHandler) CreateProduct(ctx *gin.Context) {
	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := handler.repo.Create(ctx.Request.Context(), &product); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	ctx.JSON(http.StatusCreated, product)
}

func (handler *ProductHandler) GetAllProducts(ctx *gin.Context) {
	search := ctx.Query("search")

	if search != "" {
		products, err := handler.repo.GetByName(ctx.Request.Context(), search)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search products"})
			return
		}
		ctx.JSON(http.StatusOK, products)
		return
	}

	products, err := handler.repo.GetAll(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get products"})
		return
	}

	ctx.JSON(http.StatusOK, products)
}

func (handler *ProductHandler) GetProductByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := handler.repo.GetByID(ctx.Request.Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get product"})
		return
	}

	ctx.JSON(http.StatusOK, product)
}

func (handler *ProductHandler) GetActiveProducts(ctx *gin.Context) {
	products, err := handler.repo.GetActive(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active products"})
		return
	}

	ctx.JSON(http.StatusOK, products)
}

func (handler *ProductHandler) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var product models.Product
	if err := ctx.ShouldBindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product.ID = id
	updatedProduct, err := handler.repo.Update(ctx.Request.Context(), &product)
	if err != nil {
		if err == pgx.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	ctx.JSON(http.StatusOK, updatedProduct)
}

func (handler *ProductHandler) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := handler.repo.Delete(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

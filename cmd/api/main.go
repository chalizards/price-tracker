package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/chalizards/price-tracker/internal/handler"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/scheduler"
	"github.com/chalizards/price-tracker/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	scrapeInterval := 60
	if val := os.Getenv("SCRAPE_INTERVAL_MINUTES"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			scrapeInterval = parsed
		}
	}

	db, err := repository.NewPostgresPool(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repositories
	productRepo := repository.NewProductRepository(db)
	priceRepo := repository.NewPriceRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	// Handlers
	productHandler := handler.NewProductHandler(productRepo)
	priceHandler := handler.NewPriceHandler(priceRepo)
	notificationHandler := handler.NewNotificationHandler(notificationRepo)

	router := gin.Default()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Price Tracker API is running",
		})
	})

	api := router.Group("/api")
	{
		// Products
		api.POST("/products", productHandler.CreateProduct)
		api.GET("/products", productHandler.GetAllProducts)
		api.GET("/products/active", productHandler.GetActiveProducts)
		api.GET("/products/:id", productHandler.GetProductByID)
		api.GET("/products/slug/:slug", productHandler.GetProductBySlug)
		api.PUT("/products/:id", productHandler.UpdateProduct)
		api.DELETE("/products/:id", productHandler.DeleteProduct)

		// Prices
		api.GET("/products/:id/prices", priceHandler.GetPricesByProductID)
		api.GET("/products/:id/prices/latest", priceHandler.GetLatestPrice)

		// Notifications
		api.GET("/notifications/unread", notificationHandler.GetUnreadNotifications)
		api.GET("/products/:id/notifications", notificationHandler.GetNotificationsByProductID)
		api.PATCH("/notifications/:id/read", notificationHandler.MarkAsRead)
		api.PATCH("/notifications/read-all", notificationHandler.MarkAllAsRead)
	}

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

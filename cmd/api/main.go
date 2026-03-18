package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/chalizards/price-tracker/internal/cache"
	"github.com/chalizards/price-tracker/internal/handler"
	"github.com/chalizards/price-tracker/internal/repository"
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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is required")
	}

	redisClient, err := cache.NewRedisClient(redisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	geminiSecretName := os.Getenv("GEMINI_SECRET_NAME")
	if geminiSecretName == "" {
		log.Fatal("GEMINI_SECRET_NAME is required")
	}

	geminiAPIKey, err := service.GetGeminiSecret(context.Background(), redisClient, geminiSecretName)
	if err != nil {
		log.Fatalf("Failed to get Gemini secret: %v", err)
	}
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}

	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		googleRedirectURL = "http://localhost:8080/api/auth/google/callback"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	db, err := repository.NewPostgresPool(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repositories
	productRepo := repository.NewProductRepository(db)
	storeRepo := repository.NewStoreRepository(db)
	priceRepo := repository.NewPriceRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Services
	notificationService := service.NewNotificationService(notificationRepo, priceRepo)
	trackingService := service.NewPriceTrackingService(productRepo, storeRepo, priceRepo, notificationService, geminiAPIKey)
	authService := service.NewAuthService(googleClientID, googleClientSecret, googleRedirectURL, jwtSecret, userRepo)

	// Handlers
	productHandler := handler.NewProductHandler(productRepo)
	storeHandler := handler.NewStoreHandler(storeRepo, productRepo)
	priceHandler := handler.NewPriceHandler(priceRepo)
	notificationHandler := handler.NewNotificationHandler(notificationRepo)
	scrapeHandler := handler.NewScrapeHandler(storeRepo, priceRepo, trackingService)
	authHandler := handler.NewAuthHandler(authService, frontendURL)

	router := gin.Default()

	// CORS Middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", frontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Price Tracker API is running",
		})
	})

	api := router.Group("/api")
	{
		// Auth (public)
		api.GET("/auth/google/login", authHandler.GoogleLogin)
		api.GET("/auth/google/callback", authHandler.GoogleCallback)
		api.POST("/auth/logout", authHandler.Logout)

		// Protected routes
		protected := api.Group("/")
		protected.Use(handler.AuthMiddleware(authService, userRepo))
		{
			// Auth
			protected.GET("/auth/me", authHandler.Me)

			// Products
			protected.POST("/products", productHandler.CreateProduct)
			protected.GET("/products", productHandler.GetAllProducts)
			protected.GET("/products/active", productHandler.GetActiveProducts)
			protected.GET("/products/:id", productHandler.GetProductByID)
			protected.PUT("/products/:id", productHandler.UpdateProduct)
			protected.DELETE("/products/:id", productHandler.DeleteProduct)

			// Stores
			protected.POST("/products/:id/stores", storeHandler.CreateStore)
			protected.GET("/products/:id/stores", storeHandler.GetStoresByProductID)
			protected.PUT("/stores/:id", storeHandler.UpdateStore)
			protected.DELETE("/stores/:id", storeHandler.DeleteStore)

			// Prices (by store)
			protected.GET("/stores/:id/prices", priceHandler.GetPricesByStoreID)
			protected.GET("/stores/:id/prices/latest", priceHandler.GetLatestPrice)

			// Scrape (by store)
			protected.POST("/stores/:id/scrape", scrapeHandler.ScrapeStore)

			// Notifications
			protected.GET("/notifications/unread", notificationHandler.GetUnreadNotifications)
			protected.GET("/products/:id/notifications", notificationHandler.GetNotificationsByProductID)
			protected.PATCH("/notifications/:id/read", notificationHandler.MarkAsRead)
			protected.PATCH("/notifications/read-all", notificationHandler.MarkAllAsRead)
		}
	}

	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

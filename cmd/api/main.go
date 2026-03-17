package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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

	geminiSecretName := os.Getenv("GEMINI_SECRET_NAME")
	if geminiSecretName == "" {
		log.Fatal("GEMINI_SECRET_NAME is required")
	}

	geminiAPIKey := service.GetGeminiSecret(geminiSecretName)
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
		frontendURL = "http://localhost:3000" // Default for local frontend (React/Next.js)
	}

	scrapeIntervalStr := os.Getenv("SCRAPE_INTERVAL_MINUTES")
	if scrapeIntervalStr == "" {
		log.Fatal("SCRAPE_INTERVAL_MINUTES is required")
	}
	scrapeInterval, err := strconv.Atoi(scrapeIntervalStr)
	if err != nil {
		log.Fatal("SCRAPE_INTERVAL_MINUTES must be a valid integer")
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
	userRepo := repository.NewUserRepository(db)

	// Services
	notificationService := service.NewNotificationService(notificationRepo, priceRepo)
	trackingService := service.NewPriceTrackingService(productRepo, priceRepo, notificationService, geminiAPIKey)
	authService := service.NewAuthService(googleClientID, googleClientSecret, googleRedirectURL, jwtSecret, userRepo)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Scheduler
	sc := scheduler.NewScheduler(trackingService, scrapeInterval)
	go sc.Start(ctx)

	// Handlers
	productHandler := handler.NewProductHandler(productRepo)
	priceHandler := handler.NewPriceHandler(priceRepo)
	notificationHandler := handler.NewNotificationHandler(notificationRepo)
	scrapeHandler := handler.NewScrapeHandler(productRepo, priceRepo, trackingService)
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
			protected.GET("/products/slug/:slug", productHandler.GetProductBySlug)
			protected.PUT("/products/:id", productHandler.UpdateProduct)
			protected.DELETE("/products/:id", productHandler.DeleteProduct)

			// Prices
			protected.GET("/products/:id/prices", priceHandler.GetPricesByProductID)
			protected.GET("/products/:id/prices/latest", priceHandler.GetLatestPrice)

			// Scrape
			protected.POST("/products/:id/scrape", scrapeHandler.ScrapeProduct)

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

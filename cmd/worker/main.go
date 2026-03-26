package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chalizards/price-tracker/internal/cache"
	"github.com/chalizards/price-tracker/internal/queue"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/service"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		log.Fatal("RABBITMQ_URL is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL is required")
	}

	geminiSecretName := os.Getenv("GEMINI_SECRET_NAME")
	if geminiSecretName == "" {
		log.Fatal("GEMINI_SECRET_NAME is required")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient, err := cache.NewRedisClient(redisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	geminiAPIKey, err := service.GetGeminiSecret(ctx, redisClient, geminiSecretName)
	if err != nil {
		log.Fatalf("Failed to get Gemini secret: %v", err)
	}
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is empty")
	}

	db, err := repository.NewPostgresPool(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	offerRepo := repository.NewOfferRepository(db)
	priceRepo := repository.NewPriceRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)

	notificationService := service.NewNotificationService(notificationRepo, priceRepo)
	trackingService := service.NewPriceTrackingService(productRepo, offerRepo, priceRepo, notificationService, geminiAPIKey)

	rmq, err := queue.NewRabbitMQ(rabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rmq.Close()

	if _, err := rmq.DeclareQueue(queue.ScrapeJobsQueue); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	msgs, err := rmq.Consume(queue.ScrapeJobsQueue)
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}

	log.Println("Worker started, waiting for scrape jobs...")

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Channel closed, stopping worker")
				return
			}

			var job queue.ScrapeJobMessage
			if err := json.Unmarshal(msg.Body, &job); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("Failed to nack message: %v", nackErr)
				}
				continue
			}

			offer, err := offerRepo.GetByID(ctx, job.OfferID)
			if err != nil {
				log.Printf("Failed to get offer %d: %v", job.OfferID, err)
				if nackErr := msg.Nack(false, false); nackErr != nil {
					log.Printf("Failed to nack message: %v", nackErr)
				}
				continue
			}

			if err := trackingService.ScrapeOffer(ctx, offer); err != nil {
				log.Printf("Failed to scrape offer %d: %v", job.OfferID, err)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					log.Printf("Failed to nack message: %v", nackErr)
				}
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("Failed to ack message: %v", err)
			}
			log.Printf("Successfully scraped offer %d (%s)", offer.ID, offer.Name)

		case <-ctx.Done():
			log.Println("Worker stopped")
			return
		}
	}
}

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/chalizards/price-tracker/internal/queue"
	"github.com/chalizards/price-tracker/internal/repository"
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

	intervalStr := os.Getenv("SCRAPE_INTERVAL_MINUTES")
	if intervalStr == "" {
		intervalStr = "120"
	}

	intervalMinutes, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Fatalf("Invalid SCRAPE_INTERVAL_MINUTES: %v", err)
	}

	db, err := repository.NewPostgresPool(databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	offerRepo := repository.NewOfferRepository(db)

	rmq, err := queue.NewRabbitMQ(rabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer rmq.Close()

	if _, err := rmq.DeclareQueue(queue.ScrapeJobsQueue); err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	interval := time.Duration(intervalMinutes) * time.Minute
	log.Printf("Scheduler started (interval: %v)", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			publishScrapeJobs(ctx, offerRepo, rmq)
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		}
	}
}

func publishScrapeJobs(ctx context.Context, offerRepo *repository.OfferRepository, rmq *queue.RabbitMQ) {
	offers, err := offerRepo.GetActiveOffers(ctx)
	if err != nil {
		log.Printf("Failed to get active offers: %v", err)
		return
	}

	if len(offers) == 0 {
		log.Println("No active offers to enqueue")
		return
	}

	for _, o := range offers {
		msg := queue.ScrapeJobMessage{OfferID: o.ID}
		body, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message for offer %d: %v", o.ID, err)
			continue
		}

		if err := rmq.Publish(ctx, queue.ScrapeJobsQueue, body); err != nil {
			log.Printf("Failed to publish job for offer %d: %v", o.ID, err)
			continue
		}

		log.Printf("Enqueued scrape job for offer %d (%s)", o.ID, o.Name)
	}

	log.Printf("Published %d scrape jobs", len(offers))
}

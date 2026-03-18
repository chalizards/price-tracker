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

	storeRepo := repository.NewStoreRepository(db)

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
			publishScrapeJobs(ctx, storeRepo, rmq)
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		}
	}
}

func publishScrapeJobs(ctx context.Context, storeRepo *repository.StoreRepository, rmq *queue.RabbitMQ) {
	stores, err := storeRepo.GetActiveStores(ctx)
	if err != nil {
		log.Printf("Failed to get active stores: %v", err)
		return
	}

	if len(stores) == 0 {
		log.Println("No active stores to enqueue")
		return
	}

	for _, s := range stores {
		msg := queue.ScrapeJobMessage{StoreID: s.ID}
		body, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to marshal message for store %d: %v", s.ID, err)
			continue
		}

		if err := rmq.Publish(ctx, queue.ScrapeJobsQueue, body); err != nil {
			log.Printf("Failed to publish job for store %d: %v", s.ID, err)
			continue
		}

		log.Printf("Enqueued scrape job for store %d (%s)", s.ID, s.Name)
	}

	log.Printf("Published %d scrape jobs", len(stores))
}

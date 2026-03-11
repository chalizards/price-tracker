package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/chalizards/price-tracker/internal/service"
)

type Scheduler struct {
	trackingService *service.PriceTrackingService
	interval        time.Duration
}

func NewScheduler(trackingService *service.PriceTrackingService, intervalMinutes int) *Scheduler {
	return &Scheduler{
		trackingService: trackingService,
		interval:        time.Duration(intervalMinutes) * time.Minute,
	}
}

func (scheduler *Scheduler) Start(ctx context.Context) {
	log.Printf("Scheduler started (interval: %v)", scheduler.interval)

	ticker := time.NewTicker(scheduler.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Scheduler: starting scrape cycle")
			scheduler.trackingService.ScrapeAllActive(ctx)
			log.Println("Scheduler: scrape cycle finished")
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		}
	}
}

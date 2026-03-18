package service

import (
	"context"
	"fmt"
	"log"

	"github.com/chalizards/price-tracker/internal/llm/gemini"
	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/repository"
	"github.com/chalizards/price-tracker/internal/scraper"
)

type PriceTrackingService struct {
	productRepo         *repository.ProductRepository
	storeRepo           *repository.StoreRepository
	priceRepo           *repository.PriceRepository
	notificationService *NotificationService
	geminiAPIKey        string
}

func NewPriceTrackingService(
	productRepo *repository.ProductRepository,
	storeRepo *repository.StoreRepository,
	priceRepo *repository.PriceRepository,
	notificationService *NotificationService,
	geminiAPIKey string,
) *PriceTrackingService {
	return &PriceTrackingService{
		productRepo:         productRepo,
		storeRepo:           storeRepo,
		priceRepo:           priceRepo,
		notificationService: notificationService,
		geminiAPIKey:        geminiAPIKey,
	}
}

func (s *PriceTrackingService) ScrapeStore(ctx context.Context, store *models.Store) error {
	product, err := s.productRepo.GetByID(ctx, store.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	log.Printf("Scraping store: %s (%s)", store.Name, store.URL)

	html, err := scraper.FetchHTML(ctx, store.URL)
	if err != nil {
		s.notificationService.CreateErrorNotification(ctx, product, fmt.Sprintf("failed to fetch page: %v", err))
		return fmt.Errorf("failed to fetch html: %w", err)
	}

	log.Printf("[scrape] HTML length: %d chars", len(html))

	result, err := gemini.ExtractPrice(ctx, s.geminiAPIKey, html, product.Name)
	if err != nil {
		s.notificationService.CreateErrorNotification(ctx, product, fmt.Sprintf("failed to extract price: %v", err))
		return fmt.Errorf("failed to extract price: %w", err)
	}

	for _, entry := range result.Prices {
		price := &models.Price{
			StoreID:     store.ID,
			Price:       entry.Price,
			Currency:    entry.Currency,
			PaymentType: models.PaymentType(entry.PaymentType),
		}

		if err := s.priceRepo.Create(ctx, price); err != nil {
			return fmt.Errorf("failed to save price (%s): %w", entry.PaymentType, err)
		}

		log.Printf("Price saved: %s %.2f (%s) for %s at %s", price.Currency, price.Price, price.PaymentType, product.Name, store.Name)

		s.notificationService.CheckPriceNotifications(ctx, product, store, price)
	}

	return nil
}

func (s *PriceTrackingService) ScrapeAllActive(ctx context.Context) {
	stores, err := s.storeRepo.GetActiveStores(ctx)
	if err != nil {
		log.Printf("Failed to get active stores: %v", err)
		return
	}

	if len(stores) == 0 {
		log.Println("No active stores to scrape")
		return
	}

	for i := range stores {
		if err := s.ScrapeStore(ctx, &stores[i]); err != nil {
			log.Printf("Failed to scrape store %s: %v", stores[i].Name, err)
		}
	}
}

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
	priceRepo           *repository.PriceRepository
	notificationService *NotificationService
	geminiAPIKey        string
}

func NewPriceTrackingService(
	productRepo *repository.ProductRepository,
	priceRepo *repository.PriceRepository,
	notificationService *NotificationService,
	geminiAPIKey string,
) *PriceTrackingService {
	return &PriceTrackingService{
		productRepo:         productRepo,
		priceRepo:           priceRepo,
		notificationService: notificationService,
		geminiAPIKey:        geminiAPIKey,
	}
}

func (s *PriceTrackingService) ScrapeProduct(ctx context.Context, product *models.Product) error {
	log.Printf("Scraping product: %s (%s)", product.Name, product.URL)

	html, err := scraper.FetchHTML(ctx, product.URL)
	if err != nil {
		s.notificationService.CreateErrorNotification(ctx, product, fmt.Sprintf("failed to fetch page: %v", err))
		return fmt.Errorf("failed to fetch html: %w", err)
	}

	log.Printf("[scrape] HTML length: %d chars", len(html))
	if len(html) > 500 {
		log.Printf("[scrape] HTML preview: %s", html[:500])
	} else {
		log.Printf("[scrape] HTML full: %s", html)
	}

	result, err := gemini.ExtractPrice(ctx, s.geminiAPIKey, html, product.Name)
	if err != nil {
		s.notificationService.CreateErrorNotification(ctx, product, fmt.Sprintf("failed to extract price: %v", err))
		return fmt.Errorf("failed to extract price: %w", err)
	}

	for _, entry := range result.Prices {
		price := &models.Price{
			ProductID:   product.ID,
			Price:       entry.Price,
			Currency:    entry.Currency,
			PaymentType: models.PaymentType(entry.PaymentType),
		}

		if err := s.priceRepo.Create(ctx, price); err != nil {
			return fmt.Errorf("failed to save price (%s): %w", entry.PaymentType, err)
		}

		log.Printf("Price saved: %s %.2f (%s) for %s", price.Currency, price.Price, price.PaymentType, product.Name)

		s.notificationService.CheckPriceNotifications(ctx, product, price)
	}

	return nil
}

func (s *PriceTrackingService) ScrapeAllActive(ctx context.Context) {
	products, err := s.productRepo.GetActive(ctx)
	if err != nil {
		log.Printf("Failed to get active products: %v", err)
		return
	}

	if len(products) == 0 {
		log.Println("No active products to scrape")
		return
	}

	for i := range products {
		if err := s.ScrapeProduct(ctx, &products[i]); err != nil {
			log.Printf("Failed to scrape %s: %v", products[i].Name, err)
		}
	}
}

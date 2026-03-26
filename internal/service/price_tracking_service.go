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
	offerRepo           *repository.OfferRepository
	priceRepo           *repository.PriceRepository
	notificationService *NotificationService
	geminiAPIKey        string
}

func NewPriceTrackingService(
	productRepo *repository.ProductRepository,
	offerRepo *repository.OfferRepository,
	priceRepo *repository.PriceRepository,
	notificationService *NotificationService,
	geminiAPIKey string,
) *PriceTrackingService {
	return &PriceTrackingService{
		productRepo:         productRepo,
		offerRepo:           offerRepo,
		priceRepo:           priceRepo,
		notificationService: notificationService,
		geminiAPIKey:        geminiAPIKey,
	}
}

func (s *PriceTrackingService) ScrapeOffer(ctx context.Context, offer *models.Offer) error {
	product, err := s.productRepo.GetByID(ctx, offer.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	log.Printf("Scraping offer: %s (%s)", offer.Name, offer.URL)

	html, err := scraper.FetchHTML(ctx, offer.URL)
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
			OfferID:     offer.ID,
			Price:       entry.Price,
			Currency:    entry.Currency,
			PaymentType: models.PaymentType(entry.PaymentType),
		}

		if err := s.priceRepo.Create(ctx, price); err != nil {
			return fmt.Errorf("failed to save price (%s): %w", entry.PaymentType, err)
		}

		log.Printf("Price saved: %s %.2f (%s) for %s at %s", price.Currency, price.Price, price.PaymentType, product.Name, offer.Name)

		s.notificationService.CheckPriceNotifications(ctx, product, offer, price)
	}

	return nil
}

func (s *PriceTrackingService) ScrapeAllActive(ctx context.Context) {
	offers, err := s.offerRepo.GetActiveOffers(ctx)
	if err != nil {
		log.Printf("Failed to get active offers: %v", err)
		return
	}

	if len(offers) == 0 {
		log.Println("No active offers to scrape")
		return
	}

	for i := range offers {
		if err := s.ScrapeOffer(ctx, &offers[i]); err != nil {
			log.Printf("Failed to scrape offer %s: %v", offers[i].Name, err)
		}
	}
}

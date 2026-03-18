package service

import (
	"context"
	"fmt"
	"log"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/chalizards/price-tracker/internal/repository"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	priceRepo        *repository.PriceRepository
}

func NewNotificationService(notificationRepo *repository.NotificationRepository, priceRepo *repository.PriceRepository) *NotificationService {
	return &NotificationService{notificationRepo: notificationRepo, priceRepo: priceRepo}
}

func (service *NotificationService) CheckPriceNotifications(ctx context.Context, product *models.Product, store *models.Store, newPrice *models.Price) {
	service.checkTargetReached(ctx, product, store, newPrice)
	service.checkPriceDrop(ctx, product, store, newPrice)
}

func (service *NotificationService) CreateErrorNotification(ctx context.Context, product *models.Product, message string) {
	title := fmt.Sprintf("Error scraping %s", product.Name)
	service.CreateNotification(ctx, product, nil, message, title, models.NotificationScrapeError)
}

func (service *NotificationService) checkTargetReached(ctx context.Context, product *models.Product, store *models.Store, newPrice *models.Price) {
	if product.TargetPrice == nil || newPrice.Price > *product.TargetPrice {
		return
	}

	title := fmt.Sprintf("Target price reached for %s at %s", product.Name, store.Name)
	message := fmt.Sprintf("Current price: %s %.2f via %s (target: %.2f)",
		newPrice.Currency, newPrice.Price, newPrice.PaymentType, *product.TargetPrice)

	service.CreateNotification(ctx, product, &newPrice.ID, message, title, models.NotificationTargetReached)
}

func (service *NotificationService) checkPriceDrop(ctx context.Context, product *models.Product, store *models.Store, newPrice *models.Price) {
	previousPrices, err := service.priceRepo.GetByStoreID(ctx, store.ID, newPrice.PaymentType)
	if err != nil || len(previousPrices) < 2 {
		return
	}

	previousPrice := previousPrices[1] // index 0 is the one we just inserted
	if newPrice.Price >= previousPrice.Price {
		return
	}

	drop := previousPrice.Price - newPrice.Price
	title := fmt.Sprintf("Price drop for %s at %s (%s)", product.Name, store.Name, newPrice.PaymentType)
	message := fmt.Sprintf("Price dropped by %s %.2f (from %.2f to %.2f) via %s",
		newPrice.Currency, drop, previousPrice.Price, newPrice.Price, newPrice.PaymentType)
	service.CreateNotification(ctx, product, &newPrice.ID, message, title, models.NotificationPriceDrop)
}

func (service *NotificationService) CreateNotification(ctx context.Context, product *models.Product, priceID *int, message string, title string, notificationType models.NotificationType) {
	notification := &models.Notification{
		ProductID: product.ID,
		PriceID:   priceID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
	}
	if err := service.notificationRepo.Create(ctx, notification); err != nil {
		log.Printf("Failed to create notification: %v", err)
	}
}

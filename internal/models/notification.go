package models

import "time"

type NotificationType string

const (
	NotificationPriceDrop     NotificationType = "price_drop"
	NotificationTargetReached NotificationType = "target_reached"
	NotificationScrapeError   NotificationType = "scrape_error"
)

type Notification struct {
	ID        int              `json:"id"`
	ProductID int              `json:"product_id"`
	PriceID   *int             `json:"price_id,omitempty"`
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
}

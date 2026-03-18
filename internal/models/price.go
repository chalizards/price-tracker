package models

import "time"

type PaymentType string

const (
	PaymentTypePIX    PaymentType = "pix"
	PaymentTypeCredit PaymentType = "credit"
)

type Price struct {
	ID          int         `json:"id"`
	ProductID   int         `json:"product_id"`
	Price       float64     `json:"price"`
	Currency    string      `json:"currency"`
	PaymentType PaymentType `json:"payment_type"`
	ScrapedAt   time.Time   `json:"scraped_at"`
}

package models

import "time"

type Price struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Price     float64   `json:"price"`
	Currency  string    `json:"currency"`
	ScrapedAt time.Time `json:"scraped_at"`
}

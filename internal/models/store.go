package models

import "time"

type Store struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

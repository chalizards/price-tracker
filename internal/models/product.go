package models

import "time"

type Product struct {
	ID          int        `json:"id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Store       string     `json:"store"`
	TargetPrice *float64   `json:"target_price,omitempty"`
	Active      bool       `json:"active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
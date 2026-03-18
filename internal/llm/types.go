package llm

type PriceEntry struct {
	Price       float64 `json:"price"`
	Currency    string  `json:"currency"`
	PaymentType string  `json:"payment_type"`
}

type PriceResult struct {
	Prices []PriceEntry `json:"prices"`
}

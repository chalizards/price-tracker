package repository

import (
	"context"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PriceRepository struct {
	db *pgxpool.Pool
}

func NewPriceRepository(db *pgxpool.Pool) *PriceRepository {
	return &PriceRepository{db: db}
}

func (repo *PriceRepository) Create(ctx context.Context, price *models.Price) error {
	query := `
		INSERT INTO prices (store_id, price, currency, payment_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, store_id, price, currency, payment_type, scraped_at
	`
	return repo.db.QueryRow(ctx, query,
		price.StoreID, price.Price, price.Currency, price.PaymentType,
	).Scan(&price.ID, &price.StoreID, &price.Price, &price.Currency, &price.PaymentType, &price.ScrapedAt)
}

func (repo *PriceRepository) GetByStoreID(ctx context.Context, storeID int, paymentType ...models.PaymentType) ([]models.Price, error) {
	query := `
		SELECT id, store_id, price, currency, payment_type, scraped_at
		FROM prices
		WHERE store_id = $1
	`
	args := []any{storeID}

	if len(paymentType) > 0 && paymentType[0] != "" {
		query += " AND payment_type = $2"
		args = append(args, paymentType[0])
	}

	query += " ORDER BY scraped_at DESC"

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var price models.Price
		err := rows.Scan(&price.ID, &price.StoreID, &price.Price, &price.Currency, &price.PaymentType, &price.ScrapedAt)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (repo *PriceRepository) GetLatestByStoreID(ctx context.Context, storeID int, paymentType ...models.PaymentType) (*models.Price, error) {
	query := `
		SELECT id, store_id, price, currency, payment_type, scraped_at
		FROM prices
		WHERE store_id = $1
	`
	args := []any{storeID}

	if len(paymentType) > 0 && paymentType[0] != "" {
		query += " AND payment_type = $2"
		args = append(args, paymentType[0])
	}

	query += " ORDER BY scraped_at DESC LIMIT 1"

	price := &models.Price{}
	err := repo.db.QueryRow(ctx, query, args...).Scan(
		&price.ID, &price.StoreID, &price.Price, &price.Currency, &price.PaymentType, &price.ScrapedAt,
	)
	if err != nil {
		return nil, err
	}
	return price, nil
}

func (repo *PriceRepository) GetAll(ctx context.Context) ([]models.Price, error) {
	query := `
		SELECT id, store_id, price, currency, payment_type, scraped_at
		FROM prices
		ORDER BY scraped_at DESC
	`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var p models.Price
		err := rows.Scan(&p.ID, &p.StoreID, &p.Price, &p.Currency, &p.PaymentType, &p.ScrapedAt)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

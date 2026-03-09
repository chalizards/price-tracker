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
		INSERT INTO prices (product_id, price, currency)
		VALUES ($1, $2, $3)
		RETURNING id, product_id, price, currency, scraped_at
	`
	return repo.db.QueryRow(ctx, query,
		price.ProductID, price.Price, price.Currency,
	).Scan(&price.ID, &price.ProductID, &price.Price, &price.Currency, &price.ScrapedAt)
}

func (repo *PriceRepository) GetByProductID(ctx context.Context, productID int) ([]models.Price, error) {
	query := `
		SELECT id, product_id, price, currency, scraped_at
		FROM prices
		WHERE product_id = $1
		ORDER BY scraped_at DESC
	`
	rows, err := repo.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var price models.Price
		err := rows.Scan(&price.ID, &price.ProductID, &price.Price, &price.Currency, &price.ScrapedAt)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}
	return prices, nil
}

func (repo *PriceRepository) GetLatestByProductID(ctx context.Context, productID int) (*models.Price, error) {
	query := `
		SELECT id, product_id, price, currency, scraped_at
		FROM prices
		WHERE product_id = $1
		ORDER BY scraped_at DESC
		LIMIT 1
	`
	price := &models.Price{}
	err := repo.db.QueryRow(ctx, query, productID).Scan(
		&price.ID, &price.ProductID, &price.Price, &price.Currency, &price.ScrapedAt,
	)
	if err != nil {
		return nil, err
	}
	return price, nil
}

func (repo *PriceRepository) GetAll(ctx context.Context) ([]models.Price, error) {
	query := `
		SELECT id, product_id, price, currency, scraped_at
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
		err := rows.Scan(&p.ID, &p.ProductID, &p.Price, &p.Currency, &p.ScrapedAt)
		if err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

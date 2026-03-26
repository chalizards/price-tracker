package repository

import (
	"context"
	"fmt"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OfferRepository struct {
	db *pgxpool.Pool
}

func NewOfferRepository(db *pgxpool.Pool) *OfferRepository {
	return &OfferRepository{db: db}
}

func (repo *OfferRepository) Create(ctx context.Context, offer *models.Offer) error {
	query := `
		INSERT INTO offers (product_id, name, url)
		VALUES ($1, $2, $3)
		RETURNING id, product_id, name, url, created_at, updated_at
	`
	return repo.db.QueryRow(ctx, query,
		offer.ProductID, offer.Name, offer.URL,
	).Scan(&offer.ID, &offer.ProductID, &offer.Name, &offer.URL, &offer.CreatedAt, &offer.UpdatedAt)
}

func (repo *OfferRepository) GetByID(ctx context.Context, id int) (*models.Offer, error) {
	query := `
		SELECT id, product_id, name, url, created_at, updated_at
		FROM offers
		WHERE id = $1
	`
	offer := &models.Offer{}
	err := repo.db.QueryRow(ctx, query, id).Scan(
		&offer.ID, &offer.ProductID, &offer.Name, &offer.URL, &offer.CreatedAt, &offer.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return offer, nil
}

func (repo *OfferRepository) GetByProductID(ctx context.Context, productID int) ([]models.Offer, error) {
	query := `
		SELECT id, product_id, name, url, created_at, updated_at
		FROM offers
		WHERE product_id = $1
		ORDER BY created_at DESC
	`
	rows, err := repo.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []models.Offer
	for rows.Next() {
		var offer models.Offer
		err := rows.Scan(&offer.ID, &offer.ProductID, &offer.Name, &offer.URL, &offer.CreatedAt, &offer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func (repo *OfferRepository) GetActiveOffers(ctx context.Context) ([]models.Offer, error) {
	query := `
		SELECT o.id, o.product_id, o.name, o.url, o.created_at, o.updated_at
		FROM offers o
		JOIN products p ON p.id = o.product_id
		WHERE p.active = true
		ORDER BY o.created_at DESC
	`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []models.Offer
	for rows.Next() {
		var offer models.Offer
		err := rows.Scan(&offer.ID, &offer.ProductID, &offer.Name, &offer.URL, &offer.CreatedAt, &offer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	return offers, nil
}

func (repo *OfferRepository) Update(ctx context.Context, offer *models.Offer) (*models.Offer, error) {
	query := `
		UPDATE offers
		SET name = $1, url = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, product_id, name, url, created_at, updated_at
	`
	updated := &models.Offer{}
	err := repo.db.QueryRow(ctx, query,
		offer.Name, offer.URL, offer.ID,
	).Scan(&updated.ID, &updated.ProductID, &updated.Name, &updated.URL, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (repo *OfferRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM offers WHERE id = $1`
	result, err := repo.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("offer not found")
	}
	return nil
}

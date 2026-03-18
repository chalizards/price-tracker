package repository

import (
	"context"
	"fmt"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StoreRepository struct {
	db *pgxpool.Pool
}

func NewStoreRepository(db *pgxpool.Pool) *StoreRepository {
	return &StoreRepository{db: db}
}

func (repo *StoreRepository) Create(ctx context.Context, store *models.Store) error {
	query := `
		INSERT INTO stores (product_id, name, url)
		VALUES ($1, $2, $3)
		RETURNING id, product_id, name, url, created_at, updated_at
	`
	return repo.db.QueryRow(ctx, query,
		store.ProductID, store.Name, store.URL,
	).Scan(&store.ID, &store.ProductID, &store.Name, &store.URL, &store.CreatedAt, &store.UpdatedAt)
}

func (repo *StoreRepository) GetByID(ctx context.Context, id int) (*models.Store, error) {
	query := `
		SELECT id, product_id, name, url, created_at, updated_at
		FROM stores
		WHERE id = $1
	`
	store := &models.Store{}
	err := repo.db.QueryRow(ctx, query, id).Scan(
		&store.ID, &store.ProductID, &store.Name, &store.URL, &store.CreatedAt, &store.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func (repo *StoreRepository) GetByProductID(ctx context.Context, productID int) ([]models.Store, error) {
	query := `
		SELECT id, product_id, name, url, created_at, updated_at
		FROM stores
		WHERE product_id = $1
		ORDER BY created_at DESC
	`
	rows, err := repo.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var store models.Store
		err := rows.Scan(&store.ID, &store.ProductID, &store.Name, &store.URL, &store.CreatedAt, &store.UpdatedAt)
		if err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, nil
}

func (repo *StoreRepository) GetActiveStores(ctx context.Context) ([]models.Store, error) {
	query := `
		SELECT s.id, s.product_id, s.name, s.url, s.created_at, s.updated_at
		FROM stores s
		JOIN products p ON p.id = s.product_id
		WHERE p.active = true
		ORDER BY s.created_at DESC
	`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var store models.Store
		err := rows.Scan(&store.ID, &store.ProductID, &store.Name, &store.URL, &store.CreatedAt, &store.UpdatedAt)
		if err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, nil
}

func (repo *StoreRepository) Update(ctx context.Context, store *models.Store) (*models.Store, error) {
	query := `
		UPDATE stores
		SET name = $1, url = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, product_id, name, url, created_at, updated_at
	`
	updated := &models.Store{}
	err := repo.db.QueryRow(ctx, query,
		store.Name, store.URL, store.ID,
	).Scan(&updated.ID, &updated.ProductID, &updated.Name, &updated.URL, &updated.CreatedAt, &updated.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (repo *StoreRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM stores WHERE id = $1`
	result, err := repo.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("store not found")
	}
	return nil
}

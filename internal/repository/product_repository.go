package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/chalizards/price-tracker/internal/models"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (repo *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (name, slug, url, store, target_price, active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, slug, url, store, target_price, active, created_at, updated_at
	`
	return repo.db.QueryRow(ctx, query,
		product.Name, product.Slug, product.URL, product.Store, product.TargetPrice, product.Active,
	).Scan(&product.ID, &product.Name, &product.Slug, &product.URL, &product.Store, &product.TargetPrice, &product.Active, &product.CreatedAt, &product.UpdatedAt)
}

func (repo *ProductRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
	query := `
		SELECT id, name, slug, url, store, target_price, active, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	product := &models.Product{}
	err := repo.db.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Slug, &product.URL, &product.Store, &product.TargetPrice,
		&product.Active, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (repo *ProductRepository) GetBySlug(ctx context.Context, slug string) (*models.Product, error) {
	query := `
		SELECT id, name, slug, url, store, target_price, active, created_at, updated_at
		FROM products
		WHERE slug = $1
	`
	product := &models.Product{}
	err := repo.db.QueryRow(ctx, query, slug).Scan(
		&product.ID, &product.Name, &product.Slug, &product.URL, &product.Store, &product.TargetPrice,
		&product.Active, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (repo *ProductRepository) GetByName(ctx context.Context, name string) ([]models.Product, error) {
	query := `
		SELECT id, name, slug, url, store, target_price, active, created_at, updated_at
		FROM products
		WHERE name ILIKE '%' || $1 || '%'
	`
	rows, err := repo.db.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Slug, &product.URL, &product.Store, &product.TargetPrice,
			&product.Active, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (repo *ProductRepository) GetAll(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT id, name, slug, url, store, target_price, active, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var products []models.Product

	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Slug, &product.URL, &product.Store, &product.TargetPrice,
			&product.Active, &product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}
	
	return products, nil
}

func (repo *ProductRepository) GetActive(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT id, name, url, store, target_price, active, created_at, updated_at
		FROM products
		WHERE active = true
		ORDER BY created_at DESC
	`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product

	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.URL, &product.Store, &product.TargetPrice,
			&product.Active, &product.CreatedAt, &product.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}
	return products, nil
}

func (repo *ProductRepository) Update(ctx context.Context, product *models.Product) (*models.Product, error) {
	query := `
		UPDATE products
		SET name = $1, target_price = $2, active = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, slug, url, store, target_price, active, created_at, updated_at
	`
	updatedProduct := &models.Product{}

	err := repo.db.QueryRow(ctx, query,
		product.Name, product.TargetPrice, product.Active, product.ID,
	).Scan(
		&updatedProduct.ID, &updatedProduct.Name, &updatedProduct.Slug, &updatedProduct.URL, &updatedProduct.Store, &updatedProduct.TargetPrice,
		&updatedProduct.Active, &updatedProduct.CreatedAt, &updatedProduct.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return updatedProduct, nil
}

func (repo *ProductRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := repo.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}
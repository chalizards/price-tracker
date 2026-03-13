package repository

import (
	"context"

	"github.com/chalizards/price-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT id, google_id, email, name, picture, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &models.User{}
	err := repo.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Picture,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) FindByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	query := `
		SELECT id, google_id, email, name, picture, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`
	user := &models.User{}
	err := repo.db.QueryRow(ctx, query, googleID).Scan(
		&user.ID, &user.GoogleID, &user.Email, &user.Name, &user.Picture,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) Upsert(ctx context.Context, user *models.User) (*models.User, error) {
	query := `
		INSERT INTO users (google_id, email, name, picture)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			updated_at = NOW()
		RETURNING id, google_id, email, name, picture, created_at, updated_at
	`
	result := &models.User{}
	err := repo.db.QueryRow(ctx, query,
		user.GoogleID, user.Email, user.Name, user.Picture,
	).Scan(
		&result.ID, &result.GoogleID, &result.Email, &result.Name, &result.Picture,
		&result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

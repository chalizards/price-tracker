package repository

import (
	"context"
	"fmt"
	"github.com/chalizards/price-tracker/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (repo *NotificationRepository) Create(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications (product_id, price_id, type, title, message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, product_id, price_id, type, title, message, read, created_at
	`
	return repo.db.QueryRow(ctx, query,
		notification.ProductID, notification.PriceID, notification.Type, notification.Title, notification.Message,
	).Scan(&notification.ID, &notification.ProductID, &notification.PriceID, &notification.Type, &notification.Title, &notification.Message, &notification.Read, &notification.CreatedAt)
}

func (repo *NotificationRepository) GetByProductID(ctx context.Context, productID int) ([]models.Notification, error) {
	query := `
		SELECT id, product_id, price_id, type, title, message, read, created_at
		FROM notifications
		WHERE product_id = $1
		ORDER BY created_at DESC
	`
	rows, err := repo.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(&notification.ID, &notification.ProductID, &notification.PriceID, &notification.Type, &notification.Title, &notification.Message, &notification.Read, &notification.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func (repo *NotificationRepository) GetUnread(ctx context.Context) ([]models.Notification, error) {
	query := `
		SELECT id, product_id, price_id, type, title, message, read, created_at
		FROM notifications
		WHERE read = false
		ORDER BY created_at DESC
	`
	rows, err := repo.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		err := rows.Scan(&notification.ID, &notification.ProductID, &notification.PriceID, &notification.Type, &notification.Title, &notification.Message, &notification.Read, &notification.CreatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, nil
}

func (repo *NotificationRepository) MarkAsRead(ctx context.Context, id int) error {
	query := `UPDATE notifications SET read = true WHERE id = $1`
	result, err := repo.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("notification not found")
	}
	return nil
}

func (repo *NotificationRepository) MarkAllAsRead(ctx context.Context) error {
	query := `UPDATE notifications SET read = true WHERE read = false`
	_, err := repo.db.Exec(ctx, query)
	return err
}

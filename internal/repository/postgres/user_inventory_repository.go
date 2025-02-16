package postgres

import (
	"avito-shop/internal/domain/models"
	"context"
	"database/sql"
)

type UserInventoryRepository struct {
	db *sql.DB
}

func NewUserInventoryRepository(db *sql.DB) *UserInventoryRepository {
	return &UserInventoryRepository{db: db}
}

func (r *UserInventoryRepository) AddItem(ctx context.Context, userID int64, merchandiseID int64) error {
	query := `
		INSERT INTO user_inventory (user_id, merchandise_id)
		VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, query, userID, merchandiseID)
	return err
}

func (r *UserInventoryRepository) GetUserItems(ctx context.Context, userID int64) ([]*models.InventoryItem, error) {
	query := `
		SELECT m.name, COUNT(ui.id)
		FROM user_inventory ui
		JOIN merchandise m ON ui.merchandise_id = m.id
		WHERE ui.user_id = $1
		GROUP BY m.name`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.InventoryItem
	for rows.Next() {
		item := &models.InventoryItem{}
		if err := rows.Scan(&item.Type, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

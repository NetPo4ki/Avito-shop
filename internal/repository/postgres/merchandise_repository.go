package postgres

import (
	"avito-shop/internal/domain/models"
	"context"
	"database/sql"
)

type MerchandiseRepository struct {
	db *sql.DB
}

func NewMerchandiseRepository(db *sql.DB) *MerchandiseRepository {
	return &MerchandiseRepository{db: db}
}

func (r *MerchandiseRepository) GetByName(ctx context.Context, name string) (*models.Merchandise, error) {
	merchandise := &models.Merchandise{}
	query := `
		SELECT id, name, price
		FROM merchandise
		WHERE name = $1`

	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&merchandise.ID,
		&merchandise.Name,
		&merchandise.Price,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return merchandise, nil
}

func (r *MerchandiseRepository) GetAll(ctx context.Context) ([]*models.Merchandise, error) {
	query := `
		SELECT id, name, price
		FROM merchandise
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Merchandise
	for rows.Next() {
		item := &models.Merchandise{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

package postgres

import (
	"avito-shop/internal/domain/models"
	"context"
	"database/sql"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	query := `
		INSERT INTO coin_transactions (from_user_id, to_user_id, amount, transaction_type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		transaction.FromUserID,
		transaction.ToUserID,
		transaction.Amount,
		transaction.TransactionType,
	).Scan(&transaction.ID, &transaction.CreatedAt)
}

func (r *TransactionRepository) GetUserTransactions(ctx context.Context, userID int64) ([]*models.Transaction, error) {
	query := `
		SELECT id, from_user_id, to_user_id, amount, transaction_type, created_at
		FROM coin_transactions
		WHERE from_user_id = $1 OR to_user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.Transaction
	for rows.Next() {
		transaction := &models.Transaction{}
		if err := rows.Scan(
			&transaction.ID,
			&transaction.FromUserID,
			&transaction.ToUserID,
			&transaction.Amount,
			&transaction.TransactionType,
			&transaction.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

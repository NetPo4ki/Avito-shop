package models

import "time"

const (
	TransactionTypeTransfer = "TRANSFER"
	TransactionTypePurchase = "PURCHASE"
)

type Transaction struct {
	ID              int64     `json:"id"`
	FromUserID      int64     `json:"from_user_id"`
	ToUserID        *int64    `json:"to_user_id"`
	Amount          int       `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
}

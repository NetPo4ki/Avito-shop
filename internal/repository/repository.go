package repository

import (
	"avito-shop/internal/domain/models"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateCoins(ctx context.Context, userID int64, amount int) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
}

type MerchandiseRepository interface {
	GetByName(ctx context.Context, name string) (*models.Merchandise, error)
	GetAll(ctx context.Context) ([]*models.Merchandise, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) error
	GetUserTransactions(ctx context.Context, userID int64) ([]*models.Transaction, error)
}

type UserInventoryRepository interface {
	AddItem(ctx context.Context, userID int64, merchandiseID int64) error
	GetUserItems(ctx context.Context, userID int64) ([]*models.InventoryItem, error)
}

type Repositories struct {
	Users        UserRepository
	Merchandise  MerchandiseRepository
	Transactions TransactionRepository
	Inventory    UserInventoryRepository
}

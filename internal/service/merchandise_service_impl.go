package service

import (
	"avito-shop/internal/domain/models"
	"avito-shop/internal/repository"
	"context"
	"fmt"
)

type merchandiseService struct {
	users        repository.UserRepository
	merchandise  repository.MerchandiseRepository
	inventory    repository.UserInventoryRepository
	transactions repository.TransactionRepository
}

func NewMerchandiseService(
	users repository.UserRepository,
	merchandise repository.MerchandiseRepository,
	inventory repository.UserInventoryRepository,
	transactions repository.TransactionRepository,
) MerchandiseService {
	return &merchandiseService{
		users:        users,
		merchandise:  merchandise,
		inventory:    inventory,
		transactions: transactions,
	}
}

func (s *merchandiseService) GetAll(ctx context.Context) ([]*models.Merchandise, error) {
	items, err := s.merchandise.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting merchandise: %w", err)
	}
	return items, nil
}

func (s *merchandiseService) BuyItem(ctx context.Context, userID int64, itemName string) error {
	if userID == 0 {
		return fmt.Errorf("invalid user ID")
	}

	if itemName == "" {
		return fmt.Errorf("item name is required")
	}

	item, err := s.merchandise.GetByName(ctx, itemName)
	if err != nil {
		return fmt.Errorf("error getting item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found")
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.Coins < item.Price {
		return fmt.Errorf("insufficient funds: have %d, need %d", user.Coins, item.Price)
	}

	transaction := &models.Transaction{
		FromUserID:      userID,
		ToUserID:        nil,
		Amount:          item.Price,
		TransactionType: models.TransactionTypePurchase,
	}

	if err := s.users.UpdateCoins(ctx, userID, -item.Price); err != nil {
		return fmt.Errorf("error updating user balance: %w", err)
	}

	if err := s.transactions.Create(ctx, transaction); err != nil {
		_ = s.users.UpdateCoins(ctx, userID, item.Price)
		return fmt.Errorf("error recording transaction: %w", err)
	}

	if err := s.inventory.AddItem(ctx, userID, item.ID); err != nil {
		_ = s.users.UpdateCoins(ctx, userID, item.Price)
		return fmt.Errorf("error updating inventory: %w", err)
	}

	return nil
}

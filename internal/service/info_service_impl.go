package service

import (
	"avito-shop/internal/domain/models"
	"avito-shop/internal/repository"
	"context"
	"fmt"
)

type infoService struct {
	users        repository.UserRepository
	merchandise  repository.MerchandiseRepository
	transactions repository.TransactionRepository
	inventory    repository.UserInventoryRepository
}

func NewInfoService(users repository.UserRepository, merchandise repository.MerchandiseRepository, transactions repository.TransactionRepository, inventory repository.UserInventoryRepository) InfoService {
	return &infoService{
		users:        users,
		merchandise:  merchandise,
		transactions: transactions,
		inventory:    inventory,
	}
}

func (s *infoService) GetUserInfo(ctx context.Context, userID int64) (*models.InfoResponse, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	transactions, err := s.transactions.GetUserTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting transactions: %w", err)
	}

	var received []models.CoinReceived
	var sent []models.CoinSent

	for _, t := range transactions {
		if t.ToUserID != nil && *t.ToUserID == userID {
			fromUser, err := s.users.GetByID(ctx, t.FromUserID)
			if err != nil {
				return nil, fmt.Errorf("error getting sender info: %w", err)
			}
			senderName := "Unknown"
			if fromUser != nil {
				senderName = fromUser.Username
			}
			received = append(received, models.CoinReceived{
				FromUser: senderName,
				Amount:   t.Amount,
			})
		} else {
			var toUser string
			if t.TransactionType == models.TransactionTypePurchase {
				toUser = "SHOP"
			} else {
				recipient, err := s.users.GetByID(ctx, *t.ToUserID)
				if err != nil {
					return nil, fmt.Errorf("error getting recipient info: %w", err)
				}
				toUser = "Unknown"
				if recipient != nil {
					toUser = recipient.Username
				}
			}
			sent = append(sent, models.CoinSent{
				ToUser: toUser,
				Amount: t.Amount,
			})
		}
	}

	inventory, err := s.inventory.GetUserItems(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting inventory: %w", err)
	}

	return &models.InfoResponse{
		Coins: user.Coins,
		CoinHistory: models.CoinTransactionHistory{
			Received: received,
			Sent:     sent,
		},
		Inventory: inventory,
	}, nil
}

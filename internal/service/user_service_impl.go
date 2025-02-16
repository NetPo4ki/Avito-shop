package service

import (
	"avito-shop/internal/domain/models"
	"avito-shop/internal/repository"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type userServiceImpl struct {
	users        repository.UserRepository
	transactions repository.TransactionRepository
	tokenSecret  string
}

func NewUserService(users repository.UserRepository, transactions repository.TransactionRepository, tokenSecret string) UserService {
	return &userServiceImpl{
		users:        users,
		transactions: transactions,
		tokenSecret:  tokenSecret,
	}
}

func (s *userServiceImpl) Register(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}

	existingUser, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("error checking existing user: %w", err)
	}
	if existingUser != nil {
		return fmt.Errorf("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Coins:        1000,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func (s *userServiceImpl) Login(ctx context.Context, username, password string) (string, error) {
	log.Printf("Attempting to log in user: %s", username)
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required")
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		log.Printf("User not found: %s", username)
		return "", fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return tokenString, nil
}

func (s *userServiceImpl) TransferCoins(ctx context.Context, fromUserID int64, toUsername string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if fromUserID == 0 {
		return fmt.Errorf("invalid sender ID")
	}

	fromUser, err := s.users.GetByID(ctx, fromUserID)
	if err != nil {
		return fmt.Errorf("error getting sender: %w", err)
	}
	if fromUser == nil {
		return fmt.Errorf("sender not found")
	}

	if fromUser.Coins < amount {
		return fmt.Errorf("insufficient funds: have %d, need %d", fromUser.Coins, amount)
	}

	toUser, err := s.users.GetByUsername(ctx, toUsername)
	if err != nil {
		return fmt.Errorf("error getting recipient: %w", err)
	}
	if toUser == nil {
		return fmt.Errorf("recipient not found")
	}

	transaction := &models.Transaction{
		FromUserID:      fromUserID,
		ToUserID:        &toUser.ID,
		Amount:          amount,
		TransactionType: models.TransactionTypeTransfer,
	}

	if err := s.users.UpdateCoins(ctx, fromUserID, -amount); err != nil {
		return fmt.Errorf("error updating sender balance: %w", err)
	}

	if err := s.users.UpdateCoins(ctx, toUser.ID, amount); err != nil {
		_ = s.users.UpdateCoins(ctx, fromUserID, amount)
		return fmt.Errorf("error updating recipient balance: %w", err)
	}

	if err := s.transactions.Create(ctx, transaction); err != nil {
		_ = s.users.UpdateCoins(ctx, fromUserID, amount)
		_ = s.users.UpdateCoins(ctx, toUser.ID, -amount)
		return fmt.Errorf("error recording transaction: %w", err)
	}

	return nil
}

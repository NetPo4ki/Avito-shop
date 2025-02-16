package service

import (
	"avito-shop/internal/repository/postgres"
	"avito-shop/internal/test"
	"context"
	"testing"
)

func TestMerchandiseService_BuyItem(t *testing.T) {
	db, cleanup := test.SetupTestDB(t)
	defer cleanup()

	userRepo := postgres.NewUserRepository(db)
	merchRepo := postgres.NewMerchandiseRepository(db)
	invRepo := postgres.NewUserInventoryRepository(db)
	transRepo := postgres.NewTransactionRepository(db)

	service := NewMerchandiseService(
		userRepo,
		merchRepo,
		invRepo,
		transRepo,
	)

	ctx := context.Background()
	testUser := "testuser"
	testPass := "testpass"

	userService := NewUserService(userRepo, transRepo, "test-secret")
	err := userService.Register(ctx, testUser, testPass)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	user, err := userRepo.GetByUsername(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to get test user: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO merchandise (name, price) 
		VALUES ('test-item', 100)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test merchandise: %v", err)
	}

	tests := []struct {
		name     string
		userID   int64
		itemName string
		wantErr  bool
	}{
		{
			name:     "Valid purchase",
			userID:   user.ID,
			itemName: "test-item",
			wantErr:  false,
		},
		{
			name:     "Invalid item",
			userID:   user.ID,
			itemName: "non-existent-item",
			wantErr:  true,
		},
		{
			name:     "Invalid user",
			userID:   0,
			itemName: "test-item",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.BuyItem(ctx, tt.userID, tt.itemName)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuyItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				items, err := invRepo.GetUserItems(ctx, tt.userID)
				if err != nil {
					t.Errorf("Failed to get user inventory: %v", err)
				}
				if len(items) == 0 {
					t.Error("No items in inventory after purchase")
				}

				transactions, err := transRepo.GetUserTransactions(ctx, tt.userID)
				if err != nil {
					t.Errorf("Failed to get user transactions: %v", err)
				}
				if len(transactions) == 0 {
					t.Error("No transaction record after purchase")
				}
			}
		})
	}
}

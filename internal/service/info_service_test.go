package service

import (
	"avito-shop/internal/domain/models"
	"avito-shop/internal/repository"
	"avito-shop/internal/repository/postgres"
	"avito-shop/internal/test"
	"context"
	"database/sql"
	"testing"
)

const (
	testPassword = "testpass"
	testUser     = "testuser"
)

type testSetup struct {
	db           *sql.DB
	cleanup      func()
	infoService  InfoService
	userService  UserService
	merchService MerchandiseService
	userRepo     repository.UserRepository
}

func TestInfoService_GetUserInfo(t *testing.T) {
	ts := setupInfoServiceTest(t)
	defer ts.cleanup()

	user := createTestUser(t, ts)
	runInfoServiceTests(t, ts, user)
}

func setupInfoServiceTest(t *testing.T) *testSetup {
	db, cleanup := test.SetupTestDB(t)
	userRepo := postgres.NewUserRepository(db)
	merchRepo := postgres.NewMerchandiseRepository(db)
	invRepo := postgres.NewUserInventoryRepository(db)
	transRepo := postgres.NewTransactionRepository(db)

	infoService := NewInfoService(userRepo, merchRepo, transRepo, invRepo)
	userService := NewUserService(userRepo, transRepo, "test-secret")
	merchService := NewMerchandiseService(userRepo, merchRepo, invRepo, transRepo)

	return &testSetup{
		db:           db,
		cleanup:      cleanup,
		infoService:  infoService,
		userService:  userService,
		merchService: merchService,
		userRepo:     userRepo,
	}
}

func createTestUser(t *testing.T, ts *testSetup) *models.User {
	ctx := context.Background()

	testRecipient := "recipient"
	password := testPassword

	err := ts.userService.Register(ctx, testUser, password)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	err = ts.userService.Register(ctx, testRecipient, password)
	if err != nil {
		t.Fatalf("Failed to create recipient: %v", err)
	}

	user, err := ts.userRepo.GetByUsername(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to get test user: %v", err)
	}

	_, err = ts.db.Exec(`
		INSERT INTO merchandise (name, price) 
		VALUES ('test-item', 500)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test merchandise: %v", err)
	}

	err = ts.userService.TransferCoins(ctx, user.ID, testRecipient, 200)
	if err != nil {
		t.Fatalf("Failed to transfer coins: %v", err)
	}

	err = ts.merchService.BuyItem(ctx, user.ID, "test-item")
	if err != nil {
		t.Fatalf("Failed to buy item: %v", err)
	}

	return user
}

func runInfoServiceTests(t *testing.T, ts *testSetup, user *models.User) {
	tests := []struct {
		name    string
		userID  int64
		want    *models.InfoResponse
		wantErr bool
	}{
		{
			name:   "Valid user info",
			userID: user.ID,
			want: &models.InfoResponse{
				Coins: 300,
				Inventory: []*models.InventoryItem{
					{
						Type:     "test-item",
						Quantity: 1,
					},
				},
				CoinHistory: models.CoinTransactionHistory{
					Sent: []models.CoinSent{
						{
							ToUser: "recipient",
							Amount: 200,
						},
						{
							ToUser: "SHOP",
							Amount: 500,
						},
					},
					Received: []models.CoinReceived{},
				},
			},
			wantErr: false,
		},
		{
			name:    "Non-existent user",
			userID:  0,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ts.infoService.GetUserInfo(context.Background(), tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Coins != tt.want.Coins {
					t.Errorf("Coins = %v, want %v", got.Coins, tt.want.Coins)
				}

				if len(got.Inventory) != len(tt.want.Inventory) {
					t.Errorf("Inventory length = %v, want %v", len(got.Inventory), len(tt.want.Inventory))
				} else {
					for i, item := range got.Inventory {
						if item.Type != tt.want.Inventory[i].Type {
							t.Errorf("Inventory[%d].Type = %v, want %v", i, item.Type, tt.want.Inventory[i].Type)
						}
						if item.Quantity != tt.want.Inventory[i].Quantity {
							t.Errorf("Inventory[%d].Quantity = %v, want %v", i, item.Quantity, tt.want.Inventory[i].Quantity)
						}
					}
				}

				if len(got.CoinHistory.Sent) != len(tt.want.CoinHistory.Sent) {
					t.Errorf("Sent transactions length = %v, want %v", len(got.CoinHistory.Sent), len(tt.want.CoinHistory.Sent))
				}
				if len(got.CoinHistory.Received) != len(tt.want.CoinHistory.Received) {
					t.Errorf("Received transactions length = %v, want %v", len(got.CoinHistory.Received), len(tt.want.CoinHistory.Received))
				}
			}
		})
	}
}

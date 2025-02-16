package service

import (
	"avito-shop/internal/repository/postgres"
	"avito-shop/internal/test"
	"context"
	"testing"
)

func TestUserService_Register(t *testing.T) {
	db, cleanup := test.SetupTestDB(t)
	defer cleanup()

	userRepo := postgres.NewUserRepository(db)
	transRepo := postgres.NewTransactionRepository(db)
	service := NewUserService(userRepo, transRepo, "test-secret")

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{
			name:     "Valid registration",
			username: "testuser",
			password: "testpass",
			wantErr:  false,
		},
		{
			name:     "Empty username",
			username: "",
			password: "testpass",
			wantErr:  true,
		},
		{
			name:     "Empty password",
			username: "testuser",
			password: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.ClearTestDB(t, db)

			err := service.Register(context.Background(), tt.username, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				user, err := userRepo.GetByUsername(context.Background(), tt.username)
				if err != nil {
					t.Errorf("Failed to get user after registration: %v", err)
				}
				if user == nil {
					t.Error("User not found after registration")
					return
				}
				if user.Coins != 1000 {
					t.Errorf("User initial coins = %d, want %d", user.Coins, 1000)
				}
			}
		})
	}
}

func TestUserService_TransferCoins(t *testing.T) {
	db, cleanup := test.SetupTestDB(t)
	defer cleanup()

	userRepo := postgres.NewUserRepository(db)
	transRepo := postgres.NewTransactionRepository(db)
	service := NewUserService(userRepo, transRepo, "test-secret")

	ctx := context.Background()

	sender := "sender"
	recipient := "recipient"
	password := "testpass"

	err := service.Register(ctx, sender, password)
	if err != nil {
		t.Fatalf("Failed to create sender: %v", err)
	}

	senderUser, err := userRepo.GetByUsername(ctx, sender)
	if err != nil || senderUser == nil {
		t.Fatal("Failed to get sender")
	}

	senderID := senderUser.ID

	err = service.Register(ctx, recipient, password)
	if err != nil {
		t.Fatalf("Failed to create recipient: %v", err)
	}

	tests := []struct {
		name    string
		fromID  int64
		toUser  string
		amount  int
		wantErr bool
	}{
		{
			name:    "Valid transfer",
			fromID:  senderID,
			toUser:  recipient,
			amount:  100,
			wantErr: false,
		},
		{
			name:    "Invalid amount",
			fromID:  senderID,
			toUser:  recipient,
			amount:  -100,
			wantErr: true,
		},
		{
			name:    "Insufficient funds",
			fromID:  senderID,
			toUser:  recipient,
			amount:  2000,
			wantErr: true,
		},
		{
			name:    "Non-existent recipient",
			fromID:  senderID,
			toUser:  "non-existent",
			amount:  100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.TransferCoins(ctx, tt.fromID, tt.toUser, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("TransferCoins() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				sender, err := userRepo.GetByID(ctx, tt.fromID)
				if err != nil {
					t.Errorf("Failed to get sender after transfer: %v", err)
				}
				if sender.Coins != 1000-tt.amount {
					t.Errorf("Sender coins = %d, want %d", sender.Coins, 1000-tt.amount)
				}

				recipient, err := userRepo.GetByUsername(ctx, tt.toUser)
				if err != nil {
					t.Errorf("Failed to get recipient after transfer: %v", err)
				}
				if recipient.Coins != 1000+tt.amount {
					t.Errorf("Recipient coins = %d, want %d", recipient.Coins, 1000+tt.amount)
				}

				transactions, err := transRepo.GetUserTransactions(ctx, tt.fromID)
				if err != nil {
					t.Errorf("Failed to get transactions: %v", err)
				}
				if len(transactions) == 0 {
					t.Error("No transaction record after transfer")
				}
			}
		})
	}
}

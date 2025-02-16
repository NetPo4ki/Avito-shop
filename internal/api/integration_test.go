package api

import (
	"avito-shop/internal/domain/models"
	"avito-shop/internal/repository"
	"avito-shop/internal/repository/postgres"
	"avito-shop/internal/service"
	"avito-shop/internal/test"
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testServer struct {
	db       *sql.DB
	handler  http.Handler
	cleanup  func()
	services *service.Services
}

func setupTestServer(t *testing.T) *testServer {
	db, cleanup := test.SetupTestDB(t)

	repos := &repository.Repositories{
		Users:        postgres.NewUserRepository(db),
		Merchandise:  postgres.NewMerchandiseRepository(db),
		Transactions: postgres.NewTransactionRepository(db),
		Inventory:    postgres.NewUserInventoryRepository(db),
	}

	services := service.NewServices(service.ServicesDeps{
		Repos:       repos,
		TokenSecret: "test-secret",
	})

	router := NewRouter(services)
	handler := router.Setup()

	return &testServer{
		db:       db,
		handler:  handler,
		cleanup:  cleanup,
		services: services,
	}
}

func (ts *testServer) executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	ts.handler.ServeHTTP(rr, req)
	return rr
}

func TestFullUserFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	_, err := ts.db.Exec(`
		INSERT INTO merchandise (name, price) 
		VALUES ('test-item', 100)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test merchandise: %v", err)
	}

	recipientBody := map[string]string{
		"username": "recipient",
		"password": "testpass",
	}
	body, _ := json.Marshal(recipientBody)
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	resp := ts.executeRequest(req)
	if resp.Code != http.StatusOK {
		t.Fatalf("Failed to register recipient: status code %d", resp.Code)
	}

	registerBody := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	body, _ = json.Marshal(registerBody)
	req = httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	resp = ts.executeRequest(req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	if err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	req = httptest.NewRequest("GET", "/api/buy/test-item", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp = ts.executeRequest(req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	transferBody := map[string]interface{}{
		"toUser": "recipient",
		"amount": 100,
	}
	body, _ = json.Marshal(transferBody)
	req = httptest.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp = ts.executeRequest(req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	req = httptest.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	resp = ts.executeRequest(req)

	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, resp.Code)
	}

	var infoResp struct {
		Coins       int                           `json:"coins"`
		Inventory   []*models.InventoryItem       `json:"inventory"`
		CoinHistory models.CoinTransactionHistory `json:"coinHistory"`
	}
	err = json.NewDecoder(resp.Body).Decode(&infoResp)
	if err != nil {
		t.Fatalf("Failed to decode info response: %v", err)
	}

	expectedCoins := 800
	if infoResp.Coins != expectedCoins {
		t.Errorf("Expected %d coins, got %d", expectedCoins, infoResp.Coins)
	}

	if len(infoResp.Inventory) == 0 {
		t.Error("Expected non-empty inventory")
	}
	if len(infoResp.CoinHistory.Sent) == 0 {
		t.Error("Expected non-empty sent transactions")
	}
}

func TestAuthFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	tests := []struct {
		name         string
		username     string
		password     string
		expectedCode int
	}{
		{
			name:         "Valid registration",
			username:     "newuser",
			password:     "password123",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Duplicate registration",
			username:     "newuser",
			password:     "password123",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Empty username",
			username:     "",
			password:     "password123",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"username": tt.username,
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(jsonBody))
			resp := ts.executeRequest(req)

			if resp.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.Code)
			}
		})
	}
}

func TestInfoFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	registerBody := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	resp := ts.executeRequest(req)

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	tests := []struct {
		name         string
		setupFunc    func()
		token        string
		expectedCode int
	}{
		{
			name:         "Valid token",
			token:        loginResp.Token,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Invalid token",
			token:        "invalid-token",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "No token",
			token:        "",
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			req := httptest.NewRequest("GET", "/api/info", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			resp := ts.executeRequest(req)

			if resp.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.Code)
			}
		})
	}
}

func TestErrorScenarios(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	user := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/auth", bytes.NewBuffer(body))
	resp := ts.executeRequest(req)

	var loginResp struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	tests := []struct {
		name         string
		method       string
		path         string
		body         interface{}
		token        string
		expectedCode int
	}{
		{
			name:         "Buy non-existent item",
			method:       "GET",
			path:         "/api/buy/non-existent",
			token:        loginResp.Token,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "Transfer negative amount",
			method:       "POST",
			path:         "/api/sendCoin",
			body:         map[string]interface{}{"toUser": "other", "amount": -100},
			token:        loginResp.Token,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Transfer to non-existent user",
			method:       "POST",
			path:         "/api/sendCoin",
			body:         map[string]interface{}{"toUser": "nonexistent", "amount": 100},
			token:        loginResp.Token,
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				reqBody = bytes.NewBuffer(jsonBody)
			}

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp := ts.executeRequest(req)
			if resp.Code != tt.expectedCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedCode, resp.Code)
			}
		})
	}
}

package handlers

import (
	"avito-shop/internal/api/middleware"
	"avito-shop/internal/service"
	"encoding/json"
	"net/http"
)

type TransferHandler struct {
	userService service.UserService
}

func NewTransferHandler(userService service.UserService) *TransferHandler {
	return &TransferHandler{
		userService: userService,
	}
}

type transferRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func (h *TransferHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.userService.TransferCoins(r.Context(), userID, req.ToUser, req.Amount); err != nil {
		writeError(w, "Failed to transfer coins: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

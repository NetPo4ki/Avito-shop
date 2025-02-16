package handlers

import (
	"avito-shop/internal/api/middleware"
	"avito-shop/internal/service"
	"net/http"
	"strings"
)

type BuyHandler struct {
	merchandiseService service.MerchandiseService
}

func NewBuyHandler(merchandiseService service.MerchandiseService) *BuyHandler {
	return &BuyHandler{
		merchandiseService: merchandiseService,
	}
}

func (h *BuyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	itemName := strings.TrimPrefix(r.URL.Path, "/api/buy/")
	if itemName == "" {
		writeError(w, "Item name is required", http.StatusBadRequest)
		return
	}

	if err := h.merchandiseService.BuyItem(r.Context(), userID, itemName); err != nil {
		writeError(w, "Failed to buy item: "+err.Error(), http.StatusBadRequest)
		return
	}

	writeSuccess(w)
}

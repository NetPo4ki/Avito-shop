package handlers

import (
	"avito-shop/internal/api/middleware"
	"avito-shop/internal/service"
	"net/http"
)

type InfoHandler struct {
	infoService service.InfoService
}

func NewInfoHandler(infoService service.InfoService) *InfoHandler {
	return &InfoHandler{
		infoService: infoService,
	}
}

func (h *InfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	info, err := h.infoService.GetUserInfo(r.Context(), userID)
	if err != nil {
		writeError(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, info, http.StatusOK)
}

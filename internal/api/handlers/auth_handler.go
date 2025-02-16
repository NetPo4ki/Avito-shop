package handlers

import (
	"avito-shop/internal/service"
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	userService service.UserService
}

func NewAuthHandler(userService service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if err := h.userService.Register(r.Context(), req.Username, req.Password); err != nil {
			writeError(w, "Failed to authenticate: "+err.Error(), http.StatusUnauthorized)
			return
		}
		token, err = h.userService.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			writeError(w, "Failed to login after registration: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, authResponse{Token: token}, http.StatusOK)
}

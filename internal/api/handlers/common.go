package handlers

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Errors string `json:"errors"`
}

type successResponse struct {
	Status string `json:"status"`
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, errorResponse{Errors: message}, status)
}

func writeSuccess(w http.ResponseWriter) {
	writeJSON(w, successResponse{Status: "success"}, http.StatusOK)
}

package utils

import (
	"encoding/json"
	"net/http"

	"github.com/riteshkumar/internal-transfers/internal/models"
)

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func WriteError(w http.ResponseWriter, status int, errorMsg, details string) {
	response := models.ErrorResponse{
		Error:   errorMsg,
		Message: details,
	}
	WriteJSON(w, status, response)
}

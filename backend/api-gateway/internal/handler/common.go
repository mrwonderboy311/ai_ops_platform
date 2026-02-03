// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
		"requestId": generateRequestID(),
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      data,
		"requestId": generateRequestID(),
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

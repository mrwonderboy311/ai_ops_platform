// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"
)

// Health returns the health check response
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// API handles all API requests
func API(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Placeholder for API routes
	// Will be replaced with actual handlers in subsequent stories
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "NOT_IMPLEMENTED",
			"message": "API endpoint not yet implemented",
		},
		"requestId": generateRequestID(),
	})
}

// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

var (
	hostHandler *HostHandler
	scanHandler *ScanHandler
	agentHandler *AgentHandler
)

// RegisterHandlers registers the API handlers
func RegisterHandlers(hostH *HostHandler, scanH *ScanHandler, agentH *AgentHandler) {
	hostHandler = hostH
	scanHandler = scanH
	agentHandler = agentH
}

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

	path := r.URL.Path

	// Route to appropriate handler
	if strings.HasPrefix(path, "/api/v1/agent/report") {
		// Agent reporting endpoint
		if agentHandler != nil {
			agentHandler.ServeHTTP(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Agent service not available")
		}
		return
	}

	if strings.HasPrefix(path, "/api/v1/hosts/scan-tasks/") {
		// Scan task status query
		if scanHandler != nil {
			scanHandler.GetScanStatus(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Scan service not available")
		}
		return
	}

	if strings.HasPrefix(path, "/api/v1/hosts") {
		// Host management endpoints
		if hostHandler != nil {
			hostHandler.ServeHTTP(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Host service not available")
		}
		return
	}

	// Unknown endpoint
	respondWithError(w, http.StatusNotFound, "NOT_FOUND", "API endpoint not found")
}

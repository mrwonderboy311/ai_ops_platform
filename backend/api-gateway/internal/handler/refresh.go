// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/wangjialin/myops/api-gateway/internal/service"
	apperrors "github.com/wangjialin/myops/pkg/errors"
)

// RefreshTokenHandler handles token refresh
type RefreshTokenHandler struct {
	AuthService *service.AuthService
}

// NewRefreshTokenHandler creates a new RefreshTokenHandler
func NewRefreshTokenHandler(authService *service.AuthService) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		AuthService: authService,
	}
}

// ServeHTTP handles HTTP requests for token refresh
func (h *RefreshTokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req service.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.AuthService.RefreshToken(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			statusCode := http.StatusUnauthorized
			respondWithError(w, statusCode, appErr.Code, appErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      resp,
		"requestId": generateRequestID(),
	})
}

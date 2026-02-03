// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/wangjialin/myops/api-gateway/internal/service"
	apperrors "github.com/wangjialin/myops/pkg/errors"
)

// LoginHandler handles user login
type LoginHandler struct {
	AuthService *service.AuthService
}

// NewLoginHandler creates a new LoginHandler
func NewLoginHandler(authService *service.AuthService) *LoginHandler {
	return &LoginHandler{
		AuthService: authService,
	}
}

// ServeHTTP handles HTTP requests for user login
func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.AuthService.Login(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			statusCode := http.StatusBadRequest
			if appErr.Code == "INVALID_CREDENTIALS" {
				statusCode = http.StatusUnauthorized
			}
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

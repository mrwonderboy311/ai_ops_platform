// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/wangjialin/myops/api-gateway/internal/service"
	apperrors "github.com/wangjialin/myops/pkg/errors"
)

// RegisterHandler handles user registration
type RegisterHandler struct {
	AuthService *service.AuthService
}

// NewRegisterHandler creates a new RegisterHandler
func NewRegisterHandler(authService *service.AuthService) *RegisterHandler {
	return &RegisterHandler{
		AuthService: authService,
	}
}

// ServeHTTP handles HTTP requests for user registration
func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.AuthService.Register(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			respondWithError(w, http.StatusBadRequest, appErr.Code, appErr.Message)
		} else {
			respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":      resp,
		"requestId": generateRequestID(),
	})
}

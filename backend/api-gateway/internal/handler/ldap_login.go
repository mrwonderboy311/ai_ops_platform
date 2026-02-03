// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/wangjialin/myops/api-gateway/internal/service"
	apperrors "github.com/wangjialin/myops/pkg/errors"
)

// LDAPLoginHandler handles LDAP user login
type LDAPLoginHandler struct {
	AuthService *service.AuthService
}

// NewLDAPLoginHandler creates a new LDAPLoginHandler
func NewLDAPLoginHandler(authService *service.AuthService) *LDAPLoginHandler {
	return &LDAPLoginHandler{
		AuthService: authService,
	}
}

// ServeHTTP handles HTTP requests for LDAP login
func (h *LDAPLoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req service.LDAPLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	resp, err := h.AuthService.LDAPLogin(r.Context(), &req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			statusCode := http.StatusUnauthorized
			if appErr.Code == "LDAP_NOT_CONFIGURED" {
				statusCode = http.StatusServiceUnavailable
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

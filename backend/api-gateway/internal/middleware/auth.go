package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	authorizationHeader = "Authorization"
	bearerScheme        = "Bearer "
)

// Auth validates JWT tokens from Authorization header
// For now, this is a placeholder that will be fully implemented in Story 1.4
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and public endpoints
		if isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get(authorizationHeader)
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "未提供认证令牌")
			return
		}

		if !strings.HasPrefix(authHeader, bearerScheme) {
			respondWithError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "无效的认证格式")
			return
		}

		token := strings.TrimPrefix(authHeader, bearerScheme)

		// TODO: Validate JWT token (will be implemented in Story 1.4)
		// For now, just check if token is not empty
		if token == "" {
			respondWithError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "令牌验证失败")
			return
		}

		// Token is valid (placeholder), proceed with request
		next.ServeHTTP(w, r)
	})
}

// isPublicEndpoint checks if the endpoint is public (no auth required)
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/v1/auth/register",
		"/api/v1/auth/login",
		"/api/v1/auth/ldap-login",
	}

	for _, public := range publicPaths {
		if strings.HasPrefix(path, public) {
			return true
		}
	}
	return false
}

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

func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

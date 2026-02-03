// Package middleware provides HTTP middleware for audit logging
package middleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

// AuditMiddleware creates a middleware that logs all HTTP requests
func AuditMiddleware(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Capture response writer to get status code
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Read body for non-GET requests (for tracking changes)
			var bodyBytes []byte
			if r.Method != http.MethodGet && r.Method != http.MethodHead {
				bodyBytes, _ = io.ReadAll(r.Body)
				// Restore body for downstream handlers
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			// Skip logging for health checks and static assets
			if shouldSkipLogging(r) {
				return
			}

			// Get user info from context
			var userID uuid.UUID
			var username string
			if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
				if uid, ok := userIDVal.(string); ok {
					userID, _ = uuid.Parse(uid)
				}
			}
			if usernameVal := r.Context().Value("username"); usernameVal != nil {
				if uname, ok := usernameVal.(string); ok {
					username = uname
				}
			}

			// Determine resource type from path
			resource := getResourceFromPath(r.URL.Path)

			// Create audit log entry
			auditLog := &model.AuditLog{
				ID:         uuid.New(),
				UserID:     userID,
				Username:   username,
				Action:     getActionFromMethod(r.Method, rw.status),
				Resource:   resource,
				ResourceID: getResourceIDFromPath(r.URL.Path),
				Method:     r.Method,
				Path:       r.URL.Path,
				IPAddress:  getIPAddress(r),
				UserAgent:  r.UserAgent(),
				StatusCode: rw.status,
			}

			// Add error message if status code indicates failure
			if rw.status >= 400 {
				auditLog.ErrorMsg = extractErrorInfo(rw)
			}

			// Track changes for create/update operations
			if (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) && len(bodyBytes) > 0 {
				auditLog.NewValue = string(bodyBytes)
			}

			auditLog.CreatedAt = startTime

			// Save audit log asynchronously
			go func() {
				if err := db.Create(auditLog).Error; err != nil {
					log.Printf("Failed to save audit log: %v", err)
				}
			}()
		})
	}
}

// shouldSkipLogging determines if a request should be skipped from audit logging
func shouldSkipLogging(r *http.Request) bool {
	// Skip health checks
	if r.URL.Path == "/health" {
		return true
	}

	// Skip static assets
	if strings.HasPrefix(r.URL.Path, "/static/") || strings.HasPrefix(r.URL.Path, "/assets/") {
		return true
	}

	// Skip favicon
	if r.URL.Path == "/favicon.ico" {
		return true
	}

	return false
}

// getActionFromMethod determines the action type from HTTP method and status code
func getActionFromMethod(method string, statusCode int) string {
	switch method {
	case http.MethodGet:
		return "read"
	case http.MethodPost:
		return "create"
	case http.MethodPut, http.MethodPatch:
		return "update"
	case http.MethodDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// getResourceFromPath determines resource type from URL path
func getResourceFromPath(path string) string {
	// Simple path-based resource detection
	if contains(path, "/hosts/") {
		return "hosts"
	}
	if contains(path, "/clusters/") {
		return "clusters"
	}
	if contains(path, "/users/") {
		return "users"
	}
	if contains(path, "/batch-tasks") {
		return "batch-tasks"
	}
	if contains(path, "/alert-rules") {
		return "alert-rules"
	}
	if contains(path, "/alerts") {
		return "alerts"
	}
	return "unknown"
}

// getResourceIDFromPath extracts resource ID from URL path
func getResourceIDFromPath(path string) string {
	parts := splitPath(path)
	for i, part := range parts {
		if i == len(parts)-1 && part != "" {
			return part
		}
	}
	return ""
}

// getIPAddress extracts IP address from request
func getIPAddress(r *http.Request) string {
	// Check X-Forwarded-For header (for proxied requests)
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return forwardedFor
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}

	return "unknown"
}

// extractErrorInfo extracts error information from response
func extractErrorInfo(rw *responseWriter) string {
	// Try to read response body if it contains error details
	// This is a simplified version
	return "Request failed"
}

// splitPath splits a path into parts
func splitPath(path string) []string {
	path = trimLeadingSlash(path)
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func trimLeadingSlash(s string) string {
	return strings.TrimPrefix(s, "/")
}

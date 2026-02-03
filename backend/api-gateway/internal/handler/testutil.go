// Package handler provides testing utilities
package handler

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestContext creates a test Gin context with optional userID
func TestContext(userID *uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = &http.Request{
		Header: make(http.Header),
	}

	if userID != nil {
		c.Set("userID", *userID)
	}

	return c, w
}

// SetContextUserID sets a userID in the gin context
func SetContextUserID(c *gin.Context, userID uuid.UUID) {
	c.Set("userID", userID)
}

// CreateTestRequest creates a test HTTP request with headers
func CreateTestRequest(method, url string, body []byte) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	if body != nil {
		req.Body = &mockReadCloser{buf: body}
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// mockReadCloser is a mock io.ReadCloser for testing
type mockReadCloser struct {
	buf    []byte
	offset int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.offset >= len(m.buf) {
		return 0, context.Done()
	}
	n = copy(p, m.buf[m.offset:])
	m.offset += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	return nil
}

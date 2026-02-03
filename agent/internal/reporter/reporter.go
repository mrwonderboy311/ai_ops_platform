// Package reporter handles reporting host information to the server
package reporter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Reporter reports host information to the server
type Reporter struct {
	endpoint string
	token    string
	insecure bool
	client   *http.Client
}

// ReportResponse represents the server response
type ReportResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewReporter creates a new reporter
func NewReporter(endpoint, token string, insecure bool) *Reporter {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Skip TLS verification if insecure (for development)
	if insecure {
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
		client.Transport = transport
	}

	return &Reporter{
		endpoint: endpoint,
		token:    token,
		insecure: insecure,
		client:   client,
	}
}

// Report reports the host information to the server
func (r *Reporter) Report(hostInfo interface{}) error {
	// Marshal the host info to JSON
	data, err := json.Marshal(hostInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal host info: %w", err)
	}

	// Create the request
	reportURL := fmt.Sprintf("%s/api/v1/agent/report", r.endpoint)
	req, err := http.NewRequest("POST", reportURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))
	req.Header.Set("User-Agent", "MyOps-Agent/1.0")

	// Send the request
	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send report: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var reportResp ReportResponse
	if err := json.NewDecoder(resp.Body).Decode(&reportResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !reportResp.Success {
		return fmt.Errorf("server rejected report: %s", reportResp.Message)
	}

	return nil
}

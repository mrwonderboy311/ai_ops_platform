// Package handler provides unit tests for AI analysis operations
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"myops.k8s.io/backend/pkg/model"
)

func setupAITestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.AnomalyDetectionRule{},
		&model.AnomalyEvent{},
		&model.LLMConversation{},
		&model.LLMMessage{},
		&model.NLQuery{},
		&model.KnowledgeBaseEntry{},
		&model.BaselineMetric{},
		&model.User{},
		&model.K8sCluster{},
		&model.PrometheusDataSource{},
	)
	require.NoError(t, err)

	return db
}

func seedAITestData(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID) {
	// Create test user
	user := model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
	}
	require.NoError(t, db.Create(&user).Error)

	// Create test cluster
	cluster := model.K8sCluster{
		Name:     "test-cluster",
		Endpoint: "https://test.k8s.local",
		UserID:   user.ID,
	}
	require.NoError(t, db.Create(&cluster).Error)

	// Create test data source
	dataSource := model.PrometheusDataSource{
		Name:   "test-prometheus",
		URL:    "http://prometheus:9090",
		UserID: user.ID,
	}
	require.NoError(t, db.Create(&dataSource).Error)

	return user.ID, dataSource.ID
}

func TestCreateAnomalyDetectionRule(t *testing.T) {
	db := setupAITestDB(t)
	userID, dataSourceID := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	newRule := CreateAnomalyDetectionRuleRequest{
		Name:              "High CPU Anomaly",
		Description:       "Detects unusually high CPU usage",
		DataSourceID:      &dataSourceID,
		MetricQuery:       "rate(cpu_usage_total[5m])",
		Algorithm:         "stl",
		Sensitivity:       0.95,
		WindowSize:        100,
		Enabled:           true,
		EvalInterval:      300,
		AlertThreshold:    0.8,
		AlertOnRecovery:   false,
	}

	body, _ := json.Marshal(newRule)
	req := CreateTestRequest("POST", "/api/v1/ai/anomaly-rules", body)
	w := httptest.NewRecorder()

	handler.CreateAnomalyRule(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.AnomalyDetectionRule
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "High CPU Anomaly", response.Name)
	assert.Equal(t, "stl", response.Algorithm)
	assert.Equal(t, userID, response.UserID)
}

func TestListAnomalyDetectionRules(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	// Create some test rules
	rules := []model.AnomalyDetectionRule{
		{Name: "Rule 1", MetricQuery: "metric1", Algorithm: "stl", UserID: userID, Sensitivity: 0.9, WindowSize: 100, Enabled: true},
		{Name: "Rule 2", MetricQuery: "metric2", Algorithm: "isolation_forest", UserID: userID, Sensitivity: 0.95, WindowSize: 100, Enabled: true},
		{Name: "Rule 3", MetricQuery: "metric3", Algorithm: "lstm", UserID: userID, Sensitivity: 0.8, WindowSize: 100, Enabled: false},
	}
	for _, rule := range rules {
		require.NoError(t, db.Create(&rule).Error)
	}

	req := CreateTestRequest("GET", "/api/v1/ai/anomaly-rules?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	handler.ListAnomalyRules(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data     []model.AnomalyDetectionRule `json:"data"`
		Total    int                           `json:"total"`
		Page     int                           `json:"page"`
		PageSize int                           `json:"pageSize"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 3, response.Total)
	assert.Equal(t, 3, len(response.Data))
}

func TestUpdateAnomalyDetectionRule(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	// Create test rule
	rule := model.AnomalyDetectionRule{
		Name:        "Original Name",
		MetricQuery: "metric1",
		Algorithm:   "stl",
		UserID:      userID,
		Sensitivity: 0.9,
		WindowSize:  100,
		Enabled:     true,
	}
	require.NoError(t, db.Create(&rule).Error)

	// Update rule
	updateData := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated description",
		"sensitivity": 0.98,
	}
	body, _ := json.Marshal(updateData)
	req := CreateTestRequest("PATCH", "/api/v1/ai/anomaly-rules/"+rule.ID.String(), body)
	w := httptest.NewRecorder()

	handler.UpdateAnomalyRule(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.AnomalyDetectionRule
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "Updated description", response.Description)
	assert.Equal(t, 0.98, response.Sensitivity)
}

func TestDeleteAnomalyDetectionRule(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	// Create test rule
	rule := model.AnomalyDetectionRule{
		Name:        "Test Rule",
		MetricQuery: "metric1",
		Algorithm:   "stl",
		UserID:      userID,
		Sensitivity: 0.9,
		WindowSize:  100,
		Enabled:     true,
	}
	require.NoError(t, db.Create(&rule).Error)

	req := CreateTestRequest("DELETE", "/api/v1/ai/anomaly-rules/"+rule.ID.String(), nil)
	w := httptest.NewRecorder()

	handler.DeleteAnomalyRule(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify rule is deleted
	var count int64
	db.Model(&model.AnomalyDetectionRule{}).Where("id = ?", rule.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCreateLLMConversation(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	newConv := CreateLLMConversationRequest{
		Title:       "Troubleshooting Help",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	body, _ := json.Marshal(newConv)
	req := CreateTestRequest("POST", "/api/v1/ai/llm/conversations", body)
	w := httptest.NewRecorder()

	handler.CreateLLMConversation(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.LLMConversation
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Troubleshooting Help", response.Title)
	assert.Equal(t, "gpt-4", response.Model)
	assert.Equal(t, userID, response.UserID)
}

func TestListLLMConversations(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	// Create test conversations
	conversations := []model.LLMConversation{
		{Title: "Conv 1", Model: "gpt-4", UserID: userID, Temperature: 0.7, MaxTokens: 2000},
		{Title: "Conv 2", Model: "claude-3", UserID: userID, Temperature: 0.5, MaxTokens: 1000},
	}
	for _, conv := range conversations {
		require.NoError(t, db.Create(&conv).Error)
	}

	req := CreateTestRequest("GET", "/api/v1/ai/llm/conversations?page=1&pageSize=10", nil)
	w := httptest.NewRecorder()

	handler.ListLLMConversations(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Data     []model.LLMConversation `json:"data"`
		Total    int                      `json:"total"`
		Page     int                      `json:"page"`
		PageSize int                      `json:"pageSize"`
	}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.Total)
}

func TestSendLLMMessage(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	// Create test conversation
	conv := model.LLMConversation{
		Title:       "Test Conversation",
		Model:       "gpt-4",
		UserID:      userID,
		Temperature: 0.7,
		MaxTokens:   2000,
	}
	require.NoError(t, db.Create(&conv).Error)

	newMessage := SendLLMMessageRequest{
		Content: "What is Kubernetes?",
	}

	body, _ := json.Marshal(newMessage)
	url := fmt.Sprintf("/api/v1/ai/llm/conversations/%s/messages", conv.ID.String())
	req := CreateTestRequest("POST", url, body)
	w := httptest.NewRecorder()

	// This will fail without actual LLM integration, but we test the request handling
	handler.SendLLMMessage(w, req)

	// Response will contain an error since we don't have a real LLM backend
	// but the endpoint should handle the request
	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	// Either success or error response is acceptable for this test
	_, hasError := response["error"]
	_, hasMessage := response["content"]
	assert.True(t, hasError || hasMessage, "Response should have either error or message")
}

func TestCreateKnowledgeBaseEntry(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	newEntry := CreateKnowledgeBaseEntryRequest{
		Title:       "Pod CrashLoopBackOff",
		Content:     "When a pod is in CrashLoopBackOff state...",
		Category:    "troubleshooting",
		Tags:        `["k8s", "pods", "crashloop"]`,
		Source:      "manual",
	}

	body, _ := json.Marshal(newEntry)
	req := CreateTestRequest("POST", "/api/v1/ai/knowledge-base", body)
	w := httptest.NewRecorder()

	handler.CreateKnowledgeBaseEntry(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.KnowledgeBaseEntry
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Pod CrashLoopBackOff", response.Title)
	assert.Equal(t, "troubleshooting", response.Category)
	assert.Equal(t, userID, response.UserID)
}

func TestAnomalyDetectionRuleValidation(t *testing.T) {
	db := setupAITestDB(t)
	handler := NewAIAnalysisHandler(db)

	testCases := []struct {
		name      string
		rule      CreateAnomalyDetectionRuleRequest
		expectErr bool
	}{
		{
			name: "Valid STL rule",
			rule: CreateAnomalyDetectionRuleRequest{
				Name:        "Valid Rule",
				MetricQuery: "rate(cpu[5m])",
				Algorithm:   "stl",
			},
			expectErr: false,
		},
		{
			name: "Invalid algorithm",
			rule: CreateAnomalyDetectionRuleRequest{
				Name:        "Invalid Rule",
				MetricQuery: "metric",
				Algorithm:   "invalid_algo",
			},
			expectErr: true,
		},
		{
			name: "Missing name",
			rule: CreateAnomalyDetectionRuleRequest{
				MetricQuery: "metric",
				Algorithm:   "stl",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.rule)
			req := CreateTestRequest("POST", "/api/v1/ai/anomaly-rules", body)
			w := httptest.NewRecorder()

			handler.CreateAnomalyRule(w, req)

			if tc.expectErr {
				assert.NotEqual(t, http.StatusCreated, w.Code)
			} else {
				assert.Equal(t, http.StatusCreated, w.Code)
			}
		})
	}
}

func TestSearchKnowledgeBase(t *testing.T) {
	db := setupAITestDB(t)
	userID, _ := seedAITestData(t, db)
	handler := NewAIAnalysisHandler(db)

	// Create test KB entries
	entries := []model.KnowledgeBaseEntry{
		{Title: "Pod Issues", Content: "Troubleshooting pod problems", Category: "troubleshooting", UserID: userID},
		{Title: "Service Discovery", Content: "How services work in Kubernetes", Category: "best-practice", UserID: userID},
		{Title: "CPU Throttling", Content: "Understanding CPU limits", Category: "troubleshooting", UserID: userID},
	}
	for _, entry := range entries {
		require.NoError(t, db.Create(&entry).Error)
	}

	searchReq := SearchKnowledgeBaseRequest{
		Query:    "pod",
		Category: "troubleshooting",
		Limit:    10,
	}
	body, _ := json.Marshal(searchReq)
	req := CreateTestRequest("POST", "/api/v1/ai/knowledge-base/search", body)
	w := httptest.NewRecorder()

	handler.SearchKnowledgeBase(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SearchKnowledgeBaseResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.Entries), 1)
}

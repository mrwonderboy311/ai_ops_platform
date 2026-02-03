// Package model provides data models for AI analysis
package model

import (
	"time"

	"github.com/google/uuid"
)

// AnomalyDetectionRule represents an anomaly detection rule configuration
type AnomalyDetectionRule struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_anomaly_user_id" json:"userId"`
	ClusterID    *uuid.UUID `gorm:"type:uuid;index:idx_anomaly_cluster_id" json:"clusterId,omitempty"`
	DataSourceID *uuid.UUID `gorm:"type:uuid;index:idx_anomaly_datasource_id" json:"dataSourceId,omitempty"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description,omitempty"`

	// Detection configuration
	MetricQuery  string `gorm:"type:text;not null" json:"metricQuery"` // PromQL query
	Algorithm    string `gorm:"size:50;not null" json:"algorithm"`    // stl, isolation_forest, lstm, baseline
	Sensitivity  float64 `gorm:"default:0.95" json:"sensitivity"`      // 0-1
	WindowSize   int     `gorm:"default:100" json:"windowSize"`       // data points
	MinValue     float64 `json:"minValue,omitempty"`
	MaxValue     float64 `json:"maxValue,omitempty"`

	// Schedule
	Enabled      bool   `gorm:"default:true" json:"enabled"`
	EvalInterval int    `gorm:"default:300" json:"evalInterval"` // seconds
	LastEvalAt   *time.Time `json:"lastEvalAt,omitempty"`

	// Alert configuration
	AlertThreshold   float64 `gorm:"default:0.8" json:"alertThreshold"`
	AlertOnRecovery  bool    `gorm:"default:false" json:"alertOnRecovery"`
	NotificationChannels string `gorm:"type:text" json:"notificationChannels,omitempty"` // JSON array

	// Statistics
	TotalEvaluations int       `gorm:"default:0" json:"totalEvaluations"`
	AnomaliesDetected int      `gorm:"default:0" json:"anomaliesDetected"`
	LastAnomalyAt    *time.Time `json:"lastAnomalyAt,omitempty"`

	// Relationships
	Cluster    *K8sCluster         `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	DataSource *PrometheusDataSource `gorm:"foreignKey:DataSourceID" json:"dataSource,omitempty"`
}

// AnomalyEvent represents a detected anomaly event
type AnomalyEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	RuleID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_anomaly_event_rule_id" json:"ruleId"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index:idx_anomaly_event_user_id" json:"userId"`
	ClusterID    *uuid.UUID `gorm:"type:uuid;index:idx_anomaly_event_cluster_id" json:"clusterId,omitempty"`
	Severity     string     `gorm:"size:50;not null" json:"severity"` // critical, warning, info

	// Anomaly details
	MetricName   string    `gorm:"size:255;not null" json:"metricName"`
	CurrentValue float64  `json:"currentValue"`
	ExpectedValue float64  `json:"expectedValue"`
	Deviation    float64  `json:"deviation"` // standard deviations from expected
	Confidence   float64  `json:"confidence"` // 0-1
	TimeRange    string    `gorm:"size:100" json:"timeRange,omitempty"`

	// Context
	Labels       string    `gorm:"type:text" json:"labels,omitempty"` // JSON object
	Description  string    `gorm:"type:text" json:"description"`
	Suggestions  string    `gorm:"type:text" json:"suggestions,omitempty"`

	// Resolution
	Status       string    `gorm:"size:50;default:AnomalyStatusActive" json:"status"` // active, acknowledged, resolved, false_positive
	AcknowledgedAt *time.Time `json:"acknowledgedAt,omitempty"`
	AcknowledgedBy  *uuid.UUID `json:"acknowledgedBy,omitempty"`
	ResolvedAt      *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy      *uuid.UUID `json:"resolvedBy,omitempty"`

	// Relationships
	Rule        *AnomalyDetectionRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	Cluster     *K8sCluster           `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// AlertGroup represents a group of related alerts
type AlertGroup struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_alert_group_user_id" json:"userId"`
	ClusterID *uuid.UUID `gorm:"type:uuid;index:idx_alert_group_cluster_id" json:"clusterId,omitempty"`

	// Grouping
	GroupKey   string    `gorm:"size:500;not null;index" json:"groupKey"` // hash of grouping attributes
	Name       string    `gorm:"size:500;not null" json:"name"`
	Severity   string    `gorm:"size:50" json:"severity"`

	// Alert details
	AlertCount  int       `gorm:"default:1" json:"alertCount"`
	AlertIDs    string    `gorm:"type:text" json:"alertIds"` // JSON array
	FirstAlertAt time.Time `json:"firstAlertAt"`
	LastAlertAt  time.Time `json:"lastAlertAt"`

	// Root cause analysis
	RootCause   string    `gorm:"type:text" json:"rootCause,omitempty"`
	Confidence  float64   `json:"confidence,omitempty"`
	Suggestions string    `gorm:"type:text" json:"suggestions,omitempty"`

	// Status
	Status      string    `gorm:"size:50;default:AlertGroupStatusActive" json:"status"` // active, suppressed, resolved
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`

	// Relationships
	Cluster *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// LLMConversation represents a conversation with the LLM assistant
type LLMConversation struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_llm_conversation_user_id" json:"userId"`
	ClusterID *uuid.UUID `gorm:"type:uuid;index:idx_llm_conversation_cluster_id" json:"clusterId,omitempty"`
	Title     string    `gorm:"size:500" json:"title,omitempty"`

	// Configuration
	Model      string    `gorm:"size:100;not null" json:"model"` // gpt-4, claude-3, etc.
	Temperature float64  `gorm:"default:0.7" json:"temperature"`
	MaxTokens   int      `gorm:"default:2000" json:"maxTokens"`
	SystemPrompt string `gorm:"type:text" json:"systemPrompt,omitempty"`

	// Context
	RelatedAlertIDs  string `gorm:"type:text" json:"relatedAlertIds,omitempty"` // JSON array
	RelatedAnomalyIDs string `gorm:"type:text" json:"relatedAnomalyIds,omitempty"` // JSON array

	// Relationships
	Cluster    *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	Messages   []LLMMessage `gorm:"foreignKey:ConversationID" json:"messages,omitempty"`
}

// LLMMessage represents a message in an LLM conversation
type LLMMessage struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`

	ConversationID uuid.UUID `gorm:"type:uuid;not null;index:idx_llm_message_conversation_id" json:"conversationId"`
	Role           string    `gorm:"size:20;not null" json:"role"` // user, assistant, system
	Content        string    `gorm:"type:text;not null" json:"content"`
	TokensUsed     int       `json:"tokensUsed,omitempty"`

	// Metadata
	RelatedQuery   string    `gorm:"type:text" json:"relatedQuery,omitempty"` // The query that generated this response
	RelatedData    string    `gorm:"type:text" json:"relatedData,omitempty"` // JSON: metrics, logs, traces referenced

	// Relationships
	Conversation *LLMConversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
}

// NLQuery represents a natural language query
type NLQuery struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_nl_query_user_id" json:"userId"`
	ClusterID *uuid.UUID `gorm:"type:uuid;index:idx_nl_query_cluster_id" json:"clusterId,omitempty"`

	// Query details
	OriginalText string    `gorm:"type:text;not null" json:"originalText"`
	InterpretedQuery string `gorm:"type:text;not null" json:"interpretedQuery"` // PromQL/LogQL
	QueryType    string    `gorm:"size:50;not null" json:"queryType"` // metric, log, trace, mixed
	Confidence   float64   `json:"confidence"` // 0-1, how confident the LLM was in interpretation

	// Execution
	Executed     bool      `gorm:"default:false" json:"executed"`
	ExecutedAt   *time.Time `json:"executedAt,omitempty"`
	ResultCount  int       `json:"resultCount,omitempty"`
	ExecutionTime int64    `json:"executionTime,omitempty"` // milliseconds
	Error        string    `gorm:"type:text" json:"error,omitempty"`

	// Feedback
	Helpful      *bool     `json:"helpful,omitempty"` // User feedback
	Feedback     string    `gorm:"type:text" json:"feedback,omitempty"`

	// Relationships
	Cluster *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// KnowledgeBaseEntry represents a knowledge base entry for RAG
type KnowledgeBaseEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_kb_user_id" json:"userId"`
	ClusterID *uuid.UUID `gorm:"type:uuid;index:idx_kb_cluster_id" json:"clusterId,omitempty"`

	// Content
	Title       string    `gorm:"size:500;not null" json:"title"`
	Content     string    `gorm:"type:text;not null" json:"content"`
	Category    string    `gorm:"size:100" json:"category,omitempty"` // troubleshooting, best-practice, runbook
	Tags        string    `gorm:"type:text" json:"tags,omitempty"` // JSON array

	// Vector embedding (for semantic search)
	Embedding   []float64 `gorm:"type:vector" json:"-"` // Stored separately in vector DB
	EmbeddingModel string `gorm:"size:100" json:"embeddingModel,omitempty"`

	// Metadata
	Source      string    `gorm:"size:100" json:"source,omitempty"` // manual, imported, llm-generated
	SourceURL   string    `gorm:"size:500" json:"sourceUrl,omitempty"`
	LastUsedAt  *time.Time `json:"lastUsedAt,omitempty"`
	UseCount    int       `gorm:"default:0" json:"useCount"`

	// Relationships
	Cluster *K8sCluster `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
}

// BaselineMetric represents a learned baseline for a metric
type BaselineMetric struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`

	UserID       uuid.UUID `gorm:"type:uuid;not null;index:idx_baseline_user_id" json:"userId"`
	ClusterID    *uuid.UUID `gorm:"type:uuid;index:idx_baseline_cluster_id" json:"clusterId,omitempty"`
	DataSourceID *uuid.UUID `gorm:"type:uuid;index:idx_baseline_datasource_id" json:"dataSourceId,omitempty"`

	// Metric identity
	MetricName   string  `gorm:"size:255;not null" json:"metricName"`
	Labels       string `gorm:"type:text" json:"labels,omitempty"` // JSON object for label matching

	// Baseline data (seasonal decomposition)
	Trend        string  `gorm:"type:text" json:"trend"` // JSON: trend component
	Seasonal     string  `gorm:"type:text" json:"seasonal"` // JSON: seasonal component
	Residual     string  `gorm:"type:text" json:"residual"` // JSON: residual component

	// Statistics
	Mean         float64 `json:"mean"`
	StdDev       float64 `json:"stdDev"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Percentile25 float64 `json:"percentile25"`
	Percentile50 float64 `json:"percentile50"`
	Percentile75 float64 `json:"percentile75"`
	Percentile95 float64 `json:"percentile95"`
	Percentile99 float64 `json:"percentile99"`

	// Learning config
	Granularity  int     `gorm:"default:300" json:"granularity"` // seconds
	Seasonality  string  `gorm:"size:50;default:hourly" json:"seasonality"` // hourly, daily, weekly
	TrainingWindow int   `gorm:"default:7" json:"trainingWindow"` // days
	LastTrainedAt *time.Time `json:"lastTrainedAt"`
	ValidUntil   *time.Time `json:"validUntil,omitempty"`

	// Relationships
	Cluster    *K8sCluster         `gorm:"foreignKey:ClusterID" json:"cluster,omitempty"`
	DataSource *PrometheusDataSource `gorm:"foreignKey:DataSourceID" json:"dataSource,omitempty"`
}

// Anomaly status constants
const (
	AnomalyStatusActive        = "active"
	AnomalyStatusAcknowledged  = "acknowledged"
	AnomalyStatusResolved      = "resolved"
	AnomalyStatusFalsePositive = "false_positive"
)

// Alert group status constants
const (
	AlertGroupStatusActive     = "active"
	AlertGroupStatusSuppressed = "suppressed"
	AlertGroupStatusResolved   = "resolved"
)

// Algorithm constants
const (
	AlgorithmSTL              = "stl"
	AlgorithmIsolationForest  = "isolation_forest"
	AlgorithmLSTM             = "lstm"
	AlgorithmBaseline         = "baseline"
)

// CreateAnomalyDetectionRuleRequest represents a request to create an anomaly detection rule
type CreateAnomalyDetectionRuleRequest struct {
	ClusterID    *uuid.UUID `json:"clusterId,omitempty"`
	DataSourceID *uuid.UUID `json:"dataSourceId,omitempty"`
	Name         string     `json:"name" binding:"required"`
	Description  string     `json:"description,omitempty"`
	MetricQuery  string     `json:"metricQuery" binding:"required"`
	Algorithm    string     `json:"algorithm" binding:"required,oneof=stl isolation_forest lstm baseline"`
	Sensitivity  float64    `json:"sensitivity"`
	WindowSize   int        `json:"windowSize"`
	MinValue     float64    `json:"minValue,omitempty"`
	MaxValue     float64    `json:"maxValue,omitempty"`
	Enabled      bool       `json:"enabled"`
	EvalInterval int        `json:"evalInterval"`
	AlertThreshold   float64 `json:"alertThreshold"`
	AlertOnRecovery  bool    `json:"alertOnRecovery"`
	NotificationChannels string `json:"notificationChannels,omitempty"`
}

// UpdateAnomalyDetectionRuleRequest represents a request to update an anomaly detection rule
type UpdateAnomalyDetectionRuleRequest struct {
	Name         *string  `json:"name,omitempty"`
	Description  *string  `json:"description,omitempty"`
	MetricQuery  *string  `json:"metricQuery,omitempty"`
	Algorithm    *string  `json:"algorithm,omitempty"`
	Sensitivity  *float64 `json:"sensitivity,omitempty"`
	WindowSize   *int     `json:"windowSize,omitempty"`
	MinValue     *float64 `json:"minValue,omitempty"`
	MaxValue     *float64 `json:"maxValue,omitempty"`
	Enabled      *bool    `json:"enabled,omitempty"`
	EvalInterval *int     `json:"evalInterval,omitempty"`
	AlertThreshold   *float64 `json:"alertThreshold,omitempty"`
	AlertOnRecovery  *bool    `json:"alertOnRecovery,omitempty"`
	NotificationChannels *string `json:"notificationChannels,omitempty"`
}

// ExecuteAnomalyDetectionRequest represents a request to run anomaly detection
type ExecuteAnomalyDetectionRequest struct {
	RuleID    uuid.UUID `json:"ruleId" binding:"required"`
	StartTime string   `json:"startTime,omitempty"` // RFC3339
	EndTime   string   `json:"endTime,omitempty"`   // RFC3339
}

// ExecuteAnomalyDetectionResponse represents the response from anomaly detection
type ExecuteAnomalyDetectionResponse struct {
	Anomalies      []AnomalyEvent `json:"anomalies,omitempty"`
	AnomalyCount   int            `json:"anomalyCount"`
	EvaluatedAt    time.Time      `json:"evaluatedAt"`
	Duration       int64          `json:"duration"` // milliseconds
}

// CreateLLMConversationRequest represents a request to create an LLM conversation
type CreateLLMConversationRequest struct {
	ClusterID   *uuid.UUID `json:"clusterId,omitempty"`
	Title       string     `json:"title,omitempty"`
	Model       string     `json:"model" binding:"required"`
	Temperature float64    `json:"temperature"`
	MaxTokens   int        `json:"maxTokens"`
	SystemPrompt string    `json:"systemPrompt,omitempty"`
}

// SendLLMMessageRequest represents a request to send a message in a conversation
type SendLLMMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// SendLLMMessageResponse represents the response from sending a message
type SendLLMMessageResponse struct {
	MessageID    string `json:"messageId"`
	Content      string `json:"content"`
	TokensUsed   int    `json:"tokensUsed"`
	RelatedQuery string `json:"relatedQuery,omitempty"`
	RelatedData  string `json:"relatedData,omitempty"`
}

// ProcessNLQueryRequest represents a request to process a natural language query
type ProcessNLQueryRequest struct {
	ClusterID   *uuid.UUID `json:"clusterId,omitempty"`
	Query       string     `json:"query" binding:"required"`
	Context     string     `json:"context,omitempty"` // Additional context
}

// ProcessNLQueryResponse represents the response from processing an NL query
type ProcessNLQueryRequest struct {
	QueryID          string   `json:"queryId"`
	InterpretedQuery string   `json:"interpretedQuery"`
	QueryType        string   `json:"queryType"`
	Confidence       float64  `json:"confidence"`
	Results          string   `json:"results,omitempty"` // JSON: query results
	ExecutionTime    int64    `json:"executionTime"` // milliseconds
}

// CreateKnowledgeBaseEntryRequest represents a request to create a KB entry
type CreateKnowledgeBaseEntryRequest struct {
	ClusterID *uuid.UUID `json:"clusterId,omitempty"`
	Title     string     `json:"title" binding:"required"`
	Content   string     `json:"content" binding:"required"`
	Category  string     `json:"category,omitempty"`
	Tags      string     `json:"tags,omitempty"`
	Source    string     `json:"source,omitempty"`
	SourceURL string     `json:"sourceUrl,omitempty"`
}

// SearchKnowledgeBaseRequest represents a request to search the knowledge base
type SearchKnowledgeBaseRequest struct {
	ClusterID *uuid.UUID `json:"clusterId,omitempty"`
	Query     string     `json:"query" binding:"required"`
	Category  string     `json:"category,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Limit     int        `json:"limit"`
}

// SearchKnowledgeBaseResponse represents the response from searching the KB
type SearchKnowledgeBaseResponse struct {
	Entries []KnowledgeBaseEntry `json:"entries"`
	Total   int                   `json:"total"`
}

// TrainBaselineRequest represents a request to train a baseline model
type TrainBaselineRequest struct {
	DataSourceID *uuid.UUID `json:"dataSourceId" binding:"required"`
	MetricName   string     `json:"metricName" binding:"required"`
	Labels       string     `json:"labels,omitempty"`
	Granularity  int        `json:"granularity"` // seconds
	Seasonality  string     `json:"seasonality"` // hourly, daily, weekly
	TrainingWindow int     `json:"trainingWindow"` // days
}

// TrainBaselineResponse represents the response from baseline training
type TrainBaselineResponse struct {
	BaselineID   string    `json:"baselineId"`
	MetricName   string    `json:"metricName"`
	Accuracy     float64   `json:"accuracy"`
	ValidUntil   time.Time `json:"validUntil"`
	Duration     int64     `json:"duration"` // milliseconds
}

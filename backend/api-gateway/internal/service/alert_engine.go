// Package service provides alert rule evaluation engine
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AlertEngine evaluates alert rules and creates alerts
type AlertEngine struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAlertEngine creates a new alert engine
func NewAlertEngine(db *gorm.DB, logger *zap.Logger) *AlertEngine {
	return &AlertEngine{
		db:     db,
		logger: logger,
	}
}

// EvaluateRules evaluates all enabled alert rules
func (e *AlertEngine) EvaluateRules(ctx context.Context) error {
	var rules []model.AlertRule
	if err := e.db.Where("enabled = ? AND (silenced_until IS NULL OR silenced_until < ?)", true, time.Now()).Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to fetch alert rules: %w", err)
	}

	for _, rule := range rules {
		if err := e.EvaluateRule(ctx, &rule); err != nil {
			e.logger.Error("failed to evaluate rule",
				zap.String("ruleId", rule.ID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

// EvaluateRule evaluates a single alert rule
func (e *AlertEngine) EvaluateRule(ctx context.Context, rule *model.AlertRule) error {
	// Update last evaluated time
	now := time.Now()
	rule.LastEvaluatedAt = &now
	e.db.Save(rule)

	// Get current metric value based on rule type
	value, err := e.getMetricValue(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to get metric value: %w", err)
	}

	// Check if condition is met
	conditionMet := e.checkCondition(value, rule.Operator, rule.Threshold)

	// Check for existing firing alert for this rule
	var existingAlert model.Alert
	err = e.db.Where("rule_id = ? AND status = ?", rule.ID, model.AlertStatusFiring).
		Order("started_at DESC").
		First(&existingAlert).Error

	if err == gorm.ErrRecordNotFound {
		// No existing alert
		if conditionMet {
			// Create new alert
			return e.createAlert(rule, value)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to query existing alert: %w", err)
	}

	// Existing alert found
	if conditionMet {
		// Update existing alert
		existingAlert.Value = value
		existingAlert.UpdatedAt = now
		return e.db.Save(&existingAlert).Error
	}

	// Condition no longer met, resolve the alert
	return e.resolveAlert(&existingAlert)
}

// getMetricValue retrieves the current value for the rule's metric
func (e *AlertEngine) getMetricValue(ctx context.Context, rule *model.AlertRule) (float64, error) {
	switch rule.MetricType {
	case "cpu_usage":
		return e.getHostCPUMetrics(rule)
	case "memory_usage":
		return e.getHostMemoryMetrics(rule)
	case "disk_usage":
		return e.getHostDiskMetrics(rule)
	case "cluster_cpu_usage":
		return e.getClusterCPUMetrics(rule)
	case "cluster_memory_usage":
		return e.getClusterMemoryMetrics(rule)
	case "node_status":
		return e.getNodeStatusMetrics(rule)
	case "pod_status":
		return e.getPodStatusMetrics(rule)
	default:
		return 0, fmt.Errorf("unknown metric type: %s", rule.MetricType)
	}
}

// checkCondition checks if the condition is met
func (e *AlertEngine) checkCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

// createAlert creates a new alert from a rule
func (e *AlertEngine) createAlert(rule *model.AlertRule, value float64) error {
	alert := &model.Alert{
		ID:        uuid.New(),
		RuleID:    rule.ID,
		UserID:    rule.UserID,
		Status:    model.AlertStatusFiring,
		Severity:  rule.Severity,
		Title:     fmt.Sprintf("%s: %s", rule.Severity, rule.Name),
		Description: fmt.Sprintf("Rule '%s' triggered. Current value: %.2f, Threshold: %.2f",
			rule.Name, value, rule.Threshold),
		Value:     value,
		Threshold: rule.Threshold,
		StartedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if rule.TargetType == "cluster" {
		clusterID, err := uuid.Parse(rule.TargetID)
		if err == nil {
			alert.ClusterID = &clusterID
		}
	} else if rule.TargetType == "host" {
		hostID, err := uuid.Parse(rule.TargetID)
		if err == nil {
			alert.HostID = &hostID
		}
	}

	if err := e.db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	e.logger.Info("alert created",
		zap.String("alertId", alert.ID.String()),
		zap.String("ruleId", rule.ID.String()),
		zap.String("title", alert.Title),
	)

	// Send notifications
	if rule.NotifyEmail || rule.NotifyWebhook {
		go e.sendNotifications(alert, rule)
	}

	return nil
}

// resolveAlert resolves an alert
func (e *AlertEngine) resolveAlert(alert *model.Alert) error {
	now := time.Now()
	alert.Status = model.AlertStatusResolved
	alert.ResolvedAt = &now
	alert.UpdatedAt = now

	if err := e.db.Save(alert).Error; err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	e.logger.Info("alert resolved",
		zap.String("alertId", alert.ID.String()),
		zap.String("title", alert.Title),
	)

	return nil
}

// sendNotifications sends notifications for an alert
func (e *AlertEngine) sendNotifications(alert *model.Alert, rule *model.AlertRule) {
	// Create email notification
	if rule.NotifyEmail {
		notification := &model.AlertNotification{
			ID:        uuid.New(),
			AlertID:   alert.ID,
			Type:      "email",
			Status:    "pending",
			Recipient: "", // Would be user's email
			Content:   fmt.Sprintf("Alert: %s\n\n%s", alert.Title, alert.Description),
		}
		e.db.Create(notification)
		// TODO: Actually send email
	}

	// Create webhook notification
	if rule.NotifyWebhook && rule.WebhookURL != "" {
		notification := &model.AlertNotification{
			ID:        uuid.New(),
			AlertID:   alert.ID,
			Type:      "webhook",
			Status:    "pending",
			Recipient: rule.WebhookURL,
			Content:   fmt.Sprintf(`{"alert_id":"%s","title":"%s","severity":"%s","value":%.2f}`,
				alert.ID.String(), alert.Title, alert.Severity, alert.Value),
		}
		e.db.Create(notification)
		// TODO: Actually send webhook
	}
}

// Helper methods for getting metric values
// These would be implemented based on your monitoring data

func (e *AlertEngine) getHostCPUMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement host CPU metrics retrieval
	return 0, nil
}

func (e *AlertEngine) getHostMemoryMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement host memory metrics retrieval
	return 0, nil
}

func (e *AlertEngine) getHostDiskMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement host disk metrics retrieval
	return 0, nil
}

func (e *AlertEngine) getClusterCPUMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement cluster CPU metrics retrieval
	return 0, nil
}

func (e *AlertEngine) getClusterMemoryMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement cluster memory metrics retrieval
	return 0, nil
}

func (e *AlertEngine) getNodeStatusMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement node status metrics retrieval
	return 0, nil
}

func (e *AlertEngine) getPodStatusMetrics(rule *model.AlertRule) (float64, error) {
	// TODO: Implement pod status metrics retrieval
	return 0, nil
}

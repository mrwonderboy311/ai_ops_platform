// Package handler provides HTTP handlers for API Gateway
package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

var (
	hostHandler         *HostHandler
	scanHandler         *ScanHandler
	agentHandler        *AgentHandler
	fileHandler         *FileTransferHandler
	processHandler      *ProcessManagementHandler
	batchTaskHandler    *BatchTaskHandler
	clusterHandler      *ClusterHandler
	clusterMetricsHandler *ClusterMetricsHandler
	workloadHandler     *WorkloadHandler
	alertHandler        *AlertHandler
	auditHandler        *AuditHandler
	performanceHandler  *PerformanceHandler
	notificationHandler *NotificationHandler
)

// RegisterHandlers registers the API handlers
func RegisterHandlers(hostH *HostHandler, scanH *ScanHandler, agentH *AgentHandler) {
	hostHandler = hostH
	scanHandler = scanH
	agentHandler = agentH
}

// RegisterFileHandler registers the file transfer handler
func RegisterFileHandler(fileH *FileTransferHandler) {
	fileHandler = fileH
}

// RegisterProcessHandler registers the process management handler
func RegisterProcessHandler(processH *ProcessManagementHandler) {
	processHandler = processH
}

// RegisterBatchTaskHandler registers the batch task handler
func RegisterBatchTaskHandler(batchH *BatchTaskHandler) {
	batchTaskHandler = batchH
}

// RegisterClusterHandler registers the cluster handler
func RegisterClusterHandler(clusterH *ClusterHandler) {
	clusterHandler = clusterH
}

// RegisterClusterMetricsHandler registers the cluster metrics handler
func RegisterClusterMetricsHandler(metricsH *ClusterMetricsHandler) {
	clusterMetricsHandler = metricsH
}

// RegisterWorkloadHandler registers the workload handler
func RegisterWorkloadHandler(workloadH *WorkloadHandler) {
	workloadHandler = workloadH
}

// RegisterAlertHandler registers the alert handler
func RegisterAlertHandler(alertH *AlertHandler) {
	alertHandler = alertH
}

// RegisterAuditHandler registers the audit handler
func RegisterAuditHandler(auditH *AuditHandler) {
	auditHandler = auditH
}

// RegisterPerformanceHandler registers the performance handler
func RegisterPerformanceHandler(perfH *PerformanceHandler) {
	performanceHandler = perfH
}

// RegisterNotificationHandler registers the notification handler
func RegisterNotificationHandler(notifH *NotificationHandler) {
	notificationHandler = notifH
}

// Health returns the health check response
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// API handles all API requests
func API(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := r.URL.Path
	method := r.Method

	// Route to appropriate handler
	if strings.HasPrefix(path, "/api/v1/agent/report") {
		// Agent reporting endpoint
		if agentHandler != nil {
			agentHandler.ServeHTTP(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Agent service not available")
		}
		return
	}

	if strings.HasPrefix(path, "/api/v1/hosts/scan-tasks/") {
		// Scan task status query
		if scanHandler != nil {
			scanHandler.GetScanStatus(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Scan service not available")
		}
		return
	}

	// File transfer endpoints
	if strings.HasPrefix(path, "/api/v1/files/") && fileHandler != nil {
		switch {
		case path == "/api/v1/files/list" && method == http.MethodPost:
			fileHandler.ListDirectory(w, r)
		case path == "/api/v1/files/upload" && method == http.MethodPost:
			fileHandler.UploadFile(w, r)
		case path == "/api/v1/files/download" && method == http.MethodPost:
			fileHandler.DownloadFile(w, r)
		case path == "/api/v1/files/delete" && method == http.MethodPost:
			fileHandler.DeleteFile(w, r)
		case path == "/api/v1/files/mkdir" && method == http.MethodPost:
			fileHandler.CreateDirectory(w, r)
		case path == "/api/v1/files/transfers" && method == http.MethodGet:
			fileHandler.GetTransfers(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "File operation not found")
		}
		return
	}

	// Host-specific file transfer endpoints
	if matchesPattern(path, "/api/v1/hosts/*/transfers") && method == http.MethodGet {
		if fileHandler != nil {
			fileHandler.GetTransfers(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "File service not available")
		}
		return
	}

	// Process management endpoints
	if strings.HasPrefix(path, "/api/v1/processes/") && processHandler != nil {
		switch {
		case path == "/api/v1/processes/list" && method == http.MethodPost:
			processHandler.ListProcesses(w, r)
		case path == "/api/v1/processes/get" && method == http.MethodPost:
			processHandler.GetProcess(w, r)
		case path == "/api/v1/processes/kill" && method == http.MethodPost:
			processHandler.KillProcess(w, r)
		case path == "/api/v1/processes/execute" && method == http.MethodPost:
			processHandler.ExecuteCommand(w, r)
		case path == "/api/v1/processes/executions" && method == http.MethodGet:
			processHandler.GetExecutions(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Process operation not found")
		}
		return
	}

	// Host-specific process execution endpoints
	if matchesPattern(path, "/api/v1/hosts/*/executions") && method == http.MethodGet {
		if processHandler != nil {
			processHandler.GetExecutions(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Process service not available")
		}
		return
	}

	// Batch task endpoints
	if strings.HasPrefix(path, "/api/v1/batch-tasks") && batchTaskHandler != nil {
		switch {
		case path == "/api/v1/batch-tasks" && method == http.MethodPost:
			batchTaskHandler.CreateBatchTask(w, r)
		case path == "/api/v1/batch-tasks" && method == http.MethodGet:
			batchTaskHandler.ListBatchTasks(w, r)
		case path == "/api/v1/batch-tasks/execute" && method == http.MethodPost:
			batchTaskHandler.ExecuteBatchTask(w, r)
		case path == "/api/v1/batch-tasks/cancel" && method == http.MethodPost:
			batchTaskHandler.CancelBatchTask(w, r)
		case matchesPattern(path, "/api/v1/batch-tasks/*"):
			if method == http.MethodGet {
				batchTaskHandler.GetBatchTask(w, r)
			} else if method == http.MethodDelete {
				batchTaskHandler.DeleteBatchTask(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Batch task operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Batch task operation not found")
		}
		return
	}

	// Cluster management endpoints
	if strings.HasPrefix(path, "/api/v1/clusters") && clusterHandler != nil {
		// Check for metrics endpoints first
		if clusterMetricsHandler != nil {
			switch {
			case matchesPattern(path, "/api/v1/clusters/*/metrics") && method == http.MethodGet:
				clusterMetricsHandler.GetClusterMetrics(w, r)
				return
			case matchesPattern(path, "/api/v1/clusters/*/metrics/summary") && method == http.MethodGet:
				clusterMetricsHandler.GetClusterMetricsSummary(w, r)
				return
			case matchesPattern(path, "/api/v1/clusters/*/metrics/live") && method == http.MethodGet:
				clusterMetricsHandler.GetLiveClusterMetrics(w, r)
				return
			case matchesPattern(path, "/api/v1/clusters/*/refresh") && method == http.MethodPost:
				clusterMetricsHandler.RefreshMetrics(w, r)
				return
			case matchesPattern(path, "/api/v1/clusters/*/namespaces") && method == http.MethodGet:
				clusterMetricsHandler.ListNamespaces(w, r)
				return
			case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/metrics") && method == http.MethodGet:
				clusterMetricsHandler.GetPodMetrics(w, r)
				return
			case matchesPattern(path, "/api/v1/nodes/*/metrics") && method == http.MethodGet:
				clusterMetricsHandler.GetNodeMetrics(w, r)
				return
			case matchesPattern(path, "/api/v1/nodes/*/live-metrics") && method == http.MethodGet:
				clusterMetricsHandler.GetLiveNodeMetrics(w, r)
				return
			}
		}

		switch {
		case path == "/api/v1/clusters" && method == http.MethodPost:
			clusterHandler.CreateCluster(w, r)
		case path == "/api/v1/clusters" && method == http.MethodGet:
			clusterHandler.ListClusters(w, r)
		case path == "/api/v1/clusters/test-connection" && method == http.MethodPost:
			clusterHandler.TestConnection(w, r)
		case matchesPattern(path, "/api/v1/clusters/*/nodes") && method == http.MethodGet:
			clusterHandler.GetClusterNodes(w, r)
		case matchesPattern(path, "/api/v1/clusters/*/info") && method == http.MethodGet:
			clusterHandler.GetClusterInfo(w, r)
		case matchesPattern(path, "/api/v1/clusters/*"):
			if method == http.MethodGet {
				clusterHandler.GetCluster(w, r)
			} else if method == http.MethodPut {
				clusterHandler.UpdateCluster(w, r)
			} else if method == http.MethodDelete {
				clusterHandler.DeleteCluster(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Cluster operation not found")
		}
		return
	}

	// Workload management endpoints
	if strings.HasPrefix(path, "/api/v1/clusters") && workloadHandler != nil {
		switch {
		case matchesPattern(path, "/api/v1/clusters/*/namespaces") && method == http.MethodGet:
			workloadHandler.ListNamespaces(w, r)
			return
		case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/deployments") && method == http.MethodGet:
			workloadHandler.ListDeployments(w, r)
			return
		case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/pods") && method == http.MethodGet:
			workloadHandler.ListPods(w, r)
			return
		case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/pods/*/logs") && method == http.MethodGet:
			workloadHandler.GetPodLogs(w, r)
			return
		case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/pods/*") && method == http.MethodDelete:
			workloadHandler.DeletePod(w, r)
			return
		case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/services") && method == http.MethodGet:
			workloadHandler.ListServices(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/api/v1/hosts") {
		// Host management endpoints
		if hostHandler != nil {
			hostHandler.ServeHTTP(w, r)
		} else {
			respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Host service not available")
		}
		return
	}

	// Alert management endpoints
	if strings.HasPrefix(path, "/api/v1/alerts") && alertHandler != nil {
		switch {
		case path == "/api/v1/alerts" && method == http.MethodGet:
			alertHandler.ListAlerts(w, r)
		case path == "/api/v1/alerts/statistics" && method == http.MethodGet:
			alertHandler.GetAlertStatistics(w, r)
		case matchesPattern(path, "/api/v1/alerts/*/silence") && method == http.MethodPost:
			alertHandler.SilenceAlert(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert operation not found")
		}
		return
	}

	// Alert rules endpoints
	if strings.HasPrefix(path, "/api/v1/alert-rules") && alertHandler != nil {
		switch {
		case path == "/api/v1/alert-rules" && method == http.MethodPost:
			alertHandler.CreateAlertRule(w, r)
		case path == "/api/v1/alert-rules" && method == http.MethodGet:
			alertHandler.ListAlertRules(w, r)
		case matchesPattern(path, "/api/v1/alert-rules/*"):
			if method == http.MethodGet {
				alertHandler.GetAlertRule(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				alertHandler.UpdateAlertRule(w, r)
			} else if method == http.MethodDelete {
				alertHandler.DeleteAlertRule(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule operation not found")
		}
		return
	}

	// Events endpoints
	if strings.HasPrefix(path, "/api/v1/events") && alertHandler != nil {
		switch {
		case path == "/api/v1/events" && method == http.MethodGet:
			alertHandler.ListEvents(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Event operation not found")
		}
		return
	}

	// Audit log endpoints
	if strings.HasPrefix(path, "/api/v1/audit-logs") && auditHandler != nil {
		switch {
		case path == "/api/v1/audit-logs" && method == http.MethodGet:
			auditHandler.ListAuditLogs(w, r)
		case path == "/api/v1/audit-logs/summary" && method == http.MethodGet:
			auditHandler.GetAuditLogSummary(w, r)
		case path == "/api/v1/audit-logs/user-activity" && method == http.MethodGet:
			auditHandler.GetUserActivity(w, r)
		case path == "/api/v1/audit-logs/resource-activity" && method == http.MethodGet:
			auditHandler.GetResourceActivity(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Audit log operation not found")
		}
		return
	}

	// Performance monitoring endpoints
	if strings.HasPrefix(path, "/api/v1/performance") && performanceHandler != nil {
		switch {
		case path == "/api/v1/performance/metrics" && method == http.MethodGet:
			performanceHandler.GetMetrics(w, r)
		case path == "/api/v1/performance/health" && method == http.MethodGet:
			performanceHandler.GetSystemHealth(w, r)
		case path == "/api/v1/performance/health/refresh" && method == http.MethodPost:
			performanceHandler.RefreshSystemHealth(w, r)
		case path == "/api/v1/performance/summary" && method == http.MethodGet:
			performanceHandler.GetPerformanceSummary(w, r)
		case path == "/api/v1/performance/trend" && method == http.MethodGet:
			performanceHandler.GetTrendData(w, r)
		case path == "/api/v1/performance/metrics" && method == http.MethodPost:
			performanceHandler.CollectMetric(w, r)
		case path == "/api/v1/performance/statistics" && method == http.MethodGet:
			performanceHandler.GetMetricStatistics(w, r)
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Performance operation not found")
		}
		return
	}

	// Notification endpoints
	if strings.HasPrefix(path, "/api/v1/notifications") && notificationHandler != nil {
		switch {
		case path == "/api/v1/notifications" && method == http.MethodGet:
			notificationHandler.GetNotifications(w, r)
		case path == "/api/v1/notifications" && method == http.MethodPost:
			notificationHandler.CreateNotification(w, r)
		case path == "/api/v1/notifications/unread-count" && method == http.MethodGet:
			notificationHandler.GetUnreadCount(w, r)
		case path == "/api/v1/notifications/mark-all-read" && method == http.MethodPost:
			notificationHandler.MarkAllAsRead(w, r)
		case path == "/api/v1/notifications/stats" && method == http.MethodGet:
			notificationHandler.GetNotificationStats(w, r)
		case path == "/api/v1/notifications/preferences" && method == http.MethodGet:
			notificationHandler.GetNotificationPreference(w, r)
		case path == "/api/v1/notifications/preferences" && method == http.MethodPut:
			notificationHandler.UpdateNotificationPreference(w, r)
		case matchesPattern(path, "/api/v1/notifications/*/read"):
			if method == http.MethodPost || method == http.MethodPut {
				notificationHandler.MarkAsRead(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Notification operation not found")
			}
		case matchesPattern(path, "/api/v1/notifications/*"):
			if method == http.MethodDelete {
				notificationHandler.DeleteNotification(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Notification operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Notification operation not found")
		}
		return
	}

	// Unknown endpoint
	respondWithError(w, http.StatusNotFound, "NOT_FOUND", "API endpoint not found")
}

// matchesPattern checks if a path matches a pattern with wildcards
func matchesPattern(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i := 0; i < len(patternParts); i++ {
		if patternParts[i] != "*" && patternParts[i] != pathParts[i] {
			return false
		}
	}

	return true
}

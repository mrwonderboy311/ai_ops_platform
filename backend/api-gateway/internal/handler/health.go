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
	podLogsWSHandler     *PodLogsWebSocketHandler
	podTerminalWSHandler *PodTerminalWebSocketHandler
	helmHandler         *HelmHandler
	otelHandler         *OtelHandler
	prometheusHandler   *PrometheusHandler
	grafanaHandler      *GrafanaHandler
	aiAnalysisHandler    *AIAnalysisHandler
	alertHandler        *AlertHandler
	auditHandler        *AuditHandler
	performanceHandler  *PerformanceHandler
	notificationHandler *NotificationHandler
	userManagementHandler *UserManagementHandler
	rbacHandler         *RBACHandler
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

// RegisterPodLogsWebSocketHandler registers the pod logs websocket handler
func RegisterPodLogsWebSocketHandler(wsH *PodLogsWebSocketHandler) {
	podLogsWSHandler = wsH
}

// RegisterPodTerminalWebSocketHandler registers the pod terminal websocket handler
func RegisterPodTerminalWebSocketHandler(wsH *PodTerminalWebSocketHandler) {
	podTerminalWSHandler = wsH
}

// RegisterHelmHandler registers the Helm handler
func RegisterHelmHandler(helmH *HelmHandler) {
	helmHandler = helmH
}

// RegisterOtelHandler registers the OpenTelemetry handler
func RegisterOtelHandler(otelH *OtelHandler) {
	otelHandler = otelH
}

// RegisterPrometheusHandler registers the Prometheus handler
func RegisterPrometheusHandler(promH *PrometheusHandler) {
	prometheusHandler = promH
}

// RegisterGrafanaHandler registers the Grafana handler
func RegisterGrafanaHandler(grafanaH *GrafanaHandler) {
	grafanaHandler = grafanaH
}

// RegisterAIAnalysisHandler registers the AI analysis handler
func RegisterAIAnalysisHandler(aiH *AIAnalysisHandler) {
	aiAnalysisHandler = aiH
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

// RegisterUserManagementHandler registers the user management handler
func RegisterUserManagementHandler(userMgmtH *UserManagementHandler) {
	userManagementHandler = userMgmtH
}

// RegisterRBACHandler registers the RBAC handler
func RegisterRBACHandler(rbacH *RBACHandler) {
	rbacHandler = rbacH
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
		// Websocket endpoint for pod logs streaming
		if path == "/api/v1/clusters/pod-logs/ws" && method == http.MethodGet {
			if podLogsWSHandler != nil {
				podLogsWSHandler.ServeHTTP(w, r)
			} else {
				respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "WebSocket service not available")
			}
			return
		}

		// Websocket endpoint for pod terminal
		if path == "/api/v1/clusters/pod-terminal/ws" && method == http.MethodGet {
			if podTerminalWSHandler != nil {
				podTerminalWSHandler.ServeHTTP(w, r)
			} else {
				respondWithError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "WebSocket service not available")
			}
			return
		}

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
		case matchesPattern(path, "/api/v1/clusters/*/namespaces/*/pods/*/detail") && method == http.MethodGet:
			workloadHandler.GetPodDetail(w, r)
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

	// User management endpoints
	if strings.HasPrefix(path, "/api/v1/users") && userManagementHandler != nil {
		switch {
		case path == "/api/v1/users" && method == http.MethodGet:
			userManagementHandler.ListUsers(w, r)
		case path == "/api/v1/users" && method == http.MethodPost:
			userManagementHandler.CreateUser(w, r)
		case path == "/api/v1/users/check-permission" && method == http.MethodGet:
			userManagementHandler.CheckPermission(w, r)
		case matchesPattern(path, "/api/v1/users/*/roles"):
			if method == http.MethodGet {
				userManagementHandler.GetUserRoles(w, r)
			} else if method == http.MethodPost {
				userManagementHandler.AssignRoleToUser(w, r)
			} else if method == http.MethodDelete {
				userManagementHandler.RemoveRoleFromUser(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User operation not found")
			}
		case matchesPattern(path, "/api/v1/users/*"):
			if method == http.MethodGet {
				userManagementHandler.GetUserByID(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				userManagementHandler.UpdateUser(w, r)
			} else if method == http.MethodDelete {
				userManagementHandler.DeleteUser(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User operation not found")
		}
		return
	}

	// Role management endpoints
	if strings.HasPrefix(path, "/api/v1/roles") && userManagementHandler != nil {
		switch {
		case path == "/api/v1/roles" && method == http.MethodGet:
			userManagementHandler.ListRoles(w, r)
		case path == "/api/v1/roles" && method == http.MethodPost:
			userManagementHandler.CreateRole(w, r)
		case path == "/api/v1/roles/permissions" && method == http.MethodGet:
			userManagementHandler.ListPermissions(w, r)
		case matchesPattern(path, "/api/v1/roles/*/permissions"):
			if method == http.MethodPost {
				userManagementHandler.AssignPermissionToRole(w, r)
			} else if method == http.MethodDelete {
				userManagementHandler.RemovePermissionFromRole(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role operation not found")
			}
		case matchesPattern(path, "/api/v1/roles/*"):
			if method == http.MethodGet {
				userManagementHandler.GetRoleByID(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				userManagementHandler.UpdateRole(w, r)
			} else if method == http.MethodDelete {
				userManagementHandler.DeleteRole(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role operation not found")
		}
		return
	}

	// Helm repository endpoints
	if strings.HasPrefix(path, "/api/v1/helm") && helmHandler != nil {
		// Helm releases endpoints
		if path == "/api/v1/helm/releases" && method == http.MethodGet {
			helmHandler.ListHelmRepos(w, r)
			return
		}
		if path == "/api/v1/helm/releases" && method == http.MethodPost {
			helmHandler.InstallHelmRelease(w, r)
			return
		}
		if matchesPattern(path, "/api/v1/helm/releases/*/history") && method == http.MethodGet {
			helmHandler.GetHelmReleaseHistory(w, r)
			return
		}
		if matchesPattern(path, "/api/v1/helm/releases/*/rollback") && method == http.MethodPost {
			helmHandler.RollbackHelmRelease(w, r)
			return
		}
		if matchesPattern(path, "/api/v1/helm/releases/*/upgrade") && method == http.MethodPost {
			helmHandler.UpgradeHelmRelease(w, r)
			return
		}
		if matchesPattern(path, "/api/v1/helm/releases/*"):
			if method == http.MethodGet {
				helmHandler.GetHelmRelease(w, r)
			} else if method == http.MethodDelete {
				helmHandler.UninstallHelmRelease(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Helm operation not found")
			}
			return
		}

		// Helm repository endpoints
		switch {
		case path == "/api/v1/helm/repositories" && method == http.MethodGet:
			helmHandler.ListHelmRepos(w, r)
		case path == "/api/v1/helm/repositories" && method == http.MethodPost:
			helmHandler.CreateHelmRepo(w, r)
		case path == "/api/v1/helm/repositories/test" && method == http.MethodPost:
			helmHandler.TestHelmRepo(w, r)
		case matchesPattern(path, "/api/v1/helm/repositories/*/sync") && method == http.MethodPost:
			helmHandler.SyncHelmRepo(w, r)
		case matchesPattern(path, "/api/v1/helm/repositories/*"):
			if method == http.MethodGet {
				helmHandler.GetHelmRepo(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				helmHandler.UpdateHelmRepo(w, r)
			} else if method == http.MethodDelete {
				helmHandler.DeleteHelmRepo(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Helm operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Helm operation not found")
		}
		return
	}

	// OpenTelemetry collector endpoints
	if strings.HasPrefix(path, "/api/v1/otel") && otelHandler != nil {
		switch {
		case path == "/api/v1/otel/collectors" && method == http.MethodGet:
			otelHandler.ListCollectors(w, r)
		case path == "/api/v1/otel/collectors" && method == http.MethodPost:
			otelHandler.CreateCollector(w, r)
		case matchesPattern(path, "/api/v1/otel/collectors/*/status") && method == http.MethodGet:
			otelHandler.GetCollectorStatus(w, r)
		case matchesPattern(path, "/api/v1/otel/collectors/*/start") && method == http.MethodPost:
			otelHandler.StartCollector(w, r)
		case matchesPattern(path, "/api/v1/otel/collectors/*/stop") && method == http.MethodPost:
			otelHandler.StopCollector(w, r)
		case matchesPattern(path, "/api/v1/otel/collectors/*/restart") && method == http.MethodPost:
			otelHandler.RestartCollector(w, r)
		case matchesPattern(path, "/api/v1/otel/collectors/*"):
			if method == http.MethodGet {
				otelHandler.GetCollector(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				otelHandler.UpdateCollector(w, r)
			} else if method == http.MethodDelete {
				otelHandler.DeleteCollector(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "OTel operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "OTel operation not found")
		}
		return
	}

	// Prometheus data source endpoints
	if strings.HasPrefix(path, "/api/v1/prometheus") && prometheusHandler != nil {
		// Query endpoints
		if matchesPattern(path, "/api/v1/prometheus/datasources/*/query") && method == http.MethodPost {
			prometheusHandler.ExecuteQuery(w, r)
			return
		}
		// Test connection endpoint
		if path == "/api/v1/prometheus/datasources/test" && method == http.MethodPost {
			prometheusHandler.TestDataSource(w, r)
			return
		}
		// Alert rules endpoints
		if strings.HasPrefix(path, "/api/v1/prometheus/alert-rules") {
			switch {
			case path == "/api/v1/prometheus/alert-rules" && method == http.MethodGet:
				prometheusHandler.ListAlertRules(w, r)
			case path == "/api/v1/prometheus/alert-rules" && method == http.MethodPost:
				prometheusHandler.CreateAlertRule(w, r)
			case matchesPattern(path, "/api/v1/prometheus/alert-rules/*"):
				if method == http.MethodGet {
					prometheusHandler.GetAlertRule(w, r)
				} else if method == http.MethodPut || method == http.MethodPatch {
					prometheusHandler.UpdateAlertRule(w, r)
				} else if method == http.MethodDelete {
					prometheusHandler.DeleteAlertRule(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Alert rule operation not found")
			}
			return
		}
		// Dashboard endpoints
		if strings.HasPrefix(path, "/api/v1/prometheus/dashboards") {
			switch {
			case path == "/api/v1/prometheus/dashboards" && method == http.MethodGet:
				prometheusHandler.ListDashboards(w, r)
			case path == "/api/v1/prometheus/dashboards" && method == http.MethodPost:
				prometheusHandler.CreateDashboard(w, r)
			case matchesPattern(path, "/api/v1/prometheus/dashboards/*"):
				if method == http.MethodGet {
					prometheusHandler.GetDashboard(w, r)
				} else if method == http.MethodPut || method == http.MethodPatch {
					prometheusHandler.UpdateDashboard(w, r)
				} else if method == http.MethodDelete {
					prometheusHandler.DeleteDashboard(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard operation not found")
			}
			return
		}
		// Data source endpoints
		switch {
		case path == "/api/v1/prometheus/datasources" && method == http.MethodGet:
			prometheusHandler.ListDataSources(w, r)
		case path == "/api/v1/prometheus/datasources" && method == http.MethodPost:
			prometheusHandler.CreateDataSource(w, r)
		case matchesPattern(path, "/api/v1/prometheus/datasources/*"):
			if method == http.MethodGet {
				prometheusHandler.GetDataSource(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				prometheusHandler.UpdateDataSource(w, r)
			} else if method == http.MethodDelete {
				prometheusHandler.DeleteDataSource(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Prometheus operation not found")
		}
		return
	}

	// Grafana instance endpoints
	if strings.HasPrefix(path, "/api/v1/grafana") && grafanaHandler != nil {
		// Sync endpoint
		if matchesPattern(path, "/api/v1/grafana/instances/*/sync") && method == http.MethodPost {
			grafanaHandler.SyncInstance(w, r)
			return
		}
		// Test endpoint
		if path == "/api/v1/grafana/instances/test" && method == http.MethodPost {
			grafanaHandler.TestInstance(w, r)
			return
		}
		// Dashboard endpoints
		if strings.HasPrefix(path, "/api/v1/grafana/dashboards") {
			switch {
			case path == "/api/v1/grafana/dashboards" && method == http.MethodGet:
				grafanaHandler.ListDashboards(w, r)
			case matchesPattern(path, "/api/v1/grafana/dashboards/*"):
				if method == http.MethodGet {
					grafanaHandler.GetDashboard(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Dashboard operation not found")
			}
			return
		}
		// Data source endpoints
		if strings.HasPrefix(path, "/api/v1/grafana/datasources") {
			switch {
			case path == "/api/v1/grafana/datasources" && method == http.MethodGet:
				grafanaHandler.ListDataSources(w, r)
			case matchesPattern(path, "/api/v1/grafana/datasources/*"):
				if method == http.MethodGet {
					grafanaHandler.GetDataSource(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Data source operation not found")
			}
			return
		}
		// Folder endpoints
		if strings.HasPrefix(path, "/api/v1/grafana/folders") {
			switch {
			case path == "/api/v1/grafana/folders" && method == http.MethodGet:
				grafanaHandler.ListFolders(w, r)
			case matchesPattern(path, "/api/v1/grafana/folders/*"):
				if method == http.MethodGet {
					grafanaHandler.GetFolder(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Folder operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Folder operation not found")
			}
			return
		}
		// Instance endpoints
		switch {
		case path == "/api/v1/grafana/instances" && method == http.MethodGet:
			grafanaHandler.ListInstances(w, r)
		case path == "/api/v1/grafana/instances" && method == http.MethodPost:
			grafanaHandler.CreateInstance(w, r)
		case matchesPattern(path, "/api/v1/grafana/instances/*"):
			if method == http.MethodGet {
				grafanaHandler.GetInstance(w, r)
			} else if method == http.MethodPut || method == http.MethodPatch {
				grafanaHandler.UpdateInstance(w, r)
			} else if method == http.MethodDelete {
				grafanaHandler.DeleteInstance(w, r)
			} else {
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Instance operation not found")
			}
		default:
			respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Grafana operation not found")
		}
		return
	}

	// AI Analysis endpoints
	if strings.HasPrefix(path, "/api/v1/ai") && aiAnalysisHandler != nil {
		// Anomaly detection endpoints
		if strings.HasPrefix(path, "/api/v1/ai/anomaly-events") {
			switch {
			case path == "/api/v1/ai/anomaly-events" && method == http.MethodGet:
				aiAnalysisHandler.ListAnomalyEvents(w, r)
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly event operation not found")
			}
			return
		}
		// Anomaly rules endpoints
		if strings.HasPrefix(path, "/api/v1/ai/anomaly-rules") {
			switch {
			case path == "/api/v1/ai/anomaly-rules" && method == http.MethodGet:
				aiAnalysisHandler.ListAnomalyRules(w, r)
			case path == "/api/v1/ai/anomaly-rules" && method == http.MethodPost:
				aiAnalysisHandler.CreateAnomalyRule(w, r)
			case path == "/api/v1/ai/anomaly-rules/execute" && method == http.MethodPost:
				aiAnalysisHandler.ExecuteAnomalyDetection(w, r)
			case matchesPattern(path, "/api/v1/ai/anomaly-rules/*"):
				if method == http.MethodGet {
					aiAnalysisHandler.GetAnomalyRule(w, r)
				} else if method == http.MethodPut || method == http.MethodPatch {
					aiAnalysisHandler.UpdateAnomalyRule(w, r)
				} else if method == http.MethodDelete {
					aiAnalysisHandler.DeleteAnomalyRule(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly rule operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Anomaly rule operation not found")
			}
			return
		}
		// LLM conversations endpoints
		if strings.HasPrefix(path, "/api/v1/ai/llm/conversations") {
			switch {
			case path == "/api/v1/ai/llm/conversations" && method == http.MethodGet:
				aiAnalysisHandler.ListLLMConversations(w, r)
			case path == "/api/v1/ai/llm/conversations" && method == http.MethodPost:
				aiAnalysisHandler.CreateLLMConversation(w, r)
			case matchesPattern(path, "/api/v1/ai/llm/conversations/*"):
				if method == http.MethodGet {
					aiAnalysisHandler.GetLLMConversation(w, r)
				} else if method == http.MethodDelete {
					aiAnalysisHandler.DeleteLLMConversation(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "LLM conversation operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "LLM conversation operation not found")
			}
			return
		}
		// LLM message endpoints
		if matchesPattern(path, "/api/v1/ai/llm/conversations/*/messages") && method == http.MethodPost {
			aiAnalysisHandler.SendLLMMessage(w, r)
			return
		}
	}

	// RBAC endpoints
	if strings.HasPrefix(path, "/api/v1/rbac") && rbacHandler != nil {
		// Current user endpoints
		if path == "/api/v1/rbac/me" && method == http.MethodGet {
			rbacHandler.GetCurrentUser(w, r)
			return
		}
		if path == "/api/v1/rbac/me/check-permissions" && method == http.MethodPost {
			rbacHandler.BatchCheckPermissions(w, r)
			return
		}
		// Permissions endpoints
		if strings.HasPrefix(path, "/api/v1/rbac/permissions") {
			switch {
			case path == "/api/v1/rbac/permissions" && method == http.MethodGet:
				rbacHandler.ListPermissions(w, r)
			case path == "/api/v1/rbac/permissions" && method == http.MethodPost:
				rbacHandler.CreatePermission(w, r)
			case matchesPattern(path, "/api/v1/rbac/permissions/*"):
				if method == http.MethodGet {
					rbacHandler.GetPermission(w, r)
				} else if method == http.MethodPut || method == http.MethodPatch {
					rbacHandler.UpdatePermission(w, r)
				} else if method == http.MethodDelete {
					rbacHandler.DeletePermission(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Permission operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Permission operation not found")
			}
			return
		}
		// Roles endpoints
		if strings.HasPrefix(path, "/api/v1/rbac/roles") {
			switch {
			case path == "/api/v1/rbac/roles" && method == http.MethodGet:
				rbacHandler.ListRoles(w, r)
			case path == "/api/v1/rbac/roles" && method == http.MethodPost:
				rbacHandler.CreateRole(w, r)
			case path == "/api/v1/rbac/roles/seed" && method == http.MethodPost:
				rbacHandler.SeedDefaultRoles(w, r)
			case matchesPattern(path, "/api/v1/rbac/roles/*/permissions") && method == http.MethodGet:
				rbacHandler.ListRolePermissions(w, r)
			case matchesPattern(path, "/api/v1/rbac/roles/*/permissions") && method == http.MethodPost:
				rbacHandler.AssignRolePermissions(w, r)
			case matchesPattern(path, "/api/v1/rbac/roles/*/permissions/*") && method == http.MethodDelete:
				rbacHandler.RemoveRolePermission(w, r)
			case matchesPattern(path, "/api/v1/rbac/roles/*"):
				if method == http.MethodGet {
					rbacHandler.GetRole(w, r)
				} else if method == http.MethodPut || method == http.MethodPatch {
					rbacHandler.UpdateRole(w, r)
				} else if method == http.MethodDelete {
					rbacHandler.DeleteRole(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Role operation not found")
			}
			return
		}
		// User roles endpoints
		if strings.HasPrefix(path, "/api/v1/rbac/users") {
			switch {
			case matchesPattern(path, "/api/v1/rbac/users/*/roles") && method == http.MethodGet:
				rbacHandler.ListUserRoles(w, r)
			case matchesPattern(path, "/api/v1/rbac/users/*/roles") && method == http.MethodPost:
				rbacHandler.AssignUserRole(w, r)
			case matchesPattern(path, "/api/v1/rbac/users/*/roles") && method == http.MethodDelete:
				rbacHandler.RemoveUserRole(w, r)
			case matchesPattern(path, "/api/v1/rbac/users/*/permissions") && method == http.MethodGet:
				rbacHandler.GetUserPermissions(w, r)
			case matchesPattern(path, "/api/v1/rbac/users/*/check-permission") && method == http.MethodPost:
				rbacHandler.CheckPermission(w, r)
			case matchesPattern(path, "/api/v1/rbac/users/*/audit-logs") && method == http.MethodGet:
				rbacHandler.GetAuditLogs(w, r)
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "User RBAC operation not found")
			}
			return
		}
		// Resource access policies endpoints
		if strings.HasPrefix(path, "/api/v1/rbac/policies") {
			switch {
			case path == "/api/v1/rbac/policies" && method == http.MethodGet:
				rbacHandler.ListResourceAccessPolicies(w, r)
			case path == "/api/v1/rbac/policies" && method == http.MethodPost:
				rbacHandler.CreateResourceAccessPolicy(w, r)
			case matchesPattern(path, "/api/v1/rbac/policies/*"):
				if method == http.MethodPut || method == http.MethodPatch {
					rbacHandler.UpdateResourceAccessPolicy(w, r)
				} else if method == http.MethodDelete {
					rbacHandler.DeleteResourceAccessPolicy(w, r)
				} else {
					respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Policy operation not found")
				}
			default:
				respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Policy operation not found")
			}
			return
		}
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

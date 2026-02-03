// Package k8s provides Kubernetes metrics collection operations
package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsv "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricsClient handles Kubernetes metrics collection
type MetricsClient struct {
	clientset        *kubernetes.Clientset
	metricsClientset *metricsclientset.Clientset
}

// NewMetricsClient creates a new metrics client
func NewMetricsClient(clientset *kubernetes.Clientset, config *rest.Config) (*MetricsClient, error) {
	metricsClientset, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics clientset: %w", err)
	}

	return &MetricsClient{
		clientset:        clientset,
		metricsClientset: metricsClientset,
	}, nil
}

// ClusterMetrics represents cluster-level metrics
type ClusterMetrics struct {
	Timestamp        time.Time
	CPUUsagePercent  float64
	MemoryUsageBytes int64
	MemoryTotalBytes int64
	PodCount         int32
	RunningPodCount  int32
	PendingPodCount  int32
	FailedPodCount   int32
	NodeCount        int32
	ReadyNodeCount   int32
}

// NodeMetricsData represents node metrics with additional info
type NodeMetricsData struct {
	NodeName         string
	Timestamp        time.Time
	CPUUsagePercent  float64
	MemoryUsageBytes int64
	MemoryTotalBytes int64
	DiskUsageBytes   int64
	DiskTotalBytes   int64
	PodCount         int32
	NetworkRxBytes   int64
	NetworkTxBytes   int64
	Status           string
	Ready            bool
}

// PodMetricsData represents pod metrics with additional info
type PodMetricsData struct {
	Namespace        string
	PodName          string
	Timestamp        time.Time
	CPUUsageCores    float64
	MemoryUsageBytes int64
	RestartCount     int32
	Status           string
	Ready            bool
	NodeName         string
}

// GetClusterMetrics collects cluster-level metrics
func (m *MetricsClient) GetClusterMetrics(ctx context.Context) (*ClusterMetrics, error) {
	timestamp := time.Now()

	// Get node metrics
	nodeMetricsList, err := m.metricsClientset.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Get nodes
	nodes, err := m.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get pods
	pods, err := m.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	metrics := &ClusterMetrics{
		Timestamp: timestamp,
		NodeCount: int32(len(nodes.Items)),
		PodCount:  int32(len(pods.Items)),
	}

	var totalCPUUsage, totalCPUCapacity float64
	var totalMemoryUsage, totalMemoryCapacity int64

	// Aggregate node metrics
	for _, nodeMetric := range nodeMetricsList.Items {
		nodeCPU := nodeMetric.Usage.Cpu().MilliValue()
		nodeMemory := nodeMetric.Usage.Memory().Value()

		totalCPUUsage += float64(nodeCPU)
		totalMemoryUsage += nodeMemory
	}

	for _, node := range nodes.Items {
		nodeCPU := node.Status.Capacity.Cpu().MilliValue()
		nodeMemory := node.Status.Capacity.Memory().Value()

		totalCPUCapacity += float64(nodeCPU)
		totalMemoryCapacity += nodeMemory

		// Check node readiness
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					metrics.ReadyNodeCount++
				}
				break
			}
		}
	}

	if totalCPUCapacity > 0 {
		metrics.CPUUsagePercent = (totalCPUUsage / totalCPUCapacity) * 100
	}
	metrics.MemoryUsageBytes = totalMemoryUsage
	metrics.MemoryTotalBytes = totalMemoryCapacity

	// Count pod statuses
	for _, pod := range pods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			metrics.RunningPodCount++
		case corev1.PodPending:
			metrics.PendingPodCount++
		case corev1.PodFailed:
			metrics.FailedPodCount++
		}
	}

	return metrics, nil
}

// GetNodeMetrics collects metrics for all nodes
func (m *MetricsClient) GetNodeMetrics(ctx context.Context) ([]NodeMetricsData, error) {
	timestamp := time.Now()

	// Get node metrics
	nodeMetricsList, err := m.metricsClientset.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Get nodes
	nodes, err := m.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get pods for pod counting
	pods, err := m.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Build node metrics map
	metricsMap := make(map[string]metricsv.NodeMetrics)
	for _, metric := range nodeMetricsList.Items {
		metricsMap[metric.Name] = metric
	}

	// Build node info map
	nodeMap := make(map[string]*corev1.Node)
	for i := range nodes.Items {
		nodeMap[nodes.Items[i].Name] = &nodes.Items[i]
	}

	// Count pods per node
	podCounts := make(map[string]int32)
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			podCounts[pod.Spec.NodeName]++
		}
	}

	var result []NodeMetricsData
	for nodeName, node := range nodeMap {
		metric, hasMetric := metricsMap[nodeName]
		podCount := podCounts[nodeName]

		nodeMetric := NodeMetricsData{
			NodeName:  nodeName,
			Timestamp: timestamp,
			PodCount:  podCount,
			Status:    getNodeStatus(node),
			Ready:     isNodeReady(node),
		}

		if hasMetric {
			cpuUsage := float64(metric.Usage.Cpu().MilliValue())
			memoryUsage := metric.Usage.Memory().Value()
			cpuCapacity := float64(node.Status.Capacity.Cpu().MilliValue())
			memoryCapacity := node.Status.Capacity.Memory().Value()
			storageCapacity := node.Status.Capacity.StorageEphemeral().Value()

			if cpuCapacity > 0 {
				nodeMetric.CPUUsagePercent = (cpuUsage / cpuCapacity) * 100
			}
			nodeMetric.MemoryUsageBytes = memoryUsage
			nodeMetric.MemoryTotalBytes = memoryCapacity
			nodeMetric.DiskTotalBytes = storageCapacity
		}

		result = append(result, nodeMetric)
	}

	return result, nil
}

// GetPodMetrics collects metrics for all pods
func (m *MetricsClient) GetPodMetrics(ctx context.Context, namespace string) ([]PodMetricsData, error) {
	timestamp := time.Now()

	// Get pod metrics
	podMetricsList, err := m.metricsClientset.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Get pods
	pods, err := m.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Build pod metrics map
	metricsMap := make(map[string]metricsv.PodMetrics)
	for _, metric := range podMetricsList.Items {
		key := metric.Namespace + "/" + metric.Name
		metricsMap[key] = metric
	}

	var result []PodMetricsData
	for _, pod := range pods.Items {
		key := pod.Namespace + "/" + pod.Name
		metric, hasMetric := metricsMap[key]

		podMetric := PodMetricsData{
			Namespace:    pod.Namespace,
			PodName:      pod.Name,
			Timestamp:    timestamp,
			RestartCount: getPodRestartCount(&pod),
			Status:       string(pod.Status.Phase),
			Ready:        isPodReady(&pod),
			NodeName:     pod.Spec.NodeName,
		}

		if hasMetric {
			var cpuUsage, memoryUsage int64
			for _, container := range metric.Containers {
				cpuUsage += container.Usage.Cpu().MilliValue()
				memoryUsage += container.Usage.Memory().Value()
			}
			podMetric.CPUUsageCores = float64(cpuUsage) / 1000 // Convert milli-cores to cores
			podMetric.MemoryUsageBytes = memoryUsage
		}

		result = append(result, podMetric)
	}

	return result, nil
}

// Helper functions

func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func getPodRestartCount(pod *corev1.Pod) int32 {
	var count int32
	for _, containerStatus := range pod.Status.ContainerStatuses {
		count += containerStatus.RestartCount
	}
	return count
}

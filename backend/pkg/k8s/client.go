// Package k8s provides Kubernetes client operations
package k8s

import (
	"context"
	"fmt"
	"io"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// ClusterClient represents a Kubernetes cluster client
type ClusterClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// ClusterConfig holds configuration for cluster connection
type ClusterConfig struct {
	Kubeconfig []byte
	Endpoint   string
}

// NewClusterClient creates a new Kubernetes cluster client
func NewClusterClient(config *ClusterConfig) (*ClusterClient, error) {
	var restConfig *rest.Config
	var err error

	if len(config.Kubeconfig) > 0 {
		// Create config from kubeconfig
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(config.Kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create config from kubeconfig: %w", err)
		}
	} else {
		// Use in-cluster config
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	}

	// Override endpoint if provided
	if config.Endpoint != "" {
		restConfig.Host = config.Endpoint
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &ClusterClient{
		clientset: clientset,
		config:    restConfig,
	}, nil
}

// TestConnection tests the connection to the cluster
func (c *ClusterClient) TestConnection(ctx context.Context) (*ConnectionInfo, error) {
	// Get server version
	serverVersion, err := c.clientset.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get node count
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	return &ConnectionInfo{
		Version:   serverVersion.GitVersion,
		NodeCount: int32(len(nodes.Items)),
		Success:   true,
	}, nil
}

// GetNodes retrieves all nodes from the cluster
func (c *ClusterClient) GetNodes(ctx context.Context) ([]NodeInfo, error) {
	nodeList, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodes := make([]NodeInfo, len(nodeList.Items))
	for i, node := range nodeList.Items {
		nodes[i] = NodeInfo{
			Name:             node.Name,
			InternalIP:       getNodeAddress(&node, "InternalIP"),
			ExternalIP:       getNodeAddress(&node, "ExternalIP"),
			Status:           getNodeStatus(&node),
			Roles:            getNodeRoles(&node),
			Version:          node.Status.NodeInfo.KubeletVersion,
			OSImage:          node.Status.NodeInfo.OSImage,
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			CPUCapacity:      node.Status.Capacity.Cpu().String(),
			MemoryCapacity:   node.Status.Capacity.Memory().String(),
			StorageCapacity:  node.Status.Capacity.StorageEphemeral().String(),
			CPUAllocatable:   node.Status.Allocatable.Cpu().String(),
			MemoryAllocatable: node.Status.Allocatable.Memory().String(),
			StorageAllocatable: node.Status.Allocatable.StorageEphemeral().String(),
		}
	}

	return nodes, nil
}

// GetNamespaces retrieves all namespaces from the cluster
func (c *ClusterClient) GetNamespaces(ctx context.Context) ([]string, error) {
	nsList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, len(nsList.Items))
	for i, ns := range nsList.Items {
		namespaces[i] = ns.Name
	}

	return namespaces, nil
}

// GetClusterInfo retrieves general cluster information
func (c *ClusterClient) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	// Get server version
	serverVersion, err := c.clientset.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	// Get node count
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Get namespace count
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	// Get pod count
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Get deployment count
	deployments, err := c.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	// Get service count
	services, err := c.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	// Get ingress count
	ingresses, err := c.clientset.NetworkingV1().Ingresses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	// Get configmap count
	configmaps, err := c.clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	// Get secret count
	secrets, err := c.clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return &ClusterInfo{
		Version:         serverVersion.GitVersion,
		NodeCount:       int32(len(nodes.Items)),
		NamespaceCount:  int32(len(namespaces.Items)),
		PodCount:        int32(len(pods.Items)),
		DeploymentCount: int32(len(deployments.Items)),
		ServiceCount:    int32(len(services.Items)),
		IngressCount:    int32(len(ingresses.Items)),
		ConfigMapCount:  int32(len(configmaps.Items)),
		SecretCount:     int32(len(secrets.Items)),
	}, nil
}

// Close closes the cluster client
func (c *ClusterClient) Close() error {
	// Clientset doesn't need explicit closing
	return nil
}

// GetPods retrieves all pods from a namespace
func (c *ClusterClient) GetPods(ctx context.Context, namespace string) ([]PodInfo, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	pods := make([]PodInfo, len(podList.Items))
	for i, pod := range podList.Items {
		// Get owner references
		var ownerType, ownerName string
		if len(pod.OwnerReferences) > 0 {
			ownerType = string(pod.OwnerReferences[0].Kind)
			ownerName = pod.OwnerReferences[0].Name
		}

		// Get restart count
		restartCount := int32(0)
		for _, cs := range pod.Status.ContainerStatuses {
			restartCount += cs.RestartCount
		}

		// Check if pod is ready
		ready := true
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				ready = false
				break
			}
		}

		pods[i] = PodInfo{
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			Status:       string(pod.Status.Phase),
			Phase:        string(pod.Status.Phase),
			PodIP:        pod.Status.PodIP,
			NodeName:     pod.Spec.NodeName,
			Ready:        ready,
			RestartCount: restartCount,
			OwnerType:    ownerType,
			OwnerName:    ownerName,
			CreatedAt:    pod.CreationTimestamp.Time,
		}
	}

	return pods, nil
}

// GetDeployments retrieves all deployments from a namespace
func (c *ClusterClient) GetDeployments(ctx context.Context, namespace string) ([]DeploymentInfo, error) {
	depList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	deployments := make([]DeploymentInfo, len(depList.Items))
	for i, dep := range depList.Items {
		// Get image from first container
		image := ""
		if len(dep.Spec.Template.Spec.Containers) > 0 {
			image = dep.Spec.Template.Spec.Containers[0].Image
		}

		deployments[i] = DeploymentInfo{
			Name:            dep.Name,
			Namespace:       dep.Namespace,
			Replicas:        *dep.Spec.Replicas,
			ReadyReplicas:   dep.Status.ReadyReplicas,
			UpdatedReplicas: dep.Status.UpdatedReplicas,
			AvailableReplicas: dep.Status.AvailableReplicas,
			Image:           image,
			CreatedAt:       dep.CreationTimestamp.Time,
		}
	}

	return deployments, nil
}

// GetServices retrieves all services from a namespace
func (c *ClusterClient) GetServices(ctx context.Context, namespace string) ([]ServiceInfo, error) {
	svcList, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	services := make([]ServiceInfo, len(svcList.Items))
	for i, svc := range svcList.Items {
		// Convert ports
		ports := make([]ServicePort, len(svc.Spec.Ports))
		for j, port := range svc.Spec.Ports {
			ports[j] = ServicePort{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: string(port.Protocol),
				NodePort: port.NodePort,
			}
		}

		services[i] = ServiceInfo{
			Name:       svc.Name,
			Namespace:  svc.Namespace,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			Ports:      ports,
			CreatedAt:  svc.CreationTimestamp.Time,
		}
	}

	return services, nil
}

// GetPodLogs retrieves logs from a pod
func (c *ClusterClient) GetPodLogs(ctx context.Context, namespace, podName string, tailLines int64) (string, error) {
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{
		TailLines: &tailLines,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer logs.Close()

	buf := make([]byte, 4096)
	var result string
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			result += string(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return result, nil
}

// GetPodLogStream returns a stream for pod logs (for websocket streaming)
func (c *ClusterClient) GetPodLogStream(namespace, podName, containerName string, tailLines int64) io.ReadCloser {
	options := &v1.PodLogOptions{
		Follow:     true,
		TailLines:  &tailLines,
		Timestamps: true,
	}

	if containerName != "" {
		options.Container = containerName
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, options)
	return req.Context(context.Background())
}

// DeletePod deletes a pod
func (c *ClusterClient) DeletePod(ctx context.Context, namespace, podName string) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
}

// GetPodDetail retrieves detailed information about a specific pod
func (c *ClusterClient) GetPodDetail(ctx context.Context, namespace, podName string) (*PodDetailInfo, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Get owner references
	var ownerType, ownerName string
	if len(pod.OwnerReferences) > 0 {
		ownerType = string(pod.OwnerReferences[0].Kind)
		ownerName = pod.OwnerReferences[0].Name
	}

	// Process containers
	containers := make([]ContainerInfo, len(pod.Spec.Containers))
	for i, c := range pod.Spec.Containers {
		// Find matching status
		var status *v1.ContainerStatus
		for j, cs := range pod.Status.ContainerStatuses {
			if cs.Name == c.Name {
				status = &pod.Status.ContainerStatuses[j]
				break
			}
		}

		container := ContainerInfo{
			Name:  c.Name,
			Image: c.Image,
			ImagePullPolicy: string(c.ImagePullPolicy),
		}

		if status != nil {
			container.Ready = status.Ready
			container.RestartCount = status.RestartCount
			container.State = getContainerState(status)
			container.Ready = status.Ready
		}

		// Add resource requests
		if c.Resources.Requests != nil {
			container.CPURequest = c.Resources.Requests.Cpu().String()
			container.MemoryRequest = c.Resources.Requests.Memory().String()
		}
		if c.Resources.Limits != nil {
			container.CPULimit = c.Resources.Limits.Cpu().String()
			container.MemoryLimit = c.Resources.Limits.Memory().String()
		}

		containers[i] = container
	}

	// Get events
	events, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	podEvents := make([]EventInfo, 0)
	if err == nil {
		for _, e := range events.Items {
			podEvents = append(podEvents, EventInfo{
				Type:      e.Type,
				Reason:    e.Reason,
				Message:   e.Message,
				FirstSeen: e.FirstTimestamp.Time,
				LastSeen:  e.LastTimestamp.Time,
				Count:     e.Count,
			})
		}
	}

	return &PodDetailInfo{
		Name:              pod.Name,
		Namespace:         pod.Namespace,
		Status:            string(pod.Status.Phase),
		Phase:             string(pod.Status.Phase),
		PodIP:             pod.Status.PodIP,
		HostIP:            pod.Status.HostIP,
		NodeName:          pod.Spec.NodeName,
		Ready:             isPodReady(pod),
		RestartCount:      getPodRestartCount(pod),
		OwnerType:         ownerType,
		OwnerName:         ownerName,
		Containers:        containers,
		Events:            podEvents,
		Labels:            pod.Labels,
		Annotations:       pod.Annotations,
		ServiceAccount:    pod.Spec.ServiceAccountName,
		RestartPolicy:     string(pod.Spec.RestartPolicy),
		DNSPolicy:         string(pod.Spec.DNSPolicy),
		CreatedAt:         pod.CreationTimestamp.Time,
		StartTime:         getTimePtr(pod.Status.StartTime),
		QOSClass:          string(pod.Status.QOSClass),
	}, nil
}

// ExecConfig holds configuration for pod exec
type ExecConfig struct {
	Namespace    string
	PodName      string
	Container    string
	Command      []string
	TTY          bool
	Stdin        bool
	Stdout       bool
	Stderr       bool
}

// PodExec creates an interactive exec session in a pod container
func (c *ClusterClient) PodExec(ctx context.Context, config *ExecConfig) (remotecommand.Executor, error) {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(config.PodName).
		Namespace(config.Namespace).
		SubResource("exec").
		Param("container", config.Container)

	for _, cmd := range config.Command {
		req.Param("command", cmd)
	}

	req.Param("stdin", fmt.Sprintf("%v", config.Stdin))
	req.Param("stdout", fmt.Sprintf("%v", config.Stdout))
	req.Param("stderr", fmt.Sprintf("%v", config.Stderr))
	req.Param("tty", fmt.Sprintf("%v", config.TTY))

	executor, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	return executor, nil
}

// getContainerState returns the current state of a container
func getContainerState(status *v1.ContainerStatus) string {
	if status.State.Running != nil {
		return "Running"
	}
	if status.State.Waiting != nil {
		return "Waiting: " + status.State.Waiting.Reason
	}
	if status.State.Terminated != nil {
		return "Terminated: " + status.State.Terminated.Reason
	}
	return "Unknown"
}

// isPodReady checks if all containers in the pod are ready
func isPodReady(pod *v1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if !cs.Ready {
			return false
		}
	}
	return len(pod.Status.ContainerStatuses) > 0
}

// getPodRestartCount returns total restart count for all containers
func getPodRestartCount(pod *v1.Pod) int32 {
	var count int32
	for _, cs := range pod.Status.ContainerStatuses {
		count += cs.RestartCount
	}
	return count
}

// getTimePtr converts a metav1.Time to *time.Time
func getTimePtr(t *metav1.Time) *time.Time {
	if t == nil {
		return nil
	}
	return &t.Time
}

// Helper types and functions

type ConnectionInfo struct {
	Success   bool   `json:"success"`
	Version   string `json:"version"`
	NodeCount int32  `json:"nodeCount"`
	Error     string `json:"error,omitempty"`
}

type NodeInfo struct {
	Name              string `json:"name"`
	InternalIP        string `json:"internalIp"`
	ExternalIP        string `json:"externalIp"`
	Status            string `json:"status"`
	Roles             string `json:"roles"`
	Version           string `json:"version"`
	OSImage           string `json:"osImage"`
	KernelVersion     string `json:"kernelVersion"`
	ContainerRuntime  string `json:"containerRuntime"`
	CPUCapacity       string `json:"cpuCapacity"`
	MemoryCapacity    string `json:"memoryCapacity"`
	StorageCapacity   string `json:"storageCapacity"`
	CPUAllocatable    string `json:"cpuAllocatable"`
	MemoryAllocatable string `json:"memoryAllocatable"`
	StorageAllocatable string `json:"storageAllocatable"`
}

type ClusterInfo struct {
	Version         string `json:"version"`
	NodeCount       int32  `json:"nodeCount"`
	NamespaceCount  int32  `json:"namespaceCount"`
	PodCount        int32  `json:"podCount"`
	DeploymentCount int32  `json:"deploymentCount"`
	ServiceCount    int32  `json:"serviceCount"`
	IngressCount    int32  `json:"ingressCount"`
	ConfigMapCount  int32  `json:"configMapCount"`
	SecretCount     int32  `json:"secretCount"`
}

type PodInfo struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Status       string    `json:"status"`
	Phase        string    `json:"phase"`
	PodIP        string    `json:"podIp"`
	NodeName     string    `json:"nodeName"`
	Ready        bool      `json:"ready"`
	RestartCount int32     `json:"restartCount"`
	OwnerType    string    `json:"ownerType,omitempty"`
	OwnerName    string    `json:"ownerName,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type DeploymentInfo struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	Replicas          int32     `json:"replicas"`
	ReadyReplicas     int32     `json:"readyReplicas"`
	UpdatedReplicas   int32     `json:"updatedReplicas"`
	AvailableReplicas int32     `json:"availableReplicas"`
	Image             string    `json:"image"`
	CreatedAt         time.Time `json:"createdAt"`
}

type ServiceInfo struct {
	Name      string       `json:"name"`
	Namespace string       `json:"namespace"`
	Type      string       `json:"type"`
	ClusterIP string       `json:"clusterIp"`
	Ports     []ServicePort `json:"ports"`
	CreatedAt time.Time    `json:"createdAt"`
}

type ServicePort struct {
	Name     string `json:"name,omitempty"`
	Port     int32  `json:"port"`
	Protocol string `json:"protocol"`
	NodePort int32  `json:"nodePort,omitempty"`
}

type PodDetailInfo struct {
	Name           string            `json:"name"`
	Namespace      string            `json:"namespace"`
	Status         string            `json:"status"`
	Phase          string            `json:"phase"`
	PodIP          string            `json:"podIp"`
	HostIP         string            `json:"hostIp"`
	NodeName       string            `json:"nodeName"`
	Ready          bool              `json:"ready"`
	RestartCount   int32             `json:"restartCount"`
	OwnerType      string            `json:"ownerType,omitempty"`
	OwnerName      string            `json:"ownerName,omitempty"`
	Containers     []ContainerInfo   `json:"containers"`
	Events         []EventInfo       `json:"events"`
	Labels         map[string]string `json:"labels"`
	Annotations    map[string]string `json:"annotations"`
	ServiceAccount string            `json:"serviceAccount"`
	RestartPolicy  string            `json:"restartPolicy"`
	DNSPolicy      string            `json:"dnsPolicy"`
	CreatedAt      time.Time         `json:"createdAt"`
	StartTime      *time.Time        `json:"startTime,omitempty"`
	QOSClass       string            `json:"qosClass"`
}

type ContainerInfo struct {
	Name           string `json:"name"`
	Image          string `json:"image"`
	ImagePullPolicy string `json:"imagePullPolicy"`
	Ready          bool   `json:"ready"`
	RestartCount   int32  `json:"restartCount"`
	State          string `json:"state"`
	CPURequest     string `json:"cpuRequest,omitempty"`
	MemoryRequest  string `json:"memoryRequest,omitempty"`
	CPULimit       string `json:"cpuLimit,omitempty"`
	MemoryLimit    string `json:"memoryLimit,omitempty"`
}

type EventInfo struct {
	Type      string    `json:"type"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	FirstSeen time.Time `json:"firstSeen"`
	LastSeen  time.Time `json:"lastSeen"`
	Count     int32     `json:"count"`
}

func getNodeAddress(node *v1.Node, addressType string) string {
	for _, addr := range node.Status.Addresses {
		if string(addr.Type) == addressType {
			return addr.Address
		}
	}
	return ""
}

func getNodeStatus(node *v1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			if condition.Status == v1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

func getNodeRoles(node *v1.Node) string {
	roles := []string{}
	for label := range node.Labels {
		if label == "node-role.kubernetes.io/master" || label == "node-role.kubernetes.io/control-plane" {
			roles = append(roles, "master")
		} else if label == "node-role.kubernetes.io/worker" {
			roles = append(roles, "worker")
		}
	}
	if len(roles) == 0 {
		return "worker"
	}
	return roles[0]
}

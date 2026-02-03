// Package k8s provides Kubernetes client operations
package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

// Package handler provides websocket handlers for real-time operations
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/wangjialin/myops/pkg/k8s"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

// PodLogsWebSocketHandler handles websocket connections for pod log streaming
type PodLogsWebSocketHandler struct {
	db *gorm.DB
}

// NewPodLogsWebSocketHandler creates a new pod logs websocket handler
func NewPodLogsWebSocketHandler(db *gorm.DB) *PodLogsWebSocketHandler {
	return &PodLogsWebSocketHandler{db: db}
}

// ServeHTTP handles websocket upgrade for pod logs
func (h *PodLogsWebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Get cluster ID, namespace, and pod name from URL query
	clusterID := r.URL.Query().Get("clusterId")
	namespace := r.URL.Query().Get("namespace")
	podName := r.URL.Query().Get("podName")
	containerName := r.URL.Query().Get("container")
	tailLinesStr := r.URL.Query().Get("tailLines")
	follow := r.URL.Query().Get("follow") == "true"

	if clusterID == "" || namespace == "" || podName == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Missing required parameters"))
		return
	}

	// Parse cluster ID
	clusterUUID, err := uuid.Parse(clusterID)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid cluster ID"))
		return
	}

	// Get user ID from context (if available)
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: User not authenticated"))
		return
	}

	// Verify cluster ownership
	var cluster model.K8sCluster
	if err := h.db.Where("id = ? AND user_id = ?", clusterUUID, userID).First(&cluster).Error; err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Cluster not found"))
		return
	}

	// Parse tail lines
	tailLines := int64(100)
	if tailLinesStr != "" {
		if tl, err := strconv.ParseInt(tailLinesStr, 10, 64); err == nil && tl > 0 {
			tailLines = tl
		}
	}

	// Create cluster client
	config := &k8s.ClusterConfig{
		Kubeconfig: []byte(cluster.Kubeconfig),
		Endpoint:   cluster.Endpoint,
	}

	client, err := k8s.NewClusterClient(config)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Failed to create cluster client"))
		return
	}
	defer client.Close()

	// Stream logs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle ping/pong for connection keepalive
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					cancel()
					return
				}
			}
		}
	}()

	if follow {
		// Follow mode - stream logs continuously
		if err := streamPodLogsFollow(ctx, conn, client, namespace, podName, containerName, tailLines); err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: %v", err)))
		}
	} else {
		// Static mode - get logs once
		logs, err := client.GetPodLogs(ctx, namespace, podName, tailLines)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error: %v", err)))
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte(logs))
	}
}

// streamPodLogsFollow streams pod logs in follow mode
func streamPodLogsFollow(ctx context.Context, conn *websocket.Conn, client *k8s.ClusterClient, namespace, podName, containerName string, tailLines int64) error {
	req := client.GetPodLogStream(namespace, podName, containerName, tailLines)
	if req == nil {
		return fmt.Errorf("failed to create log stream request")
	}
	defer req.Close()

	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to stream logs: %w", err)
	}
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, err := stream.Read(buf)
			if n > 0 {
				if err := conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
					return err
				}
			}
			if err != nil {
				// End of stream
				return nil
			}
		}
	}
}

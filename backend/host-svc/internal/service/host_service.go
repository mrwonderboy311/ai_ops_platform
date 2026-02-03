// Package service provides business logic for host management
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/db"
	"github.com/wangjialin/myops/pkg/model"
)

var (
	ErrHostNotFound     = errors.New("host not found")
	ErrHostAlreadyExists = errors.New("host with this IP:port already exists")
)

// HostService handles host business logic
type HostService struct {
	hostRepo *db.HostRepository
}

// NewHostService creates a new HostService
func NewHostService(hostRepo *db.HostRepository) *HostService {
	return &HostService{
		hostRepo: hostRepo,
	}
}

// CreateHost creates a new host
func (s *HostService) CreateHost(ctx context.Context, req *CreateHostRequest, registeredBy uuid.UUID) (*model.Host, error) {
	// Check if host already exists
	exists, err := s.hostRepo.ExistsIPAddress(ctx, req.IPAddress, req.Port)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrHostAlreadyExists
	}

	host := &model.Host{
		ID:          uuid.New(),
		Hostname:    req.Hostname,
		IPAddress:   req.IPAddress,
		Port:        req.Port,
		Status:      model.HostStatusPending,
		OSType:      req.OSType,
		OSVersion:   req.OSVersion,
		CPUCores:    req.CPUCores,
		MemoryGB:    req.MemoryGB,
		DiskGB:      req.DiskGB,
		Labels:      req.Labels,
		Tags:        req.Tags,
		RegisteredBy: &registeredBy,
	}

	if req.ClusterID != nil && *req.ClusterID != "" {
		clusterID, err := uuid.Parse(*req.ClusterID)
		if err != nil {
			return nil, fmt.Errorf("invalid cluster_id: %w", err)
		}
		host.ClusterID = &clusterID
	}

	if err := s.hostRepo.Create(ctx, host); err != nil {
		return nil, err
	}

	return host, nil
}

// GetHost retrieves a host by ID
func (s *HostService) GetHost(ctx context.Context, id uuid.UUID) (*model.Host, error) {
	host, err := s.hostRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return host, nil
}

// ListHosts returns a paginated list of hosts
func (s *HostService) ListHosts(ctx context.Context, filter *db.HostFilter) ([]*model.Host, int64, error) {
	return s.hostRepo.List(ctx, filter)
}

// UpdateHost updates an existing host
func (s *HostService) UpdateHost(ctx context.Context, id uuid.UUID, req *UpdateHostRequest) (*model.Host, error) {
	host, err := s.hostRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Hostname != "" {
		host.Hostname = req.Hostname
	}
	if req.Port != 0 {
		host.Port = req.Port
	}
	if req.OSType != "" {
		host.OSType = req.OSType
	}
	if req.OSVersion != "" {
		host.OSVersion = req.OSVersion
	}
	if req.CPUCores != nil {
		host.CPUCores = req.CPUCores
	}
	if req.MemoryGB != nil {
		host.MemoryGB = req.MemoryGB
	}
	if req.DiskGB != nil {
		host.DiskGB = req.DiskGB
	}
	if req.Labels != nil {
		host.Labels = req.Labels
	}
	if req.Tags != nil {
		host.Tags = req.Tags
	}
	if req.ClusterID != nil {
		if *req.ClusterID == "" {
			host.ClusterID = nil
		} else {
			clusterID, err := uuid.Parse(*req.ClusterID)
			if err != nil {
				return nil, fmt.Errorf("invalid cluster_id: %w", err)
			}
			host.ClusterID = &clusterID
		}
	}

	if err := s.hostRepo.Update(ctx, host); err != nil {
		return nil, err
	}

	return host, nil
}

// DeleteHost deletes a host
func (s *HostService) DeleteHost(ctx context.Context, id uuid.UUID) error {
	return s.hostRepo.Delete(ctx, id)
}

// ApproveHost approves a pending host registration
func (s *HostService) ApproveHost(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) (*model.Host, error) {
	host, err := s.hostRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if already approved
	if host.Status == model.HostStatusApproved {
		return host, nil
	}

	if err := s.hostRepo.Approve(ctx, id, approvedBy); err != nil {
		return nil, err
	}

	// Fetch updated host
	return s.hostRepo.FindByID(ctx, id)
}

// CreateHostRequest represents a create host request
type CreateHostRequest struct {
	Hostname    string              `json:"hostname"`
	IPAddress   string              `json:"ipAddress"`
	Port        int                 `json:"port"`
	OSType      string              `json:"osType"`
	OSVersion   string              `json:"osVersion"`
	CPUCores    *int                `json:"cpuCores"`
	MemoryGB    *int                `json:"memoryGB"`
	DiskGB      *int64              `json:"diskGB"`
	Labels      map[string]string   `json:"labels"`
	Tags        []string            `json:"tags"`
	ClusterID   *string             `json:"clusterId"`
}

// UpdateHostRequest represents an update host request
type UpdateHostRequest struct {
	Hostname  string            `json:"hostname"`
	Port      int               `json:"port"`
	OSType    string            `json:"osType"`
	OSVersion string            `json:"osVersion"`
	CPUCores  *int              `json:"cpuCores"`
	MemoryGB  *int              `json:"memoryGB"`
	DiskGB    *int64            `json:"diskGB"`
	Labels    map[string]string `json:"labels"`
	Tags      []string          `json:"tags"`
	ClusterID *string           `json:"clusterId"`
}

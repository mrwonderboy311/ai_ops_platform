// Package db provides database operations
package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/wangjialin/myops/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HostRepository handles host database operations
type HostRepository struct {
	db *gorm.DB
}

// NewHostRepository creates a new HostRepository
func NewHostRepository(db *gorm.DB) *HostRepository {
	return &HostRepository{db: db}
}

// Create creates a new host
func (r *HostRepository) Create(ctx context.Context, host *model.Host) error {
	return r.db.WithContext(ctx).Create(host).Error
}

// FindByID finds a host by ID
func (r *HostRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Host, error) {
	var host model.Host
	err := r.db.WithContext(ctx).
		Preload("RegisteredByUser").
		Preload("ApprovedByUser").
		Preload("Cluster").
		Where("id = ?", id).
		First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// FindByIPAddressAndPort finds a host by IP address and port
func (r *HostRepository) FindByIPAddressAndPort(ctx context.Context, ipAddress string, port int) (*model.Host, error) {
	var host model.Host
	err := r.db.WithContext(ctx).
		Where("ip_address = ? AND port = ?", ipAddress, port).
		First(&host).Error
	if err != nil {
		return nil, err
	}
	return &host, nil
}

// ExistsIPAddress checks if an IP address:port combination already exists
func (r *HostRepository) ExistsIPAddress(ctx context.Context, ipAddress string, port int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Host{}).
		Where("ip_address = ? AND port = ?", ipAddress, port).
		Count(&count).Error
	return count > 0, err
}

// List returns a paginated list of hosts with optional filters
func (r *HostRepository) List(ctx context.Context, filter *HostFilter) ([]*model.Host, int64, error) {
	var hosts []*model.Host
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Host{})

	// Apply filters
	if filter != nil {
		if filter.Status != "" {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.Hostname != "" {
			query = query.Where("hostname ILIKE ?", "%"+filter.Hostname+"%")
		}
		if filter.IPAddress != "" {
			query = query.Where("ip_address ILIKE ?", "%"+filter.IPAddress+"%")
		}
		if filter.RegisteredBy != nil {
			query = query.Where("registered_by = ?", filter.RegisteredBy)
		}
		if filter.Labels != nil && len(filter.Labels) > 0 {
			for key, value := range filter.Labels {
				query = query.Where("labels->? = ?", key, value)
			}
		}
		if len(filter.Tags) > 0 {
			query = query.Where("tags && ?", pq.Array(filter.Tags))
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter != nil {
		if filter.PageSize > 0 {
			query = query.Limit(filter.PageSize)
		}
		if filter.Page > 0 {
			offset := (filter.Page - 1) * filter.PageSize
			query = query.Offset(offset)
		}
		// Apply sorting
		if filter.SortBy != "" {
			order := filter.SortBy
			if filter.SortDesc {
				order += " DESC"
			}
			query = query.Order(order)
		}
	}

	// Eager load associations
	query = query.Preload("RegisteredByUser").Preload("ApprovedByUser").Preload("Cluster")

	err := query.Find(&hosts).Error
	if err != nil {
		return nil, 0, err
	}

	return hosts, total, nil
}

// Update updates an existing host
func (r *HostRepository) Update(ctx context.Context, host *model.Host) error {
	return r.db.WithContext(ctx).Save(host).Error
}

// Delete deletes a host
func (r *HostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Host{}, "id = ?", id).Error
}

// UpdateStatus updates the status of a host
func (r *HostRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.HostStatus) error {
	return r.db.WithContext(ctx).
		Model(&model.Host{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Approve approves a host registration
func (r *HostRepository) Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Host{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.HostStatusApproved,
			"approved_by": approvedBy,
			"approved_at": clause.Expr{SQL: "NOW()"},
		}).Error
}

// HostFilter represents filter options for listing hosts
type HostFilter struct {
	Page        int                 `json:"page"`
	PageSize    int                 `json:"pageSize"`
	Status      model.HostStatus    `json:"status"`
	Hostname    string              `json:"hostname"`
	IPAddress   string              `json:"ipAddress"`
	RegisteredBy *uuid.UUID         `json:"registeredBy"`
	Labels      map[string]string   `json:"labels"`
	Tags        []string            `json:"tags"`
	SortBy      string              `json:"sortBy"`
	SortDesc    bool                `json:"sortDesc"`
}

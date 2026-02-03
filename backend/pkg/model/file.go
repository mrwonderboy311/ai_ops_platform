// Package model provides data models for file operations
package model

import (
	"time"

	"github.com/google/uuid"
)

// FileTransferStatus represents the status of a file transfer
type FileTransferStatus string

const (
	FileTransferStatusPending   FileTransferStatus = "pending"
	FileTransferStatusRunning   FileTransferStatus = "running"
	FileTransferStatusCompleted FileTransferStatus = "completed"
	FileTransferStatusFailed    FileTransferStatus = "failed"
	FileTransferStatusCancelled FileTransferStatus = "cancelled"
)

// FileTransferDirection represents the direction of file transfer
type FileTransferDirection string

const (
	FileTransferDirectionUpload   FileTransferDirection = "upload"
	FileTransferDirectionDownload FileTransferDirection = "download"
)

// FileTransfer represents a file transfer operation
type FileTransfer struct {
	ID           uuid.UUID             `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	HostID       uuid.UUID             `json:"hostId" gorm:"type:uuid;not null;index"`
	Host         *Host                 `json:"host,omitempty" gorm:"foreignKey:HostID"`
	UserID       uuid.UUID             `json:"userId" gorm:"type:uuid;not null;index"`
	User         *User                 `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Direction    FileTransferDirection  `json:"direction" gorm:"type:varchar(20);not null"`
	SourcePath   string                `json:"sourcePath" gorm:"type:varchar(512);not null"`
	TargetPath   string                `json:"targetPath" gorm:"type:varchar(512);not null"`
	FileName     string                `json:"fileName" gorm:"type:varchar(256);not null"`
	FileSize     int64                 `json:"fileSize" gorm:"bigint;default:0"`
	Transferred  int64                 `json:"transferred" gorm:"bigint;default:0"`
	Status       FileTransferStatus    `json:"status" gorm:"type:varchar(20);not null;index"`
	ErrorMessage string                `json:"errorMessage,omitempty" gorm:"type:text"`
	StartedAt    *time.Time            `json:"startedAt,omitempty" gorm:"index"`
	CompletedAt  *time.Time            `json:"completedAt,omitempty"`
	CreatedAt    time.Time             `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time             `json:"updatedAt" gorm:"autoUpdateTime"`
}

// FileInfo represents information about a remote file
type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	Mode         string    `json:"mode"`
	ModTime      time.Time `json:"modTime"`
	IsDir        bool      `json:"isDir"`
	Owner        string    `json:"owner,omitempty"`
	Group        string    `json:"group,omitempty"`
	Permissions  string    `json:"permissions"`
}

// ListDirectoryRequest represents a request to list a directory
type ListDirectoryRequest struct {
	HostID   uuid.UUID `json:"hostId" binding:"required"`
	Path     string    `json:"path" binding:"required"`
	Username string    `json:"username"`
	Password string    `json:"password,omitempty"`
	Key      string    `json:"key,omitempty"`
}

// ListDirectoryResponse represents the response of listing a directory
type ListDirectoryResponse struct {
	Path      string     `json:"path"`
	Parent    string     `json:"parent,omitempty"`
	Files     []FileInfo `json:"files"`
	TotalSize int64      `json:"totalSize"`
	FileCount int        `json:"fileCount"`
	DirCount  int        `json:"dirCount"`
}

// FileUploadRequest represents a request to upload a file
type FileUploadRequest struct {
	HostID       uuid.UUID `json:"hostId" binding:"required"`
	RemotePath   string    `json:"remotePath" binding:"required"`
	Username     string    `json:"username"`
	Password     string    `json:"password,omitempty"`
	Key          string    `json:"key,omitempty"`
	Overwrite    bool      `json:"overwrite"`
}

// FileDownloadRequest represents a request to download a file
type FileDownloadRequest struct {
	HostID   uuid.UUID `json:"hostId" binding:"required"`
	RemotePath string  `json:"remotePath" binding:"required"`
	Username string    `json:"username"`
	Password string    `json:"password,omitempty"`
	Key      string    `json:"key,omitempty"`
}

// FileDeleteRequest represents a request to delete a file
type FileDeleteRequest struct {
	HostID   uuid.UUID `json:"hostId" binding:"required"`
	RemotePath string  `json:"remotePath" binding:"required"`
	Username string    `json:"username"`
	Password string    `json:"password,omitempty"`
	Key      string    `json:"key,omitempty"`
}

// FileRenameRequest represents a request to rename a file
type FileRenameRequest struct {
	HostID      uuid.UUID `json:"hostId" binding:"required"`
	OldPath     string    `json:"oldPath" binding:"required"`
	NewPath     string    `json:"newPath" binding:"required"`
	Username    string    `json:"username"`
	Password    string    `json:"password,omitempty"`
	Key         string    `json:"key,omitempty"`
}

// CreateDirectoryRequest represents a request to create a directory
type CreateDirectoryRequest struct {
	HostID   uuid.UUID `json:"hostId" binding:"required"`
	Path     string    `json:"path" binding:"required"`
	Mode     string    `json:"mode,omitempty"`
	Username string    `json:"username"`
	Password string    `json:"password,omitempty"`
	Key      string    `json:"key,omitempty"`
}

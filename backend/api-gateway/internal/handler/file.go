// Package handler provides HTTP handlers for file transfer operations
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wangjialin/myops/pkg/model"
	"github.com/wangjialin/myops/pkg/ssh"
	"gorm.io/gorm"
)

// FileTransferHandler handles file transfer operations
type FileTransferHandler struct {
	db *gorm.DB
}

// NewFileTransferHandler creates a new file transfer handler
func NewFileTransferHandler(db *gorm.DB) *FileTransferHandler {
	return &FileTransferHandler{db: db}
}

// ListDirectory handles directory listing requests
func (h *FileTransferHandler) ListDirectory(w http.ResponseWriter, r *http.Request) {
	var req model.ListDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify host exists and user has permission
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Validate and clean path
	cleanPath, err := ssh.ValidatePath(req.Path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_PATH", err.Error())
		return
	}

	// Create SFTP client
	config := &ssh.SFTPConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewSFTPClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// List files
	files, err := client.ListFiles(cleanPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "LIST_FAILED", fmt.Sprintf("Failed to list directory: %v", err))
		return
	}

	// Calculate totals
	var totalSize int64
	var fileCount, dirCount int
	for _, file := range files {
		totalSize += file.Size
		if file.IsDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	// Get parent directory
	parent := ""
	if cleanPath != "/" && cleanPath != "." {
		parent = filepath.Dir(cleanPath)
	}

	response := model.ListDirectoryResponse{
		Path:      cleanPath,
		Parent:    parent,
		Files:     files,
		TotalSize: totalSize,
		FileCount: fileCount,
		DirCount:  dirCount,
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

// UploadFile handles file upload requests
func (h *FileTransferHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		respondWithError(w, http.StatusBadRequest, "INVALID_FORM", "Failed to parse form")
		return
	}

	// Get form values
	hostIDStr := r.FormValue("hostId")
	remotePath := r.FormValue("remotePath")
	username := r.FormValue("username")
	password := r.FormValue("password")
	key := r.FormValue("key")
	overwriteStr := r.FormValue("overwrite")

	if hostIDStr == "" || remotePath == "" {
		respondWithError(w, http.StatusBadRequest, "MISSING_FIELDS", "hostId and remotePath are required")
		return
	}

	hostID, err := uuid.Parse(hostIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_HOST_ID", "Invalid host ID")
		return
	}

	_ = overwriteStr == "true" // TODO: implement overwrite logic

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify host exists
	var host model.Host
	err = h.db.Where("id = ?", hostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "NO_FILE", "No file uploaded")
		return
	}
	defer file.Close()

	// Create file transfer record
	transferID := uuid.New()
	transfer := &model.FileTransfer{
		ID:          transferID,
		HostID:      hostID,
		UserID:      userID,
		Direction:   model.FileTransferDirectionUpload,
		SourcePath:  header.Filename,
		TargetPath:  remotePath,
		FileName:    header.Filename,
		FileSize:    header.Size,
		Status:      model.FileTransferStatusRunning,
		StartedAt:   timePtr(time.Now()),
	}

	if err := h.db.Create(transfer).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create transfer record")
		return
	}

	// Create SFTP client
	config := &ssh.SFTPConfig{
		HostID:     hostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   username,
		Password:   password,
		PrivateKey: []byte(key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewSFTPClient(config)
	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Upload file
	progress := make(chan int64)
	go func() {
		for transferred := range progress {
			h.updateTransferProgress(transferID, transferred)
		}
	}()

	// Create temporary file for the upload
	tempFile, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "TEMP_FILE_ERROR", "Failed to create temp file")
		return
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Copy uploaded content to temp file
	_, err = io.Copy(tempFile, file)
	tempFile.Close()
	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "FILE_READ_ERROR", "Failed to read uploaded file")
		return
	}

	// Upload via SFTP
	targetPath := filepath.Join(remotePath, header.Filename)
	transferred, err := client.UploadFile(tempPath, targetPath, progress)
	close(progress)

	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "UPLOAD_FAILED", fmt.Sprintf("Failed to upload: %v", err))
		return
	}

	// Update transfer record as completed
	now := time.Now()
	h.db.Model(&model.FileTransfer{}).Where("id = ?", transferID).Updates(map[string]interface{}{
		"status":      model.FileTransferStatusCompleted,
		"transferred": transferred,
		"completed_at": &now,
	})

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"transferId": transferID,
			"fileName":   header.Filename,
			"size":       transferred,
			"targetPath": targetPath,
		},
	})
}

// DownloadFile handles file download requests
func (h *FileTransferHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	var req model.FileDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Create SFTP client
	config := &ssh.SFTPConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewSFTPClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Get file info
	fileInfo, err := client.GetFileInfo(req.RemotePath)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "FILE_NOT_FOUND", fmt.Sprintf("Failed to get file info: %v", err))
		return
	}

	// Create file transfer record
	transferID := uuid.New()
	transfer := &model.FileTransfer{
		ID:          transferID,
		HostID:      req.HostID,
		UserID:      userID,
		Direction:   model.FileTransferDirectionDownload,
		SourcePath:  req.RemotePath,
		TargetPath:  fileInfo.Name,
		FileName:    fileInfo.Name,
		FileSize:    fileInfo.Size,
		Status:      model.FileTransferStatusRunning,
		StartedAt:   timePtr(time.Now()),
	}

	if err := h.db.Create(transfer).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create transfer record")
		return
	}

	// Create temp file for download
	tempFile, err := os.CreateTemp("", "download-*.tmp")
	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "TEMP_FILE_ERROR", "Failed to create temp file")
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// Download file via SFTP
	progress := make(chan int64)
	go func() {
		for transferred := range progress {
			h.updateTransferProgress(transferID, transferred)
		}
	}()

	transferred, err := client.DownloadFile(req.RemotePath, tempPath, progress)
	close(progress)

	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "DOWNLOAD_FAILED", fmt.Sprintf("Failed to download: %v", err))
		return
	}

	// Read downloaded file and send to client
	downloadedFile, err := os.ReadFile(tempPath)
	if err != nil {
		h.updateTransferStatus(transferID, model.FileTransferStatusFailed, err.Error())
		respondWithError(w, http.StatusInternalServerError, "FILE_READ_ERROR", "Failed to read downloaded file")
		return
	}

	// Update transfer record as completed
	now := time.Now()
	h.db.Model(&model.FileTransfer{}).Where("id = ?", transferID).Updates(map[string]interface{}{
		"status":       model.FileTransferStatusCompleted,
		"transferred":  transferred,
		"completed_at": &now,
	})

	// Set response headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name))
	w.Header().Set("Content-Length", strconv.FormatInt(transferred, 10))

	w.Write(downloadedFile)
}

// DeleteFile handles file deletion requests
func (h *FileTransferHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	var req model.FileDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Create SFTP client
	config := &ssh.SFTPConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewSFTPClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Delete file
	if err := client.DeleteFile(req.RemotePath); err != nil {
		respondWithError(w, http.StatusInternalServerError, "DELETE_FAILED", fmt.Sprintf("Failed to delete file: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "File deleted successfully",
		"path":    req.RemotePath,
	})
}

// CreateDirectory handles directory creation requests
func (h *FileTransferHandler) CreateDirectory(w http.ResponseWriter, r *http.Request) {
	var req model.CreateDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Verify host exists
	var host model.Host
	err := h.db.Where("id = ?", req.HostID).First(&host).Error
	if err == gorm.ErrRecordNotFound {
		respondWithError(w, http.StatusNotFound, "NOT_FOUND", "Host not found")
		return
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	// Check if host is available
	if host.Status != model.HostStatusApproved && host.Status != model.HostStatusOnline {
		respondWithError(w, http.StatusForbidden, "HOST_NOT_AVAILABLE", "Host is not available")
		return
	}

	// Create SFTP client
	config := &ssh.SFTPConfig{
		HostID:     req.HostID.String(),
		IPAddress:  host.IPAddress,
		Port:       host.Port,
		Username:   req.Username,
		Password:   req.Password,
		PrivateKey: []byte(req.Key),
		Timeout:    30 * time.Second,
	}

	client, err := ssh.NewSFTPClient(config)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "CONNECTION_FAILED", fmt.Sprintf("Failed to connect: %v", err))
		return
	}
	defer client.Close()

	// Parse mode
	mode := ssh.ParseMode(req.Mode)

	// Create directory
	if err := client.CreateDirectory(req.Path, mode); err != nil {
		respondWithError(w, http.StatusInternalServerError, "CREATE_FAILED", fmt.Sprintf("Failed to create directory: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Directory created successfully",
		"path":    req.Path,
	})
}

// GetTransfers retrieves file transfer history for a user
func (h *FileTransferHandler) GetTransfers(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	var userID uuid.UUID
	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		if uid, ok := userIDVal.(string); ok {
			userID, _ = uuid.Parse(uid)
		}
	}

	if userID == (uuid.UUID{}) {
		respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse query parameters
	hostIDStr := r.URL.Query().Get("hostId")
	direction := r.URL.Query().Get("direction")
	status := r.URL.Query().Get("status")

	query := h.db.Model(&model.FileTransfer{}).Where("user_id = ?", userID)

	if hostIDStr != "" {
		hostID, err := uuid.Parse(hostIDStr)
		if err == nil {
			query = query.Where("host_id = ?", hostID)
		}
	}

	if direction != "" {
		query = query.Where("direction = ?", direction)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get host ID from path (for /api/v1/hosts/{hostId}/transfers)
	pathParts := splitPath(r.URL.Path)
	if len(pathParts) >= 4 && pathParts[3] == "hosts" && len(pathParts) >= 6 && pathParts[5] == "transfers" {
		if hostID, err := uuid.Parse(pathParts[4]); err == nil {
			query = query.Where("host_id = ?", hostID)
		}
	}

	var transfers []model.FileTransfer
	if err := query.Order("created_at DESC").Limit(100).Find(&transfers).Error; err != nil {
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve transfers")
		return
	}

	// Eager load host and user
	for i := range transfers {
		h.db.Preload("Host").Preload("User").First(&transfers[i], transfers[i].ID)
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": transfers,
	})
}

// Helper functions

func (h *FileTransferHandler) updateTransferStatus(transferID uuid.UUID, status model.FileTransferStatus, errMsg string) {
	now := time.Now()
	updates := map[string]interface{}{
		"status": status,
	}

	if status == model.FileTransferStatusFailed {
		updates["error_message"] = errMsg
	}

	if status == model.FileTransferStatusCompleted {
		updates["completed_at"] = &now
	}

	h.db.Model(&model.FileTransfer{}).Where("id = ?", transferID).Updates(updates)
}

func (h *FileTransferHandler) updateTransferProgress(transferID uuid.UUID, transferred int64) {
	h.db.Model(&model.FileTransfer{}).Where("id = ?", transferID).Update("transferred", transferred)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func splitPath(path string) []string {
	return strings.Split(strings.Trim(path, "/"), "/")
}

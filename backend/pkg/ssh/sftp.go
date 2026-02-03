// Package ssh provides SFTP file transfer functionality
package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/wangjialin/myops/pkg/model"
	"golang.org/x/crypto/ssh"
)

// SFTPClient wraps an SSH client with SFTP functionality
type SFTPClient struct {
	client *ssh.Client
	sftp   *sftp.Client
}

// SFTPConfig holds SFTP connection configuration
type SFTPConfig struct {
	HostID     string
	IPAddress  string
	Port       int
	Username   string
	Password   string
	PrivateKey []byte
	Timeout    time.Duration
}

// NewSFTPClient creates a new SFTP client
func NewSFTPClient(config *SFTPConfig) (*SFTPClient, error) {
	// Create SSH client config
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		User:            config.Username,
		Timeout:         config.Timeout,
	}

	// Add authentication method
	if config.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(config.Password),
		}
	}
	if len(config.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(config.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}

	// Establish TCP connection
	address := fmt.Sprintf("%s:%d", config.IPAddress, config.Port)
	conn, err := net.DialTimeout("tcp", address, config.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	// Create SSH connection
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("SSH handshake failed: %w", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)

	// Create SFTP client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &SFTPClient{
		client: client,
		sftp:   sftpClient,
	}, nil
}

// NewSFTPClientFromClient creates an SFTP client from an existing SSH client
func NewSFTPClientFromClient(client *ssh.Client) (*SFTPClient, error) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return &SFTPClient{
		client: client,
		sftp:   sftpClient,
	}, nil
}

// ListFiles lists files in a directory
func (c *SFTPClient) ListFiles(path string) ([]model.FileInfo, error) {
	// Get directory listing
	entries, err := c.sftp.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]model.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info := entry.Sys().(*sftp.FileStat)
		if info == nil {
			continue
		}

		files = append(files, model.FileInfo{
			Name:        entry.Name(),
			Path:        filepath.Join(path, entry.Name()),
			Size:        entry.Size(),
			Mode:        entry.Mode().String(),
			ModTime:     entry.ModTime(),
			IsDir:       entry.IsDir(),
			Permissions: fmt.Sprintf("%04o", entry.Mode().Perm()),
		})
	}

	return files, nil
}

// UploadFile uploads a file to the remote server
func (c *SFTPClient) UploadFile(localPath, remotePath string, progress chan<- int64) (int64, error) {
	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file info
	_, err = localFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create remote file
	remoteFile, err := c.sftp.Create(remotePath)
	if err != nil {
		return 0, fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Copy file content with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	var totalWritten int64

	for {
		n, err := localFile.Read(buffer)
		if n > 0 {
			written, writeErr := remoteFile.Write(buffer[:n])
			totalWritten += int64(written)

			if progress != nil {
				progress <- totalWritten
			}

			if writeErr != nil {
				return totalWritten, fmt.Errorf("failed to write to remote file: %w", writeErr)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return totalWritten, fmt.Errorf("failed to read from local file: %w", err)
		}
	}

	return totalWritten, nil
}

// DownloadFile downloads a file from the remote server
func (c *SFTPClient) DownloadFile(remotePath, localPath string, progress chan<- int64) (int64, error) {
	// Open remote file
	remoteFile, err := c.sftp.Open(remotePath)
	if err != nil {
		return 0, fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	// Get file info
	_, err = remoteFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Copy file content with progress tracking
	buffer := make([]byte, 32*1024) // 32KB buffer
	var totalWritten int64

	for {
		n, err := remoteFile.Read(buffer)
		if n > 0 {
			written, writeErr := localFile.Write(buffer[:n])
			totalWritten += int64(written)

			if progress != nil {
				progress <- totalWritten
			}

			if writeErr != nil {
				return totalWritten, fmt.Errorf("failed to write to local file: %w", writeErr)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return totalWritten, fmt.Errorf("failed to read from remote file: %w", err)
		}
	}

	return totalWritten, nil
}

// DeleteFile deletes a file on the remote server
func (c *SFTPClient) DeleteFile(path string) error {
	err := c.sftp.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// RenameFile renames a file on the remote server
func (c *SFTPClient) RenameFile(oldPath, newPath string) error {
	err := c.sftp.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	return nil
}

// CreateDirectory creates a directory on the remote server
func (c *SFTPClient) CreateDirectory(path string, mode os.FileMode) error {
	err := c.sftp.Mkdir(path)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if mode != 0 {
		err = c.sftp.Chmod(path, mode)
		if err != nil {
			return fmt.Errorf("failed to set directory permissions: %w", err)
		}
	}

	return nil
}

// GetFileInfo gets information about a file
func (c *SFTPClient) GetFileInfo(path string) (*model.FileInfo, error) {
	info, err := c.sftp.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &model.FileInfo{
		Name:        filepath.Base(path),
		Path:        path,
		Size:        info.Size(),
		Mode:        info.Mode().String(),
		ModTime:     info.ModTime(),
		IsDir:       info.IsDir(),
		Permissions: fmt.Sprintf("%04o", info.Mode().Perm()),
	}, nil
}

// Close closes the SFTP and SSH client
func (c *SFTPClient) Close() error {
	sftpErr := c.sftp.Close()
	sshErr := c.client.Close()

	if sftpErr != nil {
		return sftpErr
	}
	return sshErr
}

// ParseMode parses a mode string (like "0755") to os.FileMode
func ParseMode(modeStr string) os.FileMode {
	var mode uint32
	fmt.Sscanf(modeStr, "%o", &mode)
	return os.FileMode(mode)
}

// ValidatePath validates and cleans a file path
func ValidatePath(path string) (string, error) {
	// Clean the path
	path = filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}

	// Ensure path is absolute
	if !filepath.IsAbs(path) {
		path = "/" + path
	}

	return path, nil
}

// Package ssh provides SSH scanning functionality
package ssh

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Scanner scans hosts via SSH
type Scanner struct {
	timeout time.Duration
}

// ScanConfig represents scan configuration
type ScanConfig struct {
	IPRange      string
	Ports        []int
	Timeout      time.Duration
	MaxConcurrent int
}

// NewScanner creates a new SSH scanner
func NewScanner(timeout time.Duration) *Scanner {
	return &Scanner{
		timeout: timeout,
	}
}

// ScanRange scans an IP range for SSH hosts
func (s *Scanner) ScanRange(ctx context.Context, config *ScanConfig, resultChan chan<- *DiscoveredHost) error {
	ips, err := expandIPRange(config.IPRange)
	if err != nil {
		return fmt.Errorf("invalid IP range: %w", err)
	}

	// Create semaphore for concurrent scanning
	sem := make(chan struct{}, config.MaxConcurrent)
	var wg sync.WaitGroup

	for _, ip := range ips {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		for _, port := range config.Ports {
			wg.Add(1)
			go func(ip string, port int) {
				defer wg.Done()
				sem <- struct{}{}        // Acquire
				defer func() { <-sem }() // Release

				host := s.scanHost(ctx, ip, port)
				if host != nil {
					resultChan <- host
				}
			}(ip, port)
		}
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return nil
}

// scanHost attempts to scan a single host:port
func (s *Scanner) scanHost(ctx context.Context, ip string, port int) *DiscoveredHost {
	host := &DiscoveredHost{
		IPAddress: ip,
		Port:      port,
		Status:    "timeout",
	}

	// Try to connect with timeout
	dialer := net.Dialer{
		Timeout: s.timeout,
	}

	address := net.JoinHostPort(ip, strconv.Itoa(port))
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return host
	}
	defer conn.Close()

	// Set connection deadline
	conn.SetDeadline(time.Now().Add(s.timeout))

	// Try SSH handshake
	sshConfig := &ssh.ClientConfig{
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		ClientVersion:   "MyOps-Scanner",
		Timeout:         s.timeout,
		User:            "scan", // Dummy user for passive scanning
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		// Port is open but SSH handshake failed - still a found host
		host.Status = "open"
		return host
	}
	defer sshConn.Close()

	sshClient := ssh.NewClient(sshConn, chans, reqs)
	defer sshClient.Close()

	// Try to gather basic info without authentication
	host.Status = "success"

	// Try to execute simple commands (may fail on restricted hosts)
	session, err := sshClient.NewSession()
	if err != nil {
		return host
	}
	defer session.Close()

	// Try to get hostname
	if hostname, err := session.Output("hostname"); err == nil && len(hostname) > 0 {
		host.Hostname = string(hostname)
	}

	// Try to get OS info
	if osInfo, err := session.Output("uname -sr"); err == nil && len(osInfo) > 0 {
		host.OSType = string(osInfo)
	}

	return host
}

// DiscoveredHost represents a discovered host during scanning
type DiscoveredHost struct {
	IPAddress string
	Port      int
	Hostname  string
	OSType    string
	OSVersion string
	Status    string // "success", "open", "timeout", "error"
}

// expandIPRange expands a CIDR notation IP range to individual IPs
func expandIPRange(ipRange string) ([]string, error) {
	// Parse CIDR notation
	ip, ipNet, err := net.ParseCIDR(ipRange)
	if err != nil {
		// Try to parse as single IP
		if net.ParseIP(ipRange) != nil {
			return []string{ipRange}, nil
		}
		return nil, err
	}

	var ips []string
	// Create a copy of the IP to iterate
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
		// Safety limit for large ranges
		if len(ips) >= 65536 {
			break
		}
	}

	return ips, nil
}

// inc increments an IP address
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// GetEstimatedHostCount returns the number of hosts in a CIDR range
func GetEstimatedHostCount(ipRange string) (int, error) {
	ips, err := expandIPRange(ipRange)
	if err != nil {
		return 0, err
	}
	return len(ips), nil
}

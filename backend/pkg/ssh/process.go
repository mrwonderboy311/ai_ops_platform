// Package ssh provides SSH client for process management operations
package ssh

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/wangjialin/myops/pkg/model"
)

// ProcessClient handles process operations via SSH
type ProcessClient struct {
	client *SSHClient
}

// NewProcessClient creates a new process client
func NewProcessClient(config *SSHConfig) (*ProcessClient, error) {
	client, err := NewClient(config)
	if err != nil {
		return nil, err
	}
	return &ProcessClient{client: client}, nil
}

// Close closes the process client
func (p *ProcessClient) Close() error {
	return p.client.Close()
}

// ListProcesses lists all running processes on the remote host
func (p *ProcessClient) ListProcesses() ([]model.ProcessInfo, error) {
	// Use ps command to get process list
	// ps aux format: USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND
	cmd := "ps aux --sort=-%cpu | head -n 50" // Get top 50 processes by CPU
	output, err := p.client.ExecuteCommand(cmd, 30*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute ps command")
	}

	lines := strings.Split(strings.TrimSpace(output.Stdout), "\n")
	if len(lines) < 2 {
		return []model.ProcessInfo{}, nil
	}

	// Skip header line
	lines = lines[1:]
	processes := make([]model.ProcessInfo, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		process, err := parseProcessLine(line)
		if err != nil {
			// Skip invalid lines
			continue
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// GetProcess gets details of a specific process
func (p *ProcessClient) GetProcess(pid int32) (*model.ProcessInfo, error) {
	cmd := fmt.Sprintf("ps -p %d -o pid,user,%cpu,rss,stat,lstart,comm --no-headers", pid)
	output, err := p.client.ExecuteCommand(cmd, 10*time.Second)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get process info")
	}

	if strings.TrimSpace(output.Stdout) == "" {
		return nil, fmt.Errorf("process %d not found", pid)
	}

	process, err := parseProcessDetailLine(pid, strings.TrimSpace(output.Stdout))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse process info")
	}

	return &process, nil
}

// KillProcess kills a process with the specified signal
func (p *ProcessClient) KillProcess(pid int32, signal int32) error {
	sig := signal
	if sig == 0 {
		sig = 15 // Default to SIGTERM (15)
	}
	cmd := fmt.Sprintf("kill -%d %d", sig, pid)
	output, err := p.client.ExecuteCommand(cmd, 10*time.Second)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to kill process %d: %s", pid, output.Stderr))
	}
	return nil
}

// ExecuteCommand executes a command on the remote host
func (p *ProcessClient) ExecuteCommand(command string, timeout time.Duration, workingDir string) (*model.ExecuteCommandResponse, error) {
	var cmd string
	if workingDir != "" {
		cmd = fmt.Sprintf("cd %s && %s", workingDir, command)
	} else {
		cmd = command
	}

	start := time.Now()
	output, err := p.client.ExecuteCommand(cmd, timeout)
	duration := time.Since(start)

	if err != nil {
		return &model.ExecuteCommandResponse{
			ExitCode: -1,
			Stdout:   output.Stdout,
			Stderr:   output.Stderr,
			Duration: duration.String(),
		}, err
	}

	exitCode := int32(0)
	if output.ExitCode != nil {
		exitCode = *output.ExitCode
	}

	return &model.ExecuteCommandResponse{
		ExitCode: exitCode,
		Stdout:   output.Stdout,
		Stderr:   output.Stderr,
		Duration: duration.String(),
	}, nil
}

// parseProcessLine parses a line from ps aux output
func parseProcessLine(line string) (model.ProcessInfo, error) {
	// ps aux format: USER PID %CPU %MEM VSZ RSS TTY STAT START TIME COMMAND
	// Example: root         1  0.0  0.1 215624 11968 ?        Ss   2024 0:05 /sbin/init
	fields := strings.Fields(line)
	if len(fields) < 11 {
		return model.ProcessInfo{}, fmt.Errorf("invalid process line format")
	}

	user := fields[0]
	pid, err := strconv.ParseInt(fields[1], 10, 32)
	if err != nil {
		return model.ProcessInfo{}, err
	}
	cpuPercent, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return model.ProcessInfo{}, err
	}
	_, err = strconv.ParseFloat(fields[4], 64)
	if err != nil {
		return model.ProcessInfo{}, err
	}

	// RSS is in KB, convert to bytes
	rss, err := strconv.ParseInt(fields[5], 10, 64)
	if err != nil {
		return model.ProcessInfo{}, err
	}
	memoryBytes := rss * 1024

	tty := fields[6]
	status := parseProcessStatus(fields[7])
	start := fields[8]
	runTime := fields[9]

	// Command might have spaces, join remaining fields
	command := strings.Join(fields[10:], " ")

	// Extract process name from command
	name := command
	if parts := strings.Fields(command); len(parts) > 0 {
		// Get the base name of the command
		name = parts[0]
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
	}

	return model.ProcessInfo{
		PID:         int32(pid),
		Name:        name,
		Command:     command,
		User:        user,
		Status:      status,
		CPUPercent:  cpuPercent,
		MemoryBytes: memoryBytes,
		MemoryMB:    float64(memoryBytes) / (1024 * 1024),
		StartTime:   parseStartTime(start),
		RunTime:     runTime,
		Terminal:    tty,
	}, nil
}

// parseProcessDetailLine parses output from ps -p command
func parseProcessDetailLine(pid int32, line string) (model.ProcessInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return model.ProcessInfo{}, fmt.Errorf("invalid process detail format")
	}

	user := fields[0]
	cpuPercent, _ := strconv.ParseFloat(fields[1], 64)
	rss, _ := strconv.ParseInt(fields[2], 10, 64)
	status := parseProcessStatus(fields[3])
	startTime := parseStartTimeFromDetail(strings.Join(fields[4:9], " "))
	command := strings.Join(fields[9:], " ")

	memoryBytes := rss * 1024

	return model.ProcessInfo{
		PID:         pid,
		Name:        command,
		Command:     command,
		User:        user,
		Status:      status,
		CPUPercent:  cpuPercent,
		MemoryBytes: memoryBytes,
		MemoryMB:    float64(memoryBytes) / (1024 * 1024),
		StartTime:   startTime,
		RunTime:     "",
		Terminal:    "?",
	}, nil
}

// parseProcessStatus maps ps status to ProcessStatus
func parseProcessStatus(status string) model.ProcessStatus {
	// Common Linux process states:
	// S: Sleeping, R: Running, D: Waiting, Z: Zombie, T: Stopped
	switch {
	case strings.Contains(status, "R"):
		return model.ProcessStatusRunning
	case strings.Contains(status, "S"):
		return model.ProcessStatusSleeping
	case strings.Contains(status, "T"):
		return model.ProcessStatusStopped
	case strings.Contains(status, "Z"):
		return model.ProcessStatusZombie
	default:
		return model.ProcessStatusUnknown
	}
}

// parseStartTime parses the start time from ps output
func parseStartTime(start string) time.Time {
	// Start time can be in various formats:
	// - HH:MM (started today)
	// - Mon-DD (started this year)
	// - YYYY (started in previous years)

	now := time.Now()

	if strings.Contains(start, ":") {
		// HH:MM or HH:MM:SS format - started today
		parts := strings.Split(start, ":")
		hour, _ := strconv.Atoi(parts[0])
		min, _ := strconv.Atoi(parts[1])
		return time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
	}

	if strings.Contains(start, "-") {
		// Mon-DD format
		parts := strings.Split(start, "-")
		month := parseMonth(parts[0])
		day, _ := strconv.Atoi(parts[1])
		if month > 0 && day > 0 {
			return time.Date(now.Year(), month, day, 0, 0, 0, 0, now.Location())
		}
	}

	// Full year format
	if year, err := strconv.Atoi(start); err == nil && year > 2000 {
		return time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
	}

	return now
}

// parseStartTimeFromDetail parses the full start time from ps -p output
func parseStartTimeFromDetail(timeStr string) time.Time {
	// Format: "Wed Jan 15 10:30:00 2025"
	t, err := time.Parse("Mon Jan 2 15:04:05 2006", timeStr)
	if err != nil {
		return time.Now()
	}
	return t
}

// parseMonth parses month abbreviation
func parseMonth(month string) time.Month {
	months := map[string]time.Month{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4, "May": 5, "Jun": 6,
		"Jul": 7, "Aug": 8, "Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}
	return months[month]
}

// Package collector provides system information collection
package collector

import (
	"fmt"
	"net"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// HostInfo represents the collected host information
type HostInfo struct {
	Hostname   string           `json:"hostname"`
	IPAddress  string           `json:"ipAddress"`
	MACAddress string           `json:"macAddress,omitempty"`
	Gateway    string           `json:"gateway,omitempty"`
	OSType     string           `json:"osType"`
	OSVersion  string           `json:"osVersion"`
	KernelVersion string        `json:"kernelVersion"`
	Arch       string           `json:"arch"`
	CPUModel   string           `json:"cpuModel,omitempty"`
	CPUCores   int32            `json:"cpuCores"`
	MemoryTotal uint64          `json:"memoryTotal"` // bytes
	Disks      []DiskInfo       `json:"disks,omitempty"`
	Networks   []NetworkInfo    `json:"networks,omitempty"`
}

// DiskInfo represents disk information
type DiskInfo struct {
	Device     string `json:"device"`
	MountPoint string `json:"mountPoint"`
	Total      uint64 `json:"total"`     // bytes
	Free       uint64 `json:"free"`      // bytes
	Used       uint64 `json:"used"`      // bytes
	FileSystem string `json:"fileSystem"`
}

// NetworkInfo represents network interface information
type NetworkInfo struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	HardwareAddr string `json:"hardwareAddr,omitempty"`
	Flags     []string `json:"flags,omitempty"`
}

// Collector collects system information
type Collector struct {
	collectNetwork bool
}

// NewCollector creates a new collector
func NewCollector(collectNetwork bool) *Collector {
	return &Collector{
		collectNetwork: collectNetwork,
	}
}

// Collect collects all system information
func (c *Collector) Collect() (*HostInfo, error) {
	info := &HostInfo{}

	// Get host information
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	info.Hostname = hostInfo.Hostname
	info.OSType = hostInfo.OS
	info.OSVersion = hostInfo.Platform
	info.KernelVersion = hostInfo.KernelVersion
	info.Arch = hostInfo.KernelArch

	// Get CPU information
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
	}

	cores, err := cpu.Counts(true)
	if err == nil {
		info.CPUCores = int32(cores)
	}

	// Get memory information
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.MemoryTotal = memInfo.Total
	}

	// Get primary IP address
	ip, err := c.getPrimaryIP()
	if err == nil {
		info.IPAddress = ip
	}

	// Get network interfaces
	if c.collectNetwork {
		networks, err := c.collectNetworkInfo()
		if err == nil {
			info.Networks = networks
			// Set MAC address from first interface
			for _, n := range networks {
				if n.HardwareAddr != "" {
					info.MACAddress = n.HardwareAddr
					break
				}
			}
		}

		// Get gateway
		gateway, err := c.getDefaultGateway()
		if err == nil {
			info.Gateway = gateway
		}
	}

	// Get disk information
	disks, err := c.collectDiskInfo()
	if err == nil {
		info.Disks = disks
	}

	return info, nil
}

// getPrimaryIP gets the primary IP address
func (c *Collector) getPrimaryIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// collectNetworkInfo collects network interface information
func (c *Collector) collectNetworkInfo() ([]NetworkInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var networks []NetworkInfo
	for _, iface := range interfaces {
		// Skip loopback and down interfaces
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var addresses []string
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // Skip IPv6
			}

			addresses = append(addresses, ip.String())
		}

		if len(addresses) > 0 {
			networks = append(networks, NetworkInfo{
				Name:      iface.Name,
				Addresses: addresses,
				HardwareAddr: iface.HardwareAddr.String(),
			})
		}
	}

	return networks, nil
}

// getDefaultGateway gets the default gateway (simplified implementation)
func (c *Collector) getDefaultGateway() (string, error) {
	// This is a simplified implementation
	// In production, you'd read from /proc/net/route or use platform-specific APIs
	// For now, return a common default
	if runtime.GOOS == "darwin" {
		// macOS: could use netstat -nr
		return "", fmt.Errorf("not implemented on macOS")
	}

	// Linux: read from /proc/net/route
	// This is a placeholder - implement properly for production
	return "", fmt.Errorf("not implemented")
}

// collectDiskInfo collects disk information
func (c *Collector) collectDiskInfo() ([]DiskInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var disks []DiskInfo
	for _, partition := range partitions {
		// Skip pseudo filesystems
		if partition.Fstype == "proc" || partition.Fstype == "sysfs" ||
			partition.Fstype == "devtmpfs" || partition.Fstype == "tmpfs" {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		disks = append(disks, DiskInfo{
			Device:     partition.Device,
			MountPoint: partition.Mountpoint,
			Total:      usage.Total,
			Free:       usage.Free,
			Used:       usage.Used,
			FileSystem: partition.Fstype,
		})
	}

	return disks, nil
}

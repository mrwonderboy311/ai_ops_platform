// Package config provides auto-approval configuration
package config

import (
	"net"
	"strings"
)

// AutoApprovalConfig defines auto-approval rules
type AutoApprovalConfig struct {
	// CIDR ranges that should be auto-approved
	AutoApproveIPRanges []string `yaml:"auto_approve_ip_ranges" mapstructure:"auto_approve_ip_ranges"`
	// Tags that trigger auto-approval
	AutoApproveTags []string `yaml:"auto_approve_tags" mapstructure:"auto_approve_tags"`
}

// DefaultAutoApprovalConfig returns default auto-approval configuration
func DefaultAutoApprovalConfig() *AutoApprovalConfig {
	return &AutoApprovalConfig{
		AutoApproveIPRanges: []string{
			"10.0.0.0/8",     // Private network
			"172.16.0.0/12",  // Private network
			"192.168.0.0/16", // Private network
		},
		AutoApproveTags: []string{},
	}
}

// ShouldAutoApprove determines if a host should be auto-approved based on IP and labels
func (c *AutoApprovalConfig) ShouldAutoApprove(ip string, labels map[string]string) bool {
	// Check IP ranges
	for _, cidr := range c.AutoApproveIPRanges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil && ipNet.Contains(parsedIP) {
			return true
		}
	}

	// Check tags
	if labels != nil {
		for _, tag := range c.AutoApproveTags {
			// Check if tag exists as a key in labels
			if _, exists := labels[tag]; exists {
				return true
			}
			// Also check if any label value contains the tag
			for k, v := range labels {
				if strings.Contains(k, tag) || strings.Contains(v, tag) {
					return true
				}
			}
		}
	}

	return false
}

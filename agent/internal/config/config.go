// Package config provides agent configuration
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the agent configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Report   ReportConfig   `yaml:"report"`
	Collector CollectorConfig `yaml:"collector"`
}

// ServerConfig represents the server connection configuration
type ServerConfig struct {
	Endpoint string `yaml:"endpoint"`
	Token    string `yaml:"token"`
	Insecure bool   `yaml:"insecure"`
}

// ReportConfig represents the reporting configuration
type ReportConfig struct {
	Interval int `yaml:"interval"` // seconds
}

// CollectorConfig represents the collector configuration
type CollectorConfig struct {
	CollectProcesses bool `yaml:"collect_processes"`
	CollectNetwork   bool `yaml:"collect_network"`
}

const (
	// DefaultConfigPath is the default configuration file path
	DefaultConfigPath = "/etc/myops-agent/config.yaml"
	// DefaultReportInterval is the default reporting interval in seconds
	DefaultReportInterval = 60
	// DefaultEndpoint is the default server endpoint
	DefaultEndpoint = "https://localhost:8080"
)

// Load loads the configuration from the specified path
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if cfg.Report.Interval == 0 {
		cfg.Report.Interval = DefaultReportInterval
	}
	if cfg.Server.Endpoint == "" {
		cfg.Server.Endpoint = DefaultEndpoint
	}

	// Validate
	if cfg.Server.Token == "" {
		return nil, fmt.Errorf("server token is required")
	}

	return &cfg, nil
}

// LoadOrDefault loads the config from default path or returns default config
func LoadOrDefault() (*Config, error) {
	// Try to load from default path
	if cfg, err := Load(DefaultConfigPath); err == nil {
		return cfg, nil
	}

	// Return default config with minimal requirements
	return &Config{
		Server: ServerConfig{
			Endpoint: os.Getenv("MYOPS_SERVER_ENDPOINT"),
			Token:    os.Getenv("MYOPS_AGENT_TOKEN"),
			Insecure: os.Getenv("MYOPS_SERVER_INSECURE") == "true",
		},
		Report: ReportConfig{
			Interval: DefaultReportInterval,
		},
		Collector: CollectorConfig{
			CollectProcesses: false,
			CollectNetwork:   true,
		},
	}, nil
}

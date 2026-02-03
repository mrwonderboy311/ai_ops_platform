// Package main is the entry point for the MyOps Agent
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wangjialin/myops/agent/internal/collector"
	"github.com/wangjialin/myops/agent/internal/config"
	"github.com/wangjialin/myops/agent/internal/reporter"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("MyOps Agent v%s (built: %s)", version, buildTime)

	// Load configuration
	cfg, err := config.LoadOrDefault()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if cfg.Server.Token == "" {
		log.Fatal("Agent token is required. Set MYOPS_AGENT_TOKEN environment variable or configure in /etc/myops-agent/config.yaml")
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Server: %s", cfg.Server.Endpoint)
	log.Printf("  Report Interval: %d seconds", cfg.Report.Interval)
	log.Printf("  Collect Network: %v", cfg.Collector.CollectNetwork)

	// Create collector
	c := collector.NewCollector(cfg.Collector.CollectNetwork)

	// Create reporter
	r := reporter.NewReporter(cfg.Server.Endpoint, cfg.Server.Token, cfg.Server.Insecure)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initial report
	if err := reportOnce(c, r); err != nil {
		log.Printf("Initial report failed: %v", err)
	}

	// Start periodic reporting
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.Report.Interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping reporter...")
				return
			case <-ticker.C:
				if err := reportOnce(c, r); err != nil {
					log.Printf("Report failed: %v", err)
				}
			}
		}
	}()

	log.Println("Agent started successfully")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v", sig)
	cancel()

	log.Println("Agent stopped")
}

// reportOnce performs a single report
func reportOnce(c *collector.Collector, r *reporter.Reporter) error {
	log.Println("Collecting host information...")

	hostInfo, err := c.Collect()
	if err != nil {
		return fmt.Errorf("collection failed: %w", err)
	}

	log.Printf("Collected info for host: %s (IP: %s)", hostInfo.Hostname, hostInfo.IPAddress)

	log.Println("Sending report to server...")
	if err := r.Report(hostInfo); err != nil {
		return fmt.Errorf("report failed: %w", err)
	}

	log.Println("Report sent successfully")
	return nil
}

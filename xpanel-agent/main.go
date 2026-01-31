// xpanel-agent - VPN node agent for xpanel management system
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"xpanel-agent/config"
	"xpanel-agent/internal/agent"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Set log level
	level, err := logrus.ParseLevel(cfg.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.Logging.Format == "text" {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logger.WithFields(logrus.Fields{
		"node_id":   cfg.Node.ID,
		"node_name": cfg.Node.Name,
		"panel_url": cfg.Panel.URL,
	}).Info("Starting xpanel-agent")

	// Create and start agent
	nodeAgent := agent.NewAgent(cfg, logger)

	if err := nodeAgent.Start(); err != nil {
		logger.Fatalf("Failed to start agent: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	logger.Infof("Received signal %v, shutting down...", sig)

	// Graceful shutdown
	if err := nodeAgent.Stop(); err != nil {
		logger.Errorf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info("xpanel-agent stopped successfully")
}

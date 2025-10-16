package main

import (
	"fmt"
	"log"
	"os"
	"t-9/internal/api"
	"t-9/internal/config"
	"t-9/internal/logging"
	"t-9/internal/ws"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig
	
	// Initialize logger based on config
	logger := logging.NewLogger(logging.Info)
	logger.Info("Starting T-9 server", map[string]interface{}{
		"version": "1.0.0",
		"port":    cfg.Server.Port,
		"env":     getEnv("ENVIRONMENT", "development"),
	})

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Setup routes with WebSocket support
	r := api.SetupRoutes(hub)
	
	// Start server with configuration
	serverAddr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Server starting", map[string]interface{}{
		"address": serverAddr,
		"config":  cfg,
	})
	
	if err := r.Run(serverAddr); err != nil {
		logger.Error("Failed to start server", err, map[string]interface{}{
			"address": serverAddr,
		})
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
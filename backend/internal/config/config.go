package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logging  LoggingConfig
	CORS     CORSConfig
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// DatabaseConfig contains database-related configuration
type DatabaseConfig struct {
	URL     string
	MaxOpen int
	MaxIdle int
}

// LoggingConfig contains logging-related configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// CORSConfig contains CORS-related configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout: getEnvAsInt("SERVER_WRITE_TIMEOUT", 15),
			IdleTimeout:  getEnvAsInt("SERVER_IDLE_TIMEOUT", 60),
		},
		Database: DatabaseConfig{
			URL:     getEnv("DATABASE_URL", ""),
			MaxOpen: getEnvAsInt("DATABASE_MAX_OPEN", 25),
			MaxIdle: getEnvAsInt("DATABASE_MAX_IDLE", 25),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "INFO"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173", "http://localhost:4173"}),
			AllowedMethods: getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders: getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization"}),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsSlice gets environment variable as string slice with default value
func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple parsing for comma-separated values
		// In production, you might want to use JSON parsing
		return []string{value}
	}
	return defaultValue
}

// Default configuration instance
var DefaultConfig = LoadConfig()
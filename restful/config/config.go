package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration
type Config struct {
	Server ServerConfig `json:"server"`
	AWS    AWSConfig    `json:"aws"`
	Auth   AuthConfig   `json:"auth"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Token string `json:"token"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: getEnvInt("PORT", 8080),
			Host: getEnvString("HOST", "0.0.0.0"),
		},
		AWS: AWSConfig{
			Region:    getEnvString("AWS_REGION", "us-east-1"),
			Bucket:    getEnvString("AWS_S3_BUCKET", ""),
			AccessKey: getEnvString("AWS_ACCESS_KEY_ID", ""),
			SecretKey: getEnvString("AWS_SECRET_ACCESS_KEY", ""),
		},
		Auth: AuthConfig{
			Token: getEnvString("AUTH_TOKEN", ""),
		},
	}

	// Validate required configuration
	if cfg.AWS.Bucket == "" {
		return nil, fmt.Errorf("AWS_S3_BUCKET environment variable is required")
	}

	return cfg, nil
}

// getEnvString gets string from environment with default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets integer from environment with default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
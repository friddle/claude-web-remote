package config

import (
	"os"
	"strconv"
)

// Config server configuration
type Config struct {
	PikoUpstreamPort int
	PikoToken        string
	ListenPort       int
	EnableTLS        bool
	TLSCertFile      string
	TLSKeyFile       string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		PikoUpstreamPort: getEnvInt("PIKO_UPSTREAM_PORT", 8022),
		PikoToken:        getEnvOrDefault("PIKO_TOKEN", ""),
		ListenPort:       getEnvInt("LISTEN_PORT", 80),
		EnableTLS:        getEnvBool("ENABLE_TLS", false),
		TLSCertFile:      getEnvOrDefault("TLS_CERT_FILE", ""),
		TLSKeyFile:       getEnvOrDefault("TLS_KEY_FILE", ""),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

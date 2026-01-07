package environment

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Loader handles environment variable loading with proper priority
type Loader struct {
	envVars []string // Command-line --env flags (highest priority)
}

// NewLoader creates a new environment loader
func NewLoader(envVars []string) *Loader {
	return &Loader{envVars: envVars}
}

// Load loads environment variables with priority:
// 1. System environment variables (already in os.Environ())
// 2. .env file in current directory
// 3. Command-line --env flags (highest priority, overrides others)
func (l *Loader) Load() {
	// Priority 1: System environment variables are already loaded in os.Environ()

	// Priority 2: Load from .env file if exists
	l.loadEnvFile()

	// Priority 3: Command-line --env flags (highest priority, overrides all)
	if len(l.envVars) > 0 {
		for _, envVar := range l.envVars {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}
}

// loadEnvFile loads environment variables from .env file
func (l *Loader) loadEnvFile() {
	// Check for .env file in current directory
	envFiles := []string{".env", ".claude.env"}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			fmt.Printf("✓ Loading environment variables from: %s\n", envFile)
			l.parseEnvFile(envFile)
			return // Only load the first found file
		}
	}
}

// parseEnvFile parses a .env file and sets environment variables
func (l *Loader) parseEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}

			// Only set if not already set in system environment
			// (Command-line env vars will override these later)
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, value)
			}
		}
	}
}

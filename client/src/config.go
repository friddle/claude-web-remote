package src

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Config configuration structure
type Config struct {
	Name               string   // deprecated: piko client name
	Host               string   // remote server host (e.g., clauded.friddle.me)
	Session            string   // session ID (auto-generated if empty)
	Password           string   // password for authentication
	Flags              string   // flags to pass to claude-code
	EnvVars            []string // environment variables to pass
	Remote             string   // remote piko server address (format: host:port)
	ServerPort         int      // piko server port
	GottyPort          int      // local gotty port (auto allocated)
	AutoExit           bool     // enable 24-hour auto exit (default: true)
	InsecureSkipVerify bool     // skip HTTPS certificate verification (default: false)
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	return &Config{
		Name:               getEnvOrDefault("NAME", ""),
		Host:               getEnvOrDefault("HOST", ""),
		Session:            getEnvOrDefault("SESSION", ""),
		Password:           getEnvOrDefault("PASSWORD", ""),
		Flags:              getEnvOrDefault("FLAGS", ""),
		EnvVars:            []string{},
		Remote:             getEnvOrDefault("REMOTE", ""),
		ServerPort:         getEnvIntOrDefault("SERVER_PORT", 8022),
		GottyPort:          0,                                         // will be auto allocated on startup
		AutoExit:           getEnvBoolOrDefault("AUTO_EXIT", true),    // read auto exit setting from env, default true
		InsecureSkipVerify: getEnvBoolOrDefault("INSECURE_SKIP_VERIFY", false), // read skip cert verify from env, default false
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	// Security mechanism for default host
	isDefaultHost := c.Host == "clauded.friddle.me" || c.Host == "www.clauded.friddle.me"

	if isDefaultHost {
		// Enforce password for default host
		if c.Password == "" {
			return fmt.Errorf("password is required when using the default host (clauded.friddle.me) for security reasons")
		}

		// Generate fixed UUID session for default host
		if c.Session == "" {
			c.Session = generateSessionID()
		}

		// Clear user-provided session for default host (use auto-generated UUID)
		if c.Name != "" {
			c.Session = generateSessionID()
		}
	} else {
		// For custom hosts, session and password are optional
		if c.Session == "" && c.Name != "" {
			c.Session = c.Name
		}
		if c.Session == "" {
			c.Session = generateSessionID()
		}
	}

	// Build remote address from host
	if c.Remote == "" {
		c.Remote = c.buildRemoteAddress()
	}

	return nil
}

// GetSessionID returns the session ID (generates one if needed)
func (c *Config) GetSessionID() string {
	if c.Session == "" {
		c.Session = generateSessionID()
	}
	return c.Session
}

// IsDefaultHost checks if the host is the default public host
func (c *Config) IsDefaultHost() bool {
	return c.Host == "clauded.friddle.me" || c.Host == "www.clauded.friddle.me"
}

// ShouldShowSecurityWarning checks if security warning should be shown
func (c *Config) ShouldShowSecurityWarning() bool {
	return c.IsDefaultHost()
}

// GetRemoteHost gets the remote host address
func (c *Config) GetRemoteHost() string {
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 1 {
		return parts[0]
	}
	return c.Host
}

// GetRemotePort gets the remote port
func (c *Config) GetRemotePort() int {
	parts := strings.Split(c.Remote, ":")
	if len(parts) >= 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}
	return 8088
}

// buildRemoteAddress builds the remote address from host
func (c *Config) buildRemoteAddress() string {
	// If host already includes port, use it as-is
	if strings.Contains(c.Host, ":") {
		return c.Host
	}

	// Default to port 8088
	return fmt.Sprintf("%s:8088", c.Host)
}

// FindAvailablePort finds an available port starting from 8080
func (c *Config) FindAvailablePort() int {
	startPort := 8080
	for port := startPort; port < startPort+100; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return startPort // return default port if all are unavailable
}

// isPortAvailable checks if a port is available
func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return uuid.New().String()
}

// getEnvOrDefault gets environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvIntOrDefault gets integer environment variable or default value
func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBoolOrDefault gets boolean environment variable or default value
func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

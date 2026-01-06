package src

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

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

	// Normalize host: remove https:// prefix for internal processing
	host := c.Host
	if strings.HasPrefix(host, "https://") {
		host = strings.TrimPrefix(host, "https://")
	} else if strings.HasPrefix(host, "http://") {
		host = strings.TrimPrefix(host, "http://")
	}

	// Security mechanism for default host
	isDefaultHost := host == "clauded.friddle.me" || host == "www.clauded.friddle.me"

	if isDefaultHost {
		// For default host, auto-generate both session and password
		// Ignore user-provided session for security
		if c.Session == "" {
			c.Session = generateShortSessionID()
		} else {
			// User provided a session, but we should ignore it and warn
			c.Session = generateShortSessionID()
		}

		// Auto-generate password for default host
		if c.Password == "" {
			c.Password = generateShortPassword()
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
	return c.Host
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

// generateShortSessionID generates a 5-character session ID
func generateShortSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[randInt()%len(charset)]
	}
	return string(b)
}

// generateShortPassword generates a 5-character password
func generateShortPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[randInt()%len(charset)]
	}
	return string(b)
}

// randInt generates a random integer
func randInt() int {
	// Simple random number generator using time
	return int(time.Now().UnixNano()%1000000007)
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

package src

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"clauded-client/src/platform"
)

// Config configuration structure
type Config struct {
	Name               string   `json:"name"`               // deprecated: piko client name
	Remote             string   `json:"remote"`             // remote server address (format: https://host or host:port)
	Session            string   `json:"session"`            // session ID (auto-generated if empty)
	Password           string   `json:"-"`                  // password for authentication (hidden from JSON)
	CodeCmd            string   `json:"codecmd"`            // AI command tool to use
	Flags              string   `json:"flags"`              // flags to pass to claude-code
	EnvVars            []string `json:"-"`                  // environment variables (hidden from JSON, may contain secrets)
	GottyPort          int      `json:"port"`               // local gotty port (auto allocated)
	AutoExit           bool     `json:"auto_exit"`          // enable 24-hour auto exit (default: true)
	InsecureSkipVerify bool     `json:"insecure_skip_verify"` // skip HTTPS certificate verification
	Daemon             bool     `json:"daemon"`             // run as daemon (background mode)
	SkipInstall        bool     `json:"skip_install"`       // skip claude-code installation check
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	return &Config{
		Name:               getEnvOrDefault("NAME", ""),
		Remote:             getEnvOrDefault("REMOTE", ""),
		Session:            getEnvOrDefault("SESSION", ""),
		Password:           getEnvOrDefault("PASSWORD", ""),
		CodeCmd:            getEnvOrDefault("CODECMD", "claude"),
		Flags:              getEnvOrDefault("FLAGS", ""),
		EnvVars:            []string{},
		GottyPort:          0,                                              // will be auto allocated on startup
		AutoExit:           getEnvBoolOrDefault("AUTO_EXIT", true),         // read auto exit setting from env, default true
		InsecureSkipVerify: getEnvBoolOrDefault("INSECURE_SKIP_VERIFY", false), // read skip cert verify from env, default false
		Daemon:             getEnvBoolOrDefault("DAEMON", true),            // read daemon mode from env, default true
		SkipInstall:        false,
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Remote == "" {
		return fmt.Errorf("remote cannot be empty")
	}

	// Normalize remote: remove protocol prefix for host extraction
	remote := c.Remote
	host := c.Remote
	if strings.HasPrefix(remote, "https://") {
		host = strings.TrimPrefix(remote, "https://")
	} else if strings.HasPrefix(remote, "http://") {
		host = strings.TrimPrefix(remote, "http://")
	}

	// Extract host part (remove port if present)
	hostParts := strings.Split(host, ":")
	hostname := hostParts[0]

	// Security mechanism for default host
	isDefaultHost := platform.IsDefaultHost(hostname)

	if isDefaultHost {
		// For default host, auto-generate both session and password if not provided
		if c.Session == "" {
			c.Session = generateShortSessionID()
		}

		// Auto-generate password for default host
		if c.Password == "" {
			c.Password = generateShortPassword()
		}
	} else {
		// For custom hosts, session is optional, use short session ID
		if c.Session == "" {
			if c.Name != "" {
				c.Session = c.Name
			} else {
				c.Session = generateShortSessionID()
			}
		}
		// Password is optional for custom hosts
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

// IsDefaultHost checks if the remote is the default public host
func (c *Config) IsDefaultHost() bool {
	// Extract hostname from remote
	remote := c.Remote
	if strings.HasPrefix(remote, "https://") {
		remote = strings.TrimPrefix(remote, "https://")
	} else if strings.HasPrefix(remote, "http://") {
		remote = strings.TrimPrefix(remote, "http://")
	}

	// Extract host part (remove port if present)
	hostParts := strings.Split(remote, ":")
	hostname := hostParts[0]

	return platform.IsDefaultHost(hostname)
}

// ShouldShowSecurityWarning checks if security warning should be shown
func (c *Config) ShouldShowSecurityWarning() bool {
	return c.IsDefaultHost()
}

// GetRemoteHost gets the remote host address
func (c *Config) GetRemoteHost() string {
	remote := c.Remote
	// Remove protocol prefix if present
	if strings.HasPrefix(remote, "https://") {
		remote = strings.TrimPrefix(remote, "https://")
	} else if strings.HasPrefix(remote, "http://") {
		remote = strings.TrimPrefix(remote, "http://")
	}

	// Remove port if present
	parts := strings.Split(remote, ":")
	return parts[0]
}

// GetRemotePort gets the remote port
func (c *Config) GetRemotePort() int {
	remote := c.Remote
	// Remove protocol prefix if present
	if strings.HasPrefix(remote, "https://") {
		remote = strings.TrimPrefix(remote, "https://")
	} else if strings.HasPrefix(remote, "http://") {
		remote = strings.TrimPrefix(remote, "http://")
	}

	// Extract port if present
	parts := strings.Split(remote, ":")
	if len(parts) >= 2 {
		if port, err := strconv.Atoi(parts[1]); err == nil {
			return port
		}
	}

	// Default ports based on protocol
	if strings.HasPrefix(c.Remote, "https://") {
		return 443
	}
	return 80
}

// GetHTTPURL returns the HTTP URL for accessing the session
func (c *Config) GetHTTPURL() string {
	remote := c.Remote
	// Ensure protocol prefix
	if !strings.HasPrefix(remote, "http://") && !strings.HasPrefix(remote, "https://") {
		remote = "http://" + remote
	}

	// Parse URL to remove port
	if strings.HasPrefix(remote, "https://") {
		host := strings.TrimPrefix(remote, "https://")
		// Remove port part
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		return "https://" + host
	} else if strings.HasPrefix(remote, "http://") {
		host := strings.TrimPrefix(remote, "http://")
		// Remove port part
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		return "http://" + host
	}
	return remote
}

// GetPikoAddress returns the piko server address (host:port)
func (c *Config) GetPikoAddress() string {
	remote := c.Remote

	// Extract host and remove port for default WebSocket ports
	if strings.HasPrefix(remote, "https://") {
		remote = strings.TrimPrefix(remote, "https://")
		// Remove port part to use default 443 for wss://
		if idx := strings.Index(remote, ":"); idx != -1 {
			remote = remote[:idx]
		}
		return "https://" + remote
	} else if strings.HasPrefix(remote, "http://") {
		remote = strings.TrimPrefix(remote, "http://")
		// Remove port part to use default 80 for ws://
		if idx := strings.Index(remote, ":"); idx != -1 {
			remote = remote[:idx]
		}
		return "http://" + remote
	} else {
		// No protocol specified, use as-is
		return remote
	}
}

// FindAvailablePort finds an available port starting from DefaultGottyPortStart
func (c *Config) FindAvailablePort() int {
	startPort := platform.DefaultGottyPortStart
	for port := startPort; port < startPort+platform.GottyPortRange; port++ {
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

// generateShortSessionID generates a short session ID
func generateShortSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, platform.SessionIDLength)
	for i := range b {
		b[i] = charset[randInt()%len(charset)]
	}
	return string(b)
}

// generateShortPassword generates a short password
func generateShortPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, platform.PasswordLength)
	for i := range b {
		b[i] = charset[randInt()%len(charset)]
	}
	return string(b)
}

// randInt generates a random integer using crypto/rand
func randInt() int {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000007))
	return int(n.Int64())
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

// ToArgs converts the config to command line arguments
func (c *Config) ToArgs() []string {
	args := []string{}

	// --remote (Mandatory)
	args = append(args, "--remote", c.Remote)

	// --session (Mandatory)
	args = append(args, "--session", c.GetSessionID())

	// --password (Mandatory)
	if c.Password != "" {
		args = append(args, "--password", c.Password)
	}

	// --codecmd
	if c.CodeCmd != "" {
		args = append(args, "--codecmd", c.CodeCmd)
	}

	// --flags
	if c.Flags != "" {
		args = append(args, "--flags", c.Flags)
	}

	// --env (multiple)
	for _, env := range c.EnvVars {
		args = append(args, "--env", env)
	}

	// --auto-exit
	args = append(args, fmt.Sprintf("--auto-exit=%t", c.AutoExit))

	// --insecure-skip-verify
	if c.InsecureSkipVerify {
		args = append(args, "--insecure-skip-verify")
	}

	// --skip-install-check
	if c.SkipInstall {
		args = append(args, "--skip-install-check")
	}

	// Explicitly disable daemon mode for the child process
	args = append(args, "--daemon=false")

	return args
}

package src

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/andydunstall/piko/agent/config"
	"github.com/andydunstall/piko/agent/reverseproxy"
	"github.com/andydunstall/piko/client"
	"github.com/andydunstall/piko/pkg/log"
	"github.com/oklog/run"
	"github.com/sorenisanerd/gotty/backend/localcommand"
	"github.com/sorenisanerd/gotty/server"
)

// ServiceManager service manager
type ServiceManager struct {
	config *Config
	ctx    context.Context
	cancel context.CancelFunc
}

// NewServiceManager creates a new service manager
func NewServiceManager(config *Config) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceManager{
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start starts all services
func (sm *ServiceManager) Start() error {
	fmt.Printf("🚀 Starting clauded client\n")
	fmt.Printf("Session ID: %s\n", sm.config.GetSessionID())
	fmt.Printf("Remote server: %s\n", sm.config.Remote)
	fmt.Printf("Auto exit: %t\n", sm.config.AutoExit)

	// Show security warning for default host
	if sm.config.ShouldShowSecurityWarning() {
		fmt.Printf("\n⚠️  WARNING: You are using the public default host (clauded.friddle.me)\n")
		fmt.Printf("   This may have security risks. Your session is protected with a password.\n")
		fmt.Printf("   For better security, consider deploying your own server.\n\n")
	}

	// Auto-allocate available port
	sm.config.GottyPort = sm.config.FindAvailablePort()
	fmt.Printf("Local listening port: %d\n", sm.config.GottyPort)

	// Use oklog/run to start services
	return sm.startServices()
}

// startServices uses oklog/run to start all services
func (sm *ServiceManager) startServices() error {
	var g run.Group

	// Start piko service
	g.Add(func() error {
		err := sm.startPiko()
		if err != nil {
			fmt.Printf("Failed to start piko: %v\n", err)
			return err
		}
		// Wait for context cancellation
		<-sm.ctx.Done()
		return sm.ctx.Err()
	}, func(error) {
		// piko service will stop automatically when context is cancelled
	})

	// Start gotty service
	g.Add(func() error {
		err := sm.startGotty()
		if err != nil {
			fmt.Printf("Failed to start gotty: %v\n", err)
			return err
		}
		// Wait for context cancellation
		<-sm.ctx.Done()
		return sm.ctx.Err()
	}, func(error) {
		// gotty service will stop automatically when context is cancelled
	})

	// Signal handling
	g.Add(func() error {
		c := make(chan os.Signal, 1)

		// Set different signals based on operating system
		if runtime.GOOS == "windows" {
			// Windows supports Ctrl+C (SIGINT) and Ctrl+Break
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		} else {
			// Unix-like systems support more signals
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		}

		select {
		case sig := <-c:
			fmt.Printf("\n🛑 Received stop signal %v, shutting down services...\n", sig)
			sm.cancel() // Cancel context immediately
			return nil
		case <-sm.ctx.Done():
			return sm.ctx.Err()
		}
	}, func(error) {
		sm.cancel()
	})

	// 24-hour timeout - only enable when AutoExit is true
	if sm.config.AutoExit {
		g.Add(func() error {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				fmt.Printf("\n⏰ Service has been running for 24 hours, stopping...\n")
				sm.cancel()
				return nil
			case <-sm.ctx.Done():
				return sm.ctx.Err()
			}
		}, func(error) {
			sm.cancel()
		})
	}

	sessionID := sm.config.GetSessionID()
	fmt.Printf("✅ Services started successfully!\n")
	fmt.Printf("🌐 Access URL: http://localhost:%d\n", sm.config.GottyPort)
	if sm.config.Password != "" {
		fmt.Printf("🔐 HTTP auth: username=%s, password=%s\n", sessionID, sm.config.Password)
	} else {
		fmt.Printf("⚠️  HTTP authentication not enabled\n")
	}
	fmt.Printf("Press Ctrl+C to stop services\n")

	// Run all services
	return g.Run()
}

// Wait waits for services to run (deprecated, use Start method)
func (sm *ServiceManager) Wait() {
	fmt.Printf("⚠️  Wait method is deprecated, please use Start method\n")
}

// Stop stops all services
func (sm *ServiceManager) Stop() {
	fmt.Printf("✅ Services stopped\n")
}

// startGotty starts gotty
func (sm *ServiceManager) startGotty() error {
	// Create gotty server options
	fmt.Print("Starting gotty...")
	sessionID := sm.config.GetSessionID()
	options := &server.Options{
		Address:         "127.0.0.1",
		Port:            fmt.Sprintf("%d", sm.config.GottyPort),
		Path:            "/" + sessionID,
		PermitWrite:     true,
		TitleFormat:     "Claude Code - " + sessionID,
		WSOrigin:        ".*",                          // Allow WebSocket connections from any origin
		EnableBasicAuth: sm.config.Password != "",      // Only enable HTTP basic auth when password is not empty
	}

	if sm.config.Password != "" {
		options.Credential = sessionID + ":" + sm.config.Password // Set auth credential: username:password
	}

	// Build claude-code command with flags
	command, args := sm.getClaudeCodeCommand()

	// Create local command factory
	backendOptions := &localcommand.Options{}
	factory, err := localcommand.NewFactory(command, args, backendOptions)
	if err != nil {
		return fmt.Errorf("failed to create gotty factory: %w", err)
	}

	// Create gotty server
	srv, err := server.New(factory, options)
	if err != nil {
		return fmt.Errorf("failed to create gotty server: %w", err)
	}

	// Start gotty server in a separate goroutine
	go func() {
		err := srv.Run(sm.ctx)
		if err != nil && err != context.Canceled {
			fmt.Printf("gotty server runtime error: %v\n", err)
		}
	}()

	fmt.Print(" done\n")
	return nil
}

// startPiko starts piko client
func (sm *ServiceManager) startPiko() error {
	// Create piko configuration
	fmt.Printf("Starting piko...")
	remote := sm.config.Remote
	if strings.HasPrefix(remote, "http") {
		remote = sm.config.Remote
	} else {
		remote = fmt.Sprintf("http://%s", sm.config.Remote)
	}
	sessionID := sm.config.GetSessionID()
	conf := &config.Config{
		Connect: config.ConnectConfig{
			URL:     remote,
			Timeout: 30 * time.Second,
		},
		Listeners: []config.ListenerConfig{
			{
				EndpointID: sessionID,
				Protocol:   config.ListenerProtocolHTTP,
				Addr:       fmt.Sprintf("127.0.0.1:%d", sm.config.GottyPort),
				AccessLog:  false,
				Timeout:    30 * time.Second,
				TLS:        config.TLSConfig{},
			},
		},
		Log: log.Config{
			Level:      "info",
			Subsystems: []string{},
		},
		GracePeriod: 30 * time.Second,
	}

	// Create logger
	logger, err := log.NewLogger("info", []string{})
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	// Validate configuration
	if err := conf.Validate(); err != nil {
		return fmt.Errorf("piko configuration validation failed: %w", err)
	}

	// Parse connection URL
	connectURL, err := url.Parse(conf.Connect.URL)
	if err != nil {
		return fmt.Errorf("failed to parse connection URL: %w", err)
	}

	// Create TLS configuration
	var tlsConfig *tls.Config
	if sm.config.InsecureSkipVerify {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		fmt.Printf(" (HTTPS certificate verification skipped)")
	}

	// Create upstream client
	upstream := &client.Upstream{
		URL:       connectURL,
		TLSConfig: tlsConfig,
		Logger:    logger.WithSubsystem("client"),
	}

	// Create connections for each listener
	for _, listenerConfig := range conf.Listeners {
		ln, err := upstream.Listen(sm.ctx, listenerConfig.EndpointID)
		if err != nil {
			return fmt.Errorf("failed to listen on endpoint %s: %w", listenerConfig.EndpointID, err)
		}

		// Create HTTP proxy server
		metrics := reverseproxy.NewMetrics("proxy")
		server := reverseproxy.NewServer(listenerConfig, metrics, logger)
		if server == nil {
			return fmt.Errorf("failed to create HTTP proxy server")
		}

		// Start proxy server
		go func() {
			if err := server.Serve(ln); err != nil && err != context.Canceled {
				fmt.Printf("proxy server runtime error: %v\n", err)
			}
		}()
	}

	fmt.Print(" done\n")
	return nil
}

// getClaudeCodeCommand builds the claude-code command with flags and environment variables
func (sm *ServiceManager) getClaudeCodeCommand() (string, []string) {
	// Priority 1: Try to find 'claude' command in system PATH
	command := sm.findClaudeCommand()

	args := []string{}

	// Add flags if provided
	if sm.config.Flags != "" {
		// Parse flags string (e.g., "--model opus --fast-model sonnet")
		flagParts := strings.Fields(sm.config.Flags)
		args = append(args, flagParts...)
	}

	// Load environment variables with priority:
	// 1. System environment variables (already in os.Environ())
	// 2. .env file in current directory
	// 3. Command-line --env flags (highest priority, overrides others)
	sm.loadEnvironmentVariables()

	return command, args
}

// findClaudeCommand searches for claude/claude-code command in priority order
func (sm *ServiceManager) findClaudeCommand() string {
	// Priority 1: Check if 'claude' command exists in PATH
	if path, err := exec.LookPath("claude"); err == nil {
		fmt.Printf("✓ Using claude command from: %s\n", path)
		// Verify it's actually Claude Code
		if sm.isClaudeCodeCommand("claude") {
			return "claude"
		}
	}

	// Priority 2: Check if 'claude-code' command exists in PATH
	if path, err := exec.LookPath("claude-code"); err == nil {
		fmt.Printf("✓ Using claude-code command from: %s\n", path)
		return "claude-code"
	}

	// Priority 3: Check ~/.local/bin
	homeDir, err := os.UserHomeDir()
	if err == nil {
		localBinPath := filepath.Join(homeDir, ".local", "bin")
		claudeCodePath := filepath.Join(localBinPath, "claude-code")
		if _, err := os.Stat(claudeCodePath); err == nil {
			// Add to PATH
			currentPath := os.Getenv("PATH")
			if !strings.Contains(currentPath, localBinPath) {
				os.Setenv("PATH", localBinPath+":"+currentPath)
				fmt.Printf("✓ Added %s to PATH\n", localBinPath)
			}
			fmt.Printf("✓ Using claude-code from: %s\n", claudeCodePath)
			return "claude-code"
		}
	}

	fmt.Println("⚠️  Warning: claude/claude-code command not found in PATH")
	return "claude-code" // Default fallback
}

// isClaudeCodeCommand verifies if the command is actually Claude Code
func (sm *ServiceManager) isClaudeCodeCommand(cmd string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd, "--version")
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return false
	}

	outputStr := string(output)
	return strings.Contains(outputStr, "Claude Code") || strings.Contains(outputStr, "claude")
}

// loadEnvironmentVariables loads environment variables with proper priority
func (sm *ServiceManager) loadEnvironmentVariables() {
	// Priority 1: System environment variables are already loaded in os.Environ()

	// Priority 2: Load from .env file if exists
	sm.loadEnvFile()

	// Priority 3: Command-line --env flags (highest priority, overrides all)
	if len(sm.config.EnvVars) > 0 {
		for _, envVar := range sm.config.EnvVars {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}
}

// loadEnvFile loads environment variables from .env file
func (sm *ServiceManager) loadEnvFile() {
	// Check for .env file in current directory
	envFiles := []string{".env", ".claude.env"}

	for _, envFile := range envFiles {
		if _, err := os.Stat(envFile); err == nil {
			fmt.Printf("✓ Loading environment variables from: %s\n", envFile)
			sm.parseEnvFile(envFile)
			return // Only load the first found file
		}
	}
}

// parseEnvFile parses a .env file and sets environment variables
func (sm *ServiceManager) parseEnvFile(filename string) {
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

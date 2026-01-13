package src

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/oklog/run"

	"clauded-client/src/commands"
	"clauded-client/src/environment"
	"clauded-client/src/platform"
	"clauded-client/src/services"
)

// ServiceManager service manager
type ServiceManager struct {
	config   *Config
	ctx      context.Context
	cancel   context.CancelFunc
	notifier *Notifier
}

// NewServiceManager creates a new service manager
func NewServiceManager(config *Config) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())
	notifier := NewNotifier(config.GetHTTPURL(), config.GetSessionID())
	return &ServiceManager{
		config:   config,
		ctx:      ctx,
		cancel:   cancel,
		notifier: notifier,
	}
}

// Start starts all services
func (sm *ServiceManager) Start() error {
	fmt.Printf("🚀 Starting clauded client\n")
	fmt.Printf("Code command: %s\n", sm.config.CodeCmd)
	fmt.Printf("Session ID: %s\n", sm.config.GetSessionID())
	fmt.Printf("Remote server: %s\n", sm.config.GetHTTPURL())
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

	// If daemon mode, fork to background before starting services
	if sm.config.Daemon {
		if err := sm.daemonize(); err != nil {
			return fmt.Errorf("failed to daemonize: %w", err)
		}
		return nil
	}

	// Use oklog/run to start services
	return sm.startServices()
}

// startServices uses oklog/run to start all services
func (sm *ServiceManager) startServices() error {
	var g run.Group

	// Get the command to run
	command, args := sm.getClaudeCodeCommand()

	// Load environment variables
	envLoader := environment.NewLoader(sm.config.EnvVars)
	envLoader.Load()

	// Start piko service for gotty
	g.Add(func() error {
		pikoConfig := services.PikoConfig{
			RemoteURL:   sm.config.GetPikoAddress(),
			EndpointID:  sm.config.GetSessionID(),
			LocalAddr:   fmt.Sprintf("127.0.0.1:%d", sm.config.GottyPort),
			Timeout:     30 * time.Second,
			GracePeriod: 30 * time.Second,
			AccessLog:   false,
		}
		pikoService := services.NewPikoService(pikoConfig, sm.ctx, sm.config.InsecureSkipVerify)
		err := pikoService.Start()
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

	// Start piko services for attach-ports
	for _, port := range sm.config.AttachPorts {
		// Capture port variable for goroutine
		attachPort := port
		g.Add(func() error {
			// Use endpoint ID format: {sessionID}-{port}
			endpointID := fmt.Sprintf("%s-%d", sm.config.GetSessionID(), attachPort)
			pikoConfig := services.PikoConfig{
				RemoteURL:   sm.config.GetPikoAddress(),
				EndpointID:  endpointID,
				LocalAddr:   fmt.Sprintf("127.0.0.1:%d", attachPort),
				Timeout:     30 * time.Second,
				GracePeriod: 30 * time.Second,
				AccessLog:   false,
			}
			pikoService := services.NewPikoService(pikoConfig, sm.ctx, sm.config.InsecureSkipVerify)
			err := pikoService.Start()
			if err != nil {
				fmt.Printf("Failed to start piko for port %d: %v\n", attachPort, err)
				return err
			}
			// Wait for context cancellation
			<-sm.ctx.Done()
			return sm.ctx.Err()
		}, func(error) {
			// piko service will stop automatically when context is cancelled
		})
	}

	// Start gotty service
	g.Add(func() error {
		sessionID := sm.config.GetSessionID()
		gottyConfig := services.GottyConfig{
			Address:         "127.0.0.1",
			Port:            sm.config.GottyPort,
			Path:            "/" + sessionID,
			PermitWrite:     true,
			TitleFormat:     sm.config.CodeCmd + " - " + sessionID,
			WSOrigin:        ".*",
			EnableBasicAuth: sm.config.Password != "",
			Credential:      sm.config.AuthName + ":" + sm.config.Password,
			Command:         command,
			Args:            args,
		}
		gottyService := services.NewGottyService(gottyConfig, sm.ctx)
		err := gottyService.Start()
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
		return sm.handleSignals()
	}, func(error) {
		sm.cancel()
	})

	// Start notification watcher (only if tmux is available)
	tmuxService := NewTmuxService(sm.config.GetSessionID())
	if tmuxService.IsAvailable() {
		g.Add(func() error {
			fmt.Printf("🔔 Starting notification watcher...\n")
			watcher := NewTmuxWatcher(sm.config.GetSessionID(), sm.notifier, sm.ctx)
			if err := watcher.Start(); err != nil {
				log.Printf("Notification watcher stopped: %v", err)
			}
			return nil
		}, func(error) {
			// Watcher will stop automatically when context is cancelled
		})
	}

	// 24-hour timeout - only enable when AutoExit is true
	if sm.config.AutoExit {
		g.Add(func() error {
			timeoutHours := platform.AutoExitTimeoutHours
			timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutHours)*time.Hour)
			defer cancel()

			select {
			case <-timeoutCtx.Done():
				fmt.Printf("\n⏰ Service has been running for %d hours, stopping...\n", timeoutHours)
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

	// Save session info
	info := &SessionInfo{
		SessionID: sessionID,
		PID:       os.Getpid(),
		Port:      sm.config.GottyPort,
		StartTime: time.Now(),
		Config:    sm.config,
	}
	if err := saveSessionInfo(info); err != nil {
		fmt.Printf("Warning: failed to save session info: %v\n", err)
	}
	defer removeSessionInfo(sessionID)

	fmt.Printf("✅ Services started successfully!\n")

	// Construct remote access URL
	remoteURL := fmt.Sprintf("%s/%s", strings.TrimRight(sm.config.GetHTTPURL(), "/"), sm.config.GetSessionID())
	fmt.Printf("🌐 Access URL: %s\n", remoteURL)
	fmt.Printf("🔒 Local URL: http://localhost:%d\n", sm.config.GottyPort)

	// Show attached ports
	if len(sm.config.AttachPorts) > 0 {
		fmt.Printf("📌 Attached ports:\n")
		for _, port := range sm.config.AttachPorts {
			attachURL := fmt.Sprintf("%s/%s/%d", strings.TrimRight(sm.config.GetHTTPURL(), "/"), sm.config.GetSessionID(), port)
			fmt.Printf("   - Port %d -> %s\n", port, attachURL)
		}
	}
	
	if sm.config.Password != "" {
		fmt.Printf("🔐 HTTP auth: username=%s, password=%s\n", sessionID, sm.config.Password)
	} else {
		fmt.Printf("⚠️  HTTP authentication not enabled\n")
	}
	fmt.Printf("Press Ctrl+C to stop services\n")

	// Run all services
	return g.Run()
}

// handleSignals handles OS signals for graceful shutdown
func (sm *ServiceManager) handleSignals() error {
	c := make(chan os.Signal, 1)

	// Set different signals based on operating system
	signals := platform.GetStopSignals()
	signal.Notify(c, signals...)

	select {
	case sig := <-c:
		fmt.Printf("\n🛑 Received stop signal %v, shutting down services...\n", sig)

		// Kill tmux session if it exists
		CleanupTmuxSession(sm.config.GetSessionID())

		sm.cancel() // Cancel context immediately
		return nil
	case <-sm.ctx.Done():
		// Context cancelled elsewhere (e.g. timeout), also cleanup
		CleanupTmuxSession(sm.config.GetSessionID())
		return sm.ctx.Err()
	}
}

// Stop stops all services
func (sm *ServiceManager) Stop() {
	fmt.Printf("✅ Services stopped\n")
}

// getClaudeCodeCommand builds the claude-code command with flags and environment variables
func (sm *ServiceManager) getClaudeCodeCommand() (string, []string) {
	var command string
	finder := commands.NewFinder(sm.config.CodeCmd)

	// If custom command is specified and not "claude", use it
	if customCmd, ok := finder.FindCustomCommand(); ok {
		command = customCmd
	} else {
		// Use finder to locate the appropriate command
		command = finder.FindCommand()
	}

	args := []string{}

	// Add flags if provided
	if sm.config.Flags != "" {
		// Parse flags string (e.g., "--model opus --fast-model sonnet")
		flagParts := strings.Fields(sm.config.Flags)
		args = append(args, flagParts...)
	}

	// Use tmux for persistent sessions
	// According to gotty docs: "gotty tmux new -A -s gotty top"
	tmuxService := NewTmuxService(sm.config.GetSessionID())
	if tmuxService.IsAvailable() {
		tmuxPath, tmuxArgs, err := tmuxService.WrapCommand(command, args, sm.config.EnvVars)
		if err == nil {
			return tmuxPath, tmuxArgs
		}
		// If tmux wrapping fails, fall through to direct command
		fmt.Printf("Warning: tmux wrapping failed: %v, using direct command\n", err)
	}

	// Fallback: return command without tmux
	return command, args
}

// daemonize forks the process to background and starts services
func (sm *ServiceManager) daemonize() error {
	if !platform.SupportsDaemon() {
		return fmt.Errorf("daemon mode is not supported on this platform")
	}

	// Get the path to save the log file
	logFilePath, err := platform.GetLogFilePath()
	if err != nil {
		return fmt.Errorf("failed to get log file path: %w", err)
	}

	// Get the current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build arguments from config
	args := sm.config.ToArgs()

	// Create a new process that will run in background
	cmd := exec.Command(execPath, args...)

	// Open log file for child process
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Redirect child's stdout and stderr to log file
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// Don't set stdin - let child inherit /dev/null

	// Set process to be detached from terminal (Unix only)
	if platform.IsWindows() {
		// Windows doesn't support daemon mode
	} else {
		// Don't use Setsid - it causes issues with file descriptors
		// Just run in background without controlling terminal
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true, // Create new process group
		}
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Don't close logFile immediately - let child process keep it open
	// The file will be closed when parent process exits

	// Print completion message
	pid := cmd.Process.Pid
	fmt.Printf("✅ Service started in background (daemon mode)\n")
	fmt.Printf("   Process ID: %d\n", pid)
	fmt.Printf("🌐 Access URL: http://localhost:%d\n", sm.config.GottyPort)
	if sm.config.Password != "" {
		fmt.Printf("🔐 HTTP auth: username=%s, password=%s\n", sm.config.GetSessionID(), sm.config.Password)
	}
	fmt.Printf("Session: %s\n", sm.config.GetSessionID())
	fmt.Printf("To stop: clauded session kill %s\n", sm.config.GetSessionID())

	return nil
}

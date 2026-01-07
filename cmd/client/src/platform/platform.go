package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

const (
	// DefaultHost is the default public server host
	DefaultHost = "clauded.friddle.me"
	// DefaultHostWWW is the default public server host with www
	DefaultHostWWW = "www.clauded.friddle.me"
	// DefaultPort is the default piko server port
	DefaultPort = 8022
	// DefaultGottyPortStart is the start port for gotty auto-allocation
	DefaultGottyPortStart = 8080
	// GottyPortRange is the range of ports to try for gotty
	GottyPortRange = 100
	// DefaultRemotePort is the default remote port if not specified
	DefaultRemotePort = 8088
	// SessionIDLength is the length of auto-generated short session IDs
	SessionIDLength = 5
	// PasswordLength is the length of auto-generated passwords
	PasswordLength = 5
	// AutoExitTimeout is the auto-exit timeout duration (24 hours)
	AutoExitTimeoutHours = 24
)

// IsWindows checks if the current OS is Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsDarwin checks if the current OS is macOS
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux checks if the current OS is Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// SupportsDaemon checks if daemon mode is supported on the current platform
func SupportsDaemon() bool {
	return !IsWindows()
}

// GetConfigDirs returns the config directories for the current platform
func GetConfigDirs() []string {
	configDirs := []string{}
	home, err := os.UserHomeDir()
	if err != nil {
		return configDirs
	}

	if IsDarwin() {
		configDirs = append(configDirs,
			filepath.Join(home, "Library", "Application Support", "Claude Code"),
			filepath.Join(home, "Library", "Application Support", "claude-code"),
		)
	} else {
		// Linux/others
		configDirs = append(configDirs,
			filepath.Join(home, ".config", "claude-code"),
			filepath.Join(home, ".config", "Claude Code"),
		)
	}

	// Fallback/Common
	configDirs = append(configDirs,
		filepath.Join(home, ".claude-code"),
		filepath.Join(home, ".claude"),
	)

	return configDirs
}

// GetConfigFiles returns common config filenames to check
func GetConfigFiles() []string {
	return []string{"config.json", "credentials.json", "auth.json", "session.json"}
}

// GetStopSignals returns the appropriate stop signals for the current platform
func GetStopSignals() []os.Signal {
	if IsWindows() {
		// Windows supports Ctrl+C (SIGINT) and Ctrl+Break
		return []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}
	// Unix-like systems support more signals
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
}

// IsProcessRunning checks if a process with the given PID is running
func IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, FindProcess always succeeds, so we need to send signal 0 to check
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return true
}

// FindTmux checks if tmux is available and returns its path
func FindTmux() (string, error) {
	return exec.LookPath("tmux")
}

// IsDefaultHost checks if the host is the default public server
func IsDefaultHost(host string) bool {
	// Normalize host: remove http/https prefix
	normalizedHost := host
	if len(host) > 8 && host[:8] == "https://" {
		normalizedHost = host[8:]
	} else if len(host) > 7 && host[:7] == "http://" {
		normalizedHost = host[7:]
	}

	return normalizedHost == DefaultHost || normalizedHost == DefaultHostWWW
}

// GetLogFilePath returns the path to the log file
func GetLogFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".clauded.log"), nil
}

// GetSessionDir returns the directory where session info is stored
func GetSessionDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	dir := filepath.Join(homeDir, ".clauded", "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create session directory: %w", err)
	}
	return dir, nil
}

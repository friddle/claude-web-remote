package src

import (
	"fmt"
	"os/exec"
	"strings"

	"clauded-client/src/platform"
)

// TmuxService manages tmux session wrapping
type TmuxService struct {
	sessionID string
}

// NewTmuxService creates a new tmux service
func NewTmuxService(sessionID string) *TmuxService {
	return &TmuxService{
		sessionID: sessionID,
	}
}

// WrapCommand wraps the command in a tmux session
// Returns: (tmuxPath, tmuxArgs, error)
// If tmux is not available, returns empty strings
func (ts *TmuxService) WrapCommand(command string, args []string, envVars []string) (string, []string, error) {
	tmuxPath, err := platform.FindTmux()
	if err != nil {
		return "", nil, fmt.Errorf("tmux not found: %w", err)
	}

	// Build environment variable prefix string
	// e.g. env VAR1='val1' VAR2='val2'
	envPrefix := ""
	if len(envVars) > 0 {
		envPrefix = "env"
		for _, env := range envVars {
			// Split key=value
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				val := parts[1]
				// Simple shell escaping for single quotes
				val = strings.ReplaceAll(val, "'", "'\\''")
				envPrefix += fmt.Sprintf(" %s='%s'", key, val)
			}
		}
		envPrefix += " "
	}

	// Build command with args
	fullCommand := command
	for _, arg := range args {
		// Escape args as well just in case
		arg = strings.ReplaceAll(arg, "'", "'\\''")
		fullCommand += fmt.Sprintf(" '%s'", arg)
	}

	// Combine env + command
	finalCommand := envPrefix + fullCommand

	// Use 'tmux new-session -A' which will attach to existing session or create new one
	// The -A flag means "attach if exists, otherwise create"
	// This ensures session persistence across reconnects
	tmuxArgs := []string{
		"new-session", "-A", "-s", ts.sessionID,
		finalCommand,
	}

	fmt.Printf("✓ Using tmux persistent session: %s\n", ts.sessionID)
	return tmuxPath, tmuxArgs, nil
}

// CreateDetachedSession creates a detached tmux session running the command
func (ts *TmuxService) CreateDetachedSession(command string) error {
	tmuxPath, err := platform.FindTmux()
	if err != nil {
		return fmt.Errorf("tmux not found: %w", err)
	}

	// Check if session already exists
	checkCmd := exec.Command(tmuxPath, "has-session", "-t", ts.sessionID)
	if checkCmd.Run() == nil {
		fmt.Printf("✓ Tmux session already exists: %s\n", ts.sessionID)
		return nil
	}

	// Create new detached session
	createArgs := []string{
		"new-session", "-d", "-s", ts.sessionID,
		command,
	}
	createCmd := exec.Command(tmuxPath, createArgs...)
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	fmt.Printf("✓ Created detached tmux session: %s\n", ts.sessionID)
	return nil
}

// IsAvailable checks if tmux is available
func (ts *TmuxService) IsAvailable() bool {
	_, err := platform.FindTmux()
	return err == nil
}

// KillSession kills the tmux session if it exists
func (ts *TmuxService) KillSession() {
	if !ts.IsAvailable() {
		return
	}

	exec.Command("tmux", "kill-session", "-t", ts.sessionID).Run()
	fmt.Printf("🧹 Cleaned up tmux session: %s\n", ts.sessionID)
}

// CleanupOnSignal cleans up tmux session when receiving a signal
func CleanupTmuxSession(sessionID string) {
	if _, err := exec.LookPath("tmux"); err == nil {
		exec.Command("tmux", "kill-session", "-t", sessionID).Run()
		fmt.Printf("🧹 Cleaned up tmux session: %s\n", sessionID)
	}
}

// IsTmuxAvailable is a convenience function to check tmux availability
func IsTmuxAvailable() bool {
	_, err := platform.FindTmux()
	return err == nil
}

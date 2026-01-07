package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Finder handles command discovery and verification
type Finder struct {
	codeCmd string
}

// NewFinder creates a new command finder
func NewFinder(codeCmd string) *Finder {
	return &Finder{codeCmd: codeCmd}
}

// FindCommand searches for and returns the appropriate AI command tool
func (f *Finder) FindCommand() string {
	// Map of supported commands
	commandMap := map[string]string{
		"claude":    "claude",
		"opencode":  "opencode",
		"kimi":      "kimi",
		"gemini":    "gemini",
	}

	// Get the actual command name
	cmdName, ok := commandMap[f.codeCmd]
	if !ok {
		fmt.Printf("⚠️  Unknown codecmd: %s, falling back to 'claude'\n", f.codeCmd)
		cmdName = "claude"
	}

	// Special handling for 'claude' command (check multiple variants)
	if f.codeCmd == "claude" {
		return f.findClaudeCommand()
	}

	// For other commands, just check if they exist in PATH
	if path, err := exec.LookPath(cmdName); err == nil {
		fmt.Printf("✓ Using %s command from: %s\n", f.codeCmd, path)
		return cmdName
	}

	fmt.Printf("⚠️  Warning: %s command not found in PATH\n", cmdName)
	return cmdName
}

// findClaudeCommand searches for claude/claude-code command with priority
func (f *Finder) findClaudeCommand() string {
	// Priority 1: Check if 'claude' command exists in PATH
	if path, err := exec.LookPath("claude"); err == nil {
		fmt.Printf("✓ Using claude command from: %s\n", path)
		if f.IsClaudeCodeCommand("claude") {
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

// IsClaudeCodeCommand verifies if the command is actually Claude Code
func (f *Finder) IsClaudeCodeCommand(cmd string) bool {
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

// IsInstalled checks if the specified command is installed
func (f *Finder) IsInstalled() bool {
	// First check for 'claude' command (priority)
	if _, err := exec.LookPath("claude"); err == nil {
		// Verify it's actually Claude Code
		if f.IsClaudeCodeCommand("claude") {
			return true
		}
	}

	// Fallback to 'claude-code' command
	_, err := exec.LookPath("claude-code")
	return err == nil
}

// GetVersion returns the installed claude-code version
func (f *Finder) GetVersion() (string, error) {
	// Try 'claude' command first
	if _, err := exec.LookPath("claude"); err == nil {
		if f.IsClaudeCodeCommand("claude") {
			cmd := exec.Command("claude", "--version")
			output, err := cmd.Output()
			if err == nil {
				return strings.TrimSpace(string(output)), nil
			}
		}
	}

	// Fallback to 'claude-code'
	cmd := exec.Command("claude-code", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get claude-code version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// FindCustomCommand finds a custom command if specified
func (f *Finder) FindCustomCommand() (string, bool) {
	if f.codeCmd == "" || f.codeCmd == "claude" {
		return "", false
	}

	if path, err := exec.LookPath(f.codeCmd); err == nil {
		fmt.Printf("✓ Using custom command: %s\n", path)
		return path, true
	}

	return "", false
}

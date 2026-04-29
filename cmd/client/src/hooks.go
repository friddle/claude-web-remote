package src

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type hookEntry struct {
	Notification []hookCommand `json:"Notification,omitempty"`
}

type claudeSettings struct {
	Hooks hookEntry `json:"hooks"`
}

func getNotifyCommand() string {
	if runtime.GOOS == "darwin" {
		return `osascript -e 'display notification "AI 正在等待你的指令..." with title "Claude 需要授权" sound name "default"'`
	}
	return `notify-send -u critical -i dialog-question 'Claude 需要授权' 'AI 正在等待你的指令...'`
}

func SetupClaudeCodeHooks(codeCmd string) error {
	if codeCmd != "claude" {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %w", err)
	}

	configDir := filepath.Join(home, ".claude")
	settingsFile := filepath.Join(configDir, "settings.json")

	notifyCmd := getNotifyCommand()

	var settings claudeSettings

	existing, err := os.ReadFile(settingsFile)
	if err == nil {
		if err := json.Unmarshal(existing, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %w", err)
		}
	}

	settings.Hooks.Notification = []hookCommand{
		{
			Type:    "command",
			Command: notifyCmd,
		},
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	fmt.Printf("✅ Claude Code notification hooks configured: %s\n", settingsFile)
	return nil
}

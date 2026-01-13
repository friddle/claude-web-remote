package src

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"time"
)

// TmuxWatcher monitors a tmux session and sends notifications
type TmuxWatcher struct {
	sessionName string
	notifier    *Notifier
	detector    *TaskDetector
	ctx         context.Context
	interval    time.Duration
}

// NewTmuxWatcher creates a new tmux watcher
func NewTmuxWatcher(sessionName string, notifier *Notifier, ctx context.Context) *TmuxWatcher {
	detector := NewTaskDetector(notifier)
	return &TmuxWatcher{
		sessionName: sessionName,
		notifier:    notifier,
		detector:    detector,
		ctx:         ctx,
		interval:    2 * time.Second, // Check every 2 seconds
	}
}

// Start begins monitoring the tmux session
func (tw *TmuxWatcher) Start() error {
	log.Printf("Starting tmux watcher for session: %s", tw.sessionName)

	// Wait a bit for tmux session to be ready
	time.Sleep(3 * time.Second)

	ticker := time.NewTicker(tw.interval)
	defer ticker.Stop()

	var lastCaptured string

	for {
		select {
		case <-tw.ctx.Done():
			log.Println("Tmux watcher stopped")
			return nil
		case <-ticker.C:
			// Capture output from tmux session
			output, err := tw.captureOutput()
			if err != nil {
				// Session might not be ready yet, log and continue
				log.Printf("Failed to capture tmux output: %v", err)
				continue
			}

			// Only process new output
			if output != lastCaptured && output != "" {
				// Detect only new lines
				lines := strings.Split(output, "\n")
				newLines := lines

				if lastCaptured != "" {
					lastLines := strings.Split(lastCaptured, "\n")
					// Find where new content starts
					for i, line := range lastLines {
						if i < len(lines) && lines[i] == line {
							newLines = lines[i+1:]
							break
						}
					}
				}

				// Process new lines
				newOutput := strings.Join(newLines, "\n")
				if newOutput != "" {
					tw.processOutput(newOutput)
				}

				lastCaptured = output
			}
		}
	}
}

// captureOutput captures the current tmux session output
func (tw *TmuxWatcher) captureOutput() (string, error) {
	// Use tmux capture-pane to get output
	cmd := exec.Command("tmux", "capture-pane", "-t", tw.sessionName, "-p", "-S", "-1000")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// processOutput processes the captured output for task detection
func (tw *TmuxWatcher) processOutput(output string) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for task completion
		if tw.detector.detectCompletion(line) {
			log.Printf("✓ Task completion detected: %s", line)
			if err := tw.notifier.PublishTaskCompleted("Task Completed", line); err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}

		// Check for errors
		if tw.detector.detectError(line) {
			log.Printf("✗ Error detected: %s", line)
			if err := tw.notifier.PublishError("Error Detected", line); err != nil {
				log.Printf("Failed to send error notification: %v", err)
			}
		}
	}
}

// IsSessionActive checks if the tmux session is active
func (tw *TmuxWatcher) IsSessionActive() bool {
	cmd := exec.Command("tmux", "has-session", "-t", tw.sessionName)
	err := cmd.Run()
	return err == nil
}

// GetSessionInfo returns information about the tmux session
func (tw *TmuxWatcher) GetSessionInfo() (string, error) {
	cmd := exec.Command("tmux", "list-panes", "-t", tw.sessionName, "-F", "#{pane_width}x#{pane_height}")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

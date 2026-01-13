package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// NotificationType notification type
type NotificationType string

const (
	TaskCompleted NotificationType = "task_completed"
	Error         NotificationType = "error"
	Progress      NotificationType = "progress"
	SystemStatus  NotificationType = "system_status"
)

// Notification notification message
type Notification struct {
	SessionID string                 `json:"session_id"`
	Type      NotificationType        `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

// PublishRequest notification publish request
type PublishRequest struct {
	SessionID string                 `json:"session_id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
}

// Notifier sends notifications to the server
type Notifier struct {
	serverURL    string
	sessionID    string
	httpClient   *http.Client
	enabled      bool
}

// NewNotifier creates a new notifier
func NewNotifier(serverURL, sessionID string) *Notifier {
	return &Notifier{
		serverURL: serverURL,
		sessionID: sessionID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		enabled: true,
	}
}

// Publish sends a notification to the server
func (n *Notifier) Publish(notifType NotificationType, data map[string]interface{}) error {
	if !n.enabled {
		return nil
	}

	// Construct notification URL
	notifyURL := fmt.Sprintf("%s/api/v1/notifications/publish", n.serverURL)

	// Build request payload
	req := PublishRequest{
		SessionID: n.sessionID,
		Type:      string(notifType),
		Data:      data,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Send POST request
	resp, err := n.httpClient.Post(notifyURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notification failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✓ Notification sent: type=%s, session=%s", notifType, n.sessionID)
	return nil
}

// PublishTaskCompleted sends a task completion notification
func (n *Notifier) PublishTaskCompleted(taskName, output string) error {
	return n.Publish(TaskCompleted, map[string]interface{}{
		"task_name": taskName,
		"output":    output,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// PublishError sends an error notification
func (n *Notifier) PublishError(errorMsg, details string) error {
	return n.Publish(Error, map[string]interface{}{
		"error":   errorMsg,
		"details": details,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// PublishProgress sends a progress update notification
func (n *Notifier) PublishProgress(message string, percentage int) error {
	return n.Publish(Progress, map[string]interface{}{
		"message":    message,
		"percentage": percentage,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// SetEnabled enables or disables notifications
func (n *Notifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// IsEnabled returns whether notifications are enabled
func (n *Notifier) IsEnabled() bool {
	return n.enabled
}

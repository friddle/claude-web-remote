package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
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
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	Type      NotificationType        `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Subscriber notification subscriber
type Subscriber struct {
	ID       string
	SessionID string
	Channel  chan Notification
	WebhookURL string
	EventTypes []NotificationType
}

// Service notification service
type Service struct {
	subscribers map[string][]*Subscriber
	mu          sync.RWMutex
	notifyQueue chan Notification
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewService creates a new notification service
func NewService() *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		subscribers: make(map[string][]*Subscriber),
		notifyQueue: make(chan Notification, 1000),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start starts the notification service
func (s *Service) Start() {
	log.Println("Starting notification service...")
	go s.processNotifications()
}

// Stop stops the notification service
func (s *Service) Stop() {
	log.Println("Stopping notification service...")
	s.cancel()
	close(s.notifyQueue)
}

// Publish publishes a notification
func (s *Service) Publish(sessionID string, notifType NotificationType, data map[string]interface{}) {
	notification := Notification{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Type:      notifType,
		Data:      data,
		Timestamp: time.Now(),
	}

	select {
	case s.notifyQueue <- notification:
	default:
		log.Printf("Notification queue full, dropping notification for session %s", sessionID)
	}
}

// SubscribeSSE subscribes to notifications via SSE
func (s *Service) SubscribeSSE(sessionID string) chan Notification {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan Notification, 100)
	subscriber := &Subscriber{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Channel:   ch,
		EventTypes: []NotificationType{TaskCompleted, Error, Progress, SystemStatus},
	}

	s.subscribers[sessionID] = append(s.subscribers[sessionID], subscriber)
	return ch
}

// SubscribeWebhook subscribes to notifications via webhook
func (s *Service) SubscribeWebhook(sessionID, webhookURL string, eventTypes []NotificationType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	subscriber := &Subscriber{
		ID:         uuid.New().String(),
		SessionID:  sessionID,
		WebhookURL: webhookURL,
		EventTypes: eventTypes,
	}

	s.subscribers[sessionID] = append(s.subscribers[sessionID], subscriber)
	log.Printf("Subscribed webhook for session %s: %s", sessionID, webhookURL)
	return nil
}

// Unsubscribe removes a subscriber
func (s *Service) Unsubscribe(sessionID, subscriberID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs := s.subscribers[sessionID]
	for i, sub := range subs {
		if sub.ID == subscriberID {
			// Remove subscriber from slice
			s.subscribers[sessionID] = append(subs[:i], subs[i+1:]...)
			close(sub.Channel)
			break
		}
	}
}

// processNotifications processes notifications from the queue
func (s *Service) processNotifications() {
	for {
		select {
		case notif, ok := <-s.notifyQueue:
			if !ok {
				return
			}
			s.distributeNotification(notif)
		case <-s.ctx.Done():
			return
		}
	}
}

// distributeNotification distributes notification to all subscribers
func (s *Service) distributeNotification(notif Notification) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subs, exists := s.subscribers[notif.SessionID]
	if !exists {
		return
	}

	for _, sub := range subs {
		// Check if subscriber is interested in this event type
		if !s.isEventTypeMatch(notif.Type, sub.EventTypes) {
			continue
		}

		// Send to SSE subscribers
		if sub.Channel != nil {
			select {
			case sub.Channel <- notif:
			default:
				log.Printf("Subscriber channel full for session %s", sub.SessionID)
			}
		}

		// Send to webhook subscribers
		if sub.WebhookURL != "" {
			go s.sendWebhook(sub.WebhookURL, notif)
		}
	}
}

// isEventTypeMatch checks if event type matches subscriber's interests
func (s *Service) isEventTypeMatch(eventType NotificationType, eventTypes []NotificationType) bool {
	for _, et := range eventTypes {
		if et == eventType {
			return true
		}
	}
	return false
}

// sendWebhook sends notification to webhook URL
func (s *Service) sendWebhook(url string, notif Notification) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	data, err := json.Marshal(notif)
	if err != nil {
		log.Printf("Failed to marshal notification: %v", err)
		return
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Printf("Failed to send webhook to %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Webhook returned non-OK status: %d", resp.StatusCode)
	}
}

// GetSubscribers returns all subscribers for a session
func (s *Service) GetSubscribers(sessionID string) []*Subscriber {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subs, exists := s.subscribers[sessionID]
	if !exists {
		return []*Subscriber{}
	}

	// Return a copy to avoid race conditions
	result := make([]*Subscriber, len(subs))
	copy(result, subs)
	return result
}

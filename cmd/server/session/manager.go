package session

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session session information
type Session struct {
	ID        string
	CreatedAt time.Time
	LastSeen  time.Time
	Metadata  map[string]interface{}
	mu        sync.RWMutex
}

// Manager session manager
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewManager creates a new session manager
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]*Session),
	}
}

// Create creates a new session
func (m *Manager) Create() *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	m.sessions[session.ID] = session
	return session
}

// Get gets a session by ID
func (m *Manager) Get(id string) (*Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[id]
	if !exists {
		return nil, false
	}

	// Update last seen
	session.mu.Lock()
	session.LastSeen = time.Now()
	session.mu.Unlock()

	return session, true
}

// Delete deletes a session
func (m *Manager) Delete(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
}

// UpdateMetadata updates session metadata
func (m *Manager) UpdateMetadata(id string, metadata map[string]interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return false
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	for k, v := range metadata {
		session.Metadata[k] = v
	}
	return true
}

// Cleanup cleans up expired sessions
func (m *Manager) Cleanup(timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for id, session := range m.sessions {
		session.mu.RLock()
		if now.Sub(session.LastSeen) > timeout {
			delete(m.sessions, id)
		}
		session.mu.RUnlock()
	}
}

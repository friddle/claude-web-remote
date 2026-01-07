package src

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"clauded-client/src/platform"
)

// SessionInfo stores information about a running daemon session
type SessionInfo struct {
	SessionID string    `json:"session_id"`
	PID       int       `json:"pid"`
	Port      int       `json:"port"`
	StartTime time.Time `json:"start_time"`
	Config    *Config   `json:"config"`
}

// getSessionDir returns the directory where session info is stored
func getSessionDir() (string, error) {
	return platform.GetSessionDir()
}

// saveSessionInfo saves the session information to disk
func saveSessionInfo(info *SessionInfo) error {
	dir, err := getSessionDir()
	if err != nil {
		return err
	}

	file := filepath.Join(dir, info.SessionID+".json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, data, 0644)
}

// loadSessionInfo loads session information from disk
func loadSessionInfo(sessionID string) (*SessionInfo, error) {
	dir, err := getSessionDir()
	if err != nil {
		return nil, err
	}

	file := filepath.Join(dir, sessionID+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var info SessionInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// removeSessionInfo removes the session information from disk
func removeSessionInfo(sessionID string) error {
	dir, err := getSessionDir()
	if err != nil {
		return err
	}

	file := filepath.Join(dir, sessionID+".json")
	return os.Remove(file)
}

// ListSessions lists all running sessions
func ListSessions() {
	dir, err := getSessionDir()
	if err != nil {
		fmt.Printf("Error getting session directory: %v\n", err)
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("Error reading session directory: %v\n", err)
		return
	}

	var sessions []*SessionInfo
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			sessionID := strings.TrimSuffix(entry.Name(), ".json")
			info, err := loadSessionInfo(sessionID)
			if err != nil {
				continue
			}
			
			// Check if process is still running
			if isProcessRunning(info.PID) {
				sessions = append(sessions, info)
			} else {
				// Clean up stale session file
				removeSessionInfo(sessionID)
			}
		}
	}

	if len(sessions) == 0 {
		fmt.Println("No active sessions found.")
		return
	}

	// Sort by start time
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.Before(sessions[j].StartTime)
	})

	fmt.Println("Active clauded sessions:")
	fmt.Printf("%-10s %-8s %-8s %-25s %s\n", "SESSION ID", "PID", "PORT", "STARTED", "FLAGS")
	fmt.Println(strings.Repeat("-", 80))

	for _, s := range sessions {
		flags := s.Config.Flags
		if flags == "" {
			flags = "(default)"
		}
		fmt.Printf("%-10s %-8d %-8d %-25s %s\n", 
			s.SessionID, 
			s.PID, 
			s.Port, 
			s.StartTime.Format("2006-01-02 15:04:05"), 
			flags)
	}
}

// KillSession kills a specific session or all sessions
func KillSession(target string) {
	dir, err := getSessionDir()
	if err != nil {
		fmt.Printf("Error getting session directory: %v\n", err)
		return
	}

	if target == "all" {
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Printf("Error reading session directory: %v\n", err)
			return
		}

		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				sessionID := strings.TrimSuffix(entry.Name(), ".json")
				if killSingleSession(sessionID) {
					count++
				}
			}
		}
		fmt.Printf("Killed %d sessions.\n", count)
	} else {
		if killSingleSession(target) {
			fmt.Printf("Session %s killed.\n", target)
		} else {
			fmt.Printf("Failed to kill session %s (not found or error).\n", target)
		}
	}
}

// killSingleSession kills a single session by ID
func killSingleSession(sessionID string) bool {
	info, err := loadSessionInfo(sessionID)
	if err != nil {
		return false
	}

	// Kill the process
	process, err := os.FindProcess(info.PID)
	if err == nil {
		// Use SIGTERM for graceful shutdown
		process.Signal(syscall.SIGTERM)
	}

	// Remove the file
	removeSessionInfo(sessionID)
	return true
}

// isProcessRunning checks if a process is running
func isProcessRunning(pid int) bool {
	return platform.IsProcessRunning(pid)
}

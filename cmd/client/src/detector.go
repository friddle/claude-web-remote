package src

import (
	"bufio"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
)

// TaskDetector detects task completion from output
type TaskDetector struct {
	notifier      *Notifier
	completionPatterns []string
	errorPatterns     []string
	lastOutput        string
	mu                sync.Mutex
}

// NewTaskDetector creates a new task detector
func NewTaskDetector(notifier *Notifier) *TaskDetector {
	return &TaskDetector{
		notifier: notifier,
		completionPatterns: []string{
			"✓",                    // Checkmark
			"✅",                   // Green checkmark
			"Done",                 // Done
			"Completed",            // Completed
			"Finished",             // Finished
			"Success",              // Success
			"Build successful",     // Build successful
			"Tests passed",         // Tests passed
			"All tests passed",     // All tests passed
			"Installation complete", // Installation complete
			"Deployment complete",   // Deployment complete
		},
		errorPatterns: []string{
			"Error:",              // Error:
			"ERROR",               // ERROR
			"Failed",              // Failed
			"Exception",           // Exception
			"fatal:",              // fatal:
			"Fatal error",         // Fatal error
			"panic:",              // panic:
			"Cannot find module",  // Node.js error
			"Compilation failed",  // Compilation error
		},
	}
}

// DetectFromReader reads output and detects task completion
func (td *TaskDetector) DetectFromReader(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	var buffer strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		log.Printf("[DETECTOR] %s", line)

		// Add to buffer
		buffer.WriteString(line + "\n")

		// Keep last 1000 characters in buffer
		output := buffer.String()
		if len(output) > 1000 {
			output = output[len(output)-1000:]
			buffer.Reset()
			buffer.WriteString(output)
		}

		// Detect task completion
		if td.detectCompletion(line) {
			td.mu.Lock()
			td.lastOutput = output
			td.mu.Unlock()

			// Send notification
			log.Println("✓ Task completion detected")
			if err := td.notifier.PublishTaskCompleted("Task Completed", output); err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}

		// Detect errors
		if td.detectError(line) {
			log.Println("✗ Error detected")
			if err := td.notifier.PublishError("Error Detected", output); err != nil {
				log.Printf("Failed to send error notification: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

// detectCompletion checks if the line contains a completion pattern
func (td *TaskDetector) detectCompletion(line string) bool {
	for _, pattern := range td.completionPatterns {
		if strings.Contains(line, pattern) {
			// Additional checks to avoid false positives
			// Skip if it's part of an error message or warning
			if !td.isErrorLine(line) && !td.isWarningLine(line) {
				return true
			}
		}
	}
	return false
}

// detectError checks if the line contains an error pattern
func (td *TaskDetector) detectError(line string) bool {
	for _, pattern := range td.errorPatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}
	return false
}

// isErrorLine checks if the line is an error message
func (td *TaskDetector) isErrorLine(line string) bool {
	errorIndicators := []string{
		"error:", "ERROR:", "Error:", "failed", "Failed", "exception",
	}
	for _, indicator := range errorIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}
	return false
}

// isWarningLine checks if the line is a warning message
func (td *TaskDetector) isWarningLine(line string) bool {
	warningIndicators := []string{
		"warning:", "WARNING:", "Warning:", "warn:",
	}
	for _, indicator := range warningIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}
	return false
}

// DetectCompletionFromString detects completion from a string
func (td *TaskDetector) DetectCompletionFromString(output string) bool {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if td.detectCompletion(line) {
			return true
		}
	}
	return false
}

// AddCompletionPattern adds a custom completion pattern
func (td *TaskDetector) AddCompletionPattern(pattern string) {
	td.completionPatterns = append(td.completionPatterns, pattern)
}

// AddErrorPattern adds a custom error pattern
func (td *TaskDetector) AddErrorPattern(pattern string) {
	td.errorPatterns = append(td.errorPatterns, pattern)
}

// AddRegexPattern adds a regex pattern for detection
func (td *TaskDetector) AddRegexPattern(pattern string, isCompletion bool) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	if isCompletion {
		td.completionPatterns = append(td.completionPatterns, pattern)
		_ = re // Use regex for matching (implementation can be added later)
	} else {
		td.errorPatterns = append(td.errorPatterns, pattern)
	}

	return nil
}

// GetLastOutput returns the last detected output
func (td *TaskDetector) GetLastOutput() string {
	td.mu.Lock()
	defer td.mu.Unlock()
	return td.lastOutput
}

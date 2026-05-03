// Package log provides structured error logging to rotating daily files.
package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/torosent/c9s/internal/clock"
)

// Entry represents a single error log entry.
type Entry struct {
	Time     time.Time `json:"time"`
	Op       string    `json:"op"`
	Resource string    `json:"resource"`
	Message  string    `json:"message"`
	Detail   string    `json:"detail"`
}

// Logger writes structured error entries to rotating daily log files.
type Logger struct {
	dir     string
	clock   clock.Clock
	mu      sync.Mutex
	current *os.File
	date    string
}

// New creates a new error logger that writes to the given directory.
// Log files are named errors-<YYYY-MM-DD>.log and rotate automatically.
func New(dir string, c clock.Clock) (*Logger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}

	l := &Logger{
		dir:   dir,
		clock: c,
	}

	// Open initial file
	if err := l.rotate(); err != nil {
		return nil, err
	}

	return l, nil
}

// Log writes an error entry to the current log file.
// Automatically rotates to a new file on date change.
func (l *Logger) Log(e Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if date changed
	currentDate := l.clock.Now().Format("2006-01-02")
	if currentDate != l.date {
		if err := l.rotate(); err != nil {
			return err
		}
	}

	// Encode entry as JSON
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}

	// Write line
	if _, err := l.current.Write(data); err != nil {
		return fmt.Errorf("write entry: %w", err)
	}
	if _, err := l.current.Write([]byte("\n")); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}

	// Sync to disk
	if err := l.current.Sync(); err != nil {
		return fmt.Errorf("sync file: %w", err)
	}

	return nil
}

// rotate opens a new log file for the current date.
func (l *Logger) rotate() error {
	// Close previous file
	if l.current != nil {
		if err := l.current.Close(); err != nil {
			return fmt.Errorf("close previous file: %w", err)
		}
	}

	// Open new file
	l.date = l.clock.Now().Format("2006-01-02")
	path := filepath.Join(l.dir, fmt.Sprintf("errors-%s.log", l.date))

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	l.current = f
	return nil
}

// Close closes the logger and its current file.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.current != nil {
		return l.current.Close()
	}
	return nil
}

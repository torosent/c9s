package log_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/log"
)

func TestLogger_New(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))

	logger, err := log.New(dir, clk)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if logger == nil {
		t.Fatal("New() returned nil logger")
	}

	// Should create file
	expected := filepath.Join(dir, "errors-2026-05-02.log")
	if _, err := os.Stat(expected); err != nil {
		t.Fatalf("expected file %s not created: %v", expected, err)
	}
}

func TestLogger_Log(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))

	logger, err := log.New(dir, clk)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	entry := log.Entry{
		Time:     clk.Now(),
		Op:       "container.start",
		Resource: "api-server",
		Message:  "Failed to start container",
		Detail:   "port 8080 already in use",
	}

	if err := logger.Log(entry); err != nil {
		t.Fatalf("Log() error: %v", err)
	}

	// Read file and verify JSON
	path := filepath.Join(dir, "errors-2026-05-02.log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	var decoded log.Entry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	if decoded.Op != entry.Op {
		t.Errorf("Op = %q, want %q", decoded.Op, entry.Op)
	}
	if decoded.Resource != entry.Resource {
		t.Errorf("Resource = %q, want %q", decoded.Resource, entry.Resource)
	}
	if decoded.Message != entry.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, entry.Message)
	}
	if decoded.Detail != entry.Detail {
		t.Errorf("Detail = %q, want %q", decoded.Detail, entry.Detail)
	}
}

func TestLogger_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))

	logger, err := log.New(dir, clk)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	entries := []log.Entry{
		{Time: clk.Now(), Op: "op1", Resource: "res1", Message: "msg1", Detail: "detail1"},
		{Time: clk.Now(), Op: "op2", Resource: "res2", Message: "msg2", Detail: "detail2"},
		{Time: clk.Now(), Op: "op3", Resource: "res3", Message: "msg3", Detail: "detail3"},
	}

	for _, e := range entries {
		if err := logger.Log(e); err != nil {
			t.Fatalf("Log() error: %v", err)
		}
	}

	// Read file and verify all entries
	path := filepath.Join(dir, "errors-2026-05-02.log")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	lines := strings.Count(string(data), "\n")
	if lines != 3 {
		t.Errorf("got %d lines, want 3", lines)
	}
}

func TestLogger_DateRollover(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 23, 59, 0, 0, time.UTC))

	logger, err := log.New(dir, clk)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Log on day 1
	if err := logger.Log(log.Entry{Time: clk.Now(), Op: "op1", Message: "day1"}); err != nil {
		t.Fatalf("Log() day1 error: %v", err)
	}

	// Advance to next day
	clk.Advance(2 * time.Hour)

	// Log on day 2
	if err := logger.Log(log.Entry{Time: clk.Now(), Op: "op2", Message: "day2"}); err != nil {
		t.Fatalf("Log() day2 error: %v", err)
	}

	// Should have two files
	file1 := filepath.Join(dir, "errors-2026-05-02.log")
	file2 := filepath.Join(dir, "errors-2026-05-03.log")

	if _, err := os.Stat(file1); err != nil {
		t.Errorf("file1 not found: %v", err)
	}
	if _, err := os.Stat(file2); err != nil {
		t.Errorf("file2 not found: %v", err)
	}

	// Verify contents
	data1, _ := os.ReadFile(file1)
	if !strings.Contains(string(data1), "day1") {
		t.Errorf("file1 missing 'day1'")
	}
	if strings.Contains(string(data1), "day2") {
		t.Errorf("file1 should not contain 'day2'")
	}

	data2, _ := os.ReadFile(file2)
	if !strings.Contains(string(data2), "day2") {
		t.Errorf("file2 missing 'day2'")
	}
	if strings.Contains(string(data2), "day1") {
		t.Errorf("file2 should not contain 'day1'")
	}
}

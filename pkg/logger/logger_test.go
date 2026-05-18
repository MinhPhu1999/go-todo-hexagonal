package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWritesJSONLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "nested", "app.jsonl")

	appLogger, err := New(Config{
		Format:   "json",
		FilePath: logPath,
		Level:    "debug",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	appLogger.Info("logger ready", "component", "test")

	if err := appLogger.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	logs := string(content)
	if !strings.Contains(logs, `"msg":"logger ready"`) {
		t.Fatalf("log file does not contain message: %s", logs)
	}
	if !strings.Contains(logs, `"component":"test"`) {
		t.Fatalf("log file does not contain attributes: %s", logs)
	}
}

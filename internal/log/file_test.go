package log

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	logger, closer, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer closer.Close()

	logger.Info("test message")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Fatal("log file is empty")
	}
}

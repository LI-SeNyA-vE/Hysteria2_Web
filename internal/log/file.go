package log

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

func Open(path string) (*slog.Logger, io.Closer, error) {
	if path == "" {
		path = "./panel.log"
	}

	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, nil, fmt.Errorf("create log directory: %w", err)
		}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	handler := slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler), f, nil
}

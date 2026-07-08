package hysteria

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func (m *Manager) binaryPath() string {
	return filepath.Join(m.dataDir, "bin", "hysteria")
}

func (m *Manager) isBinaryInstalled() bool {
	_, err := os.Stat(m.binaryPath())
	return err == nil
}

func (m *Manager) downloadBinary(ctx context.Context) error {
	arch := runtime.GOARCH // amd64, arm64, ...
	goos := runtime.GOOS   // linux, darwin

	filename := fmt.Sprintf("hysteria-%s-%s", goos, arch)
	url := fmt.Sprintf(
		"https://github.com/apernet/hysteria/releases/download/%s/%s",
		hysteriaTag, filename,
	)

	binDir := filepath.Join(m.dataDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("создание bin директории: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("скачивание hysteria2: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub вернул HTTP %d (url: %s)", resp.StatusCode, url)
	}

	tmpPath := m.binaryPath() + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("создание временного файла: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("запись бинаря: %w", err)
	}
	f.Close()

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, m.binaryPath())
}

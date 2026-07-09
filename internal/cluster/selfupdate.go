package cluster

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const updateGithubRepo = "LI-SeNyA-vE/Hysteria2_Web"

// selfUpdate скачивает бинарь нужной версии, заменяет текущий и завершает процесс.
// systemd перезапустит его автоматически с новым бинарём.
func selfUpdate(ctx context.Context, targetVersion string) error {
	arch := runtime.GOARCH
	goos := runtime.GOOS
	assetName := fmt.Sprintf("panel-%s-%s", goos, arch)
	url := fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/%s",
		updateGithubRepo, targetVersion, assetName,
	)

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("путь к бинарю: %w", err)
	}

	log.Printf("nodeagent: скачиваю %s → %s", url, exePath)
	if err := downloadBinary(ctx, url, exePath+".new"); err != nil {
		return fmt.Errorf("скачивание: %w", err)
	}

	if err := os.Rename(exePath+".new", exePath); err != nil {
		if copyErr := copyBinary(exePath+".new", exePath); copyErr != nil {
			_ = os.Remove(exePath + ".new")
			return fmt.Errorf("замена бинаря: %w / %v", err, copyErr)
		}
		_ = os.Remove(exePath + ".new")
	}

	log.Printf("nodeagent: обновление до %s установлено, перезапуск...", targetVersion)
	go func() {
		time.Sleep(200 * time.Millisecond)
		os.Exit(0)
	}()
	return nil
}

func downloadBinary(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		_ = os.Remove(dest)
	}
	return err
}

func copyBinary(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	out.Close()
	return err
}

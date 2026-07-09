package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"hysteria2-web/internal/version"
)

const githubRepo = "LI-SeNyA-vE/Hysteria2_Web"

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func fetchLatestRelease(ctx context.Context) (*githubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github вернул %d", resp.StatusCode)
	}
	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func (s *Server) handleCheckUpdate(w http.ResponseWriter, r *http.Request) {
	rel, err := fetchLatestRelease(r.Context())
	if err != nil {
		writeErr(w, http.StatusBadGateway, "не удалось проверить обновления: "+err.Error())
		return
	}
	current := version.Version
	writeJSON(w, http.StatusOK, map[string]any{
		"currentVersion":  current,
		"latestVersion":   rel.TagName,
		"updateAvailable": rel.TagName != current && current != "dev",
		"releaseUrl":      rel.HTMLURL,
	})
}

func (s *Server) handleApplyUpdate(w http.ResponseWriter, r *http.Request) {
	rel, err := fetchLatestRelease(r.Context())
	if err != nil {
		writeErr(w, http.StatusBadGateway, "не удалось получить версию: "+err.Error())
		return
	}

	arch := runtime.GOARCH // amd64, arm64
	goos := runtime.GOOS  // linux, darwin
	assetName := fmt.Sprintf("panel-%s-%s", goos, arch)
	downloadURL := fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/%s",
		githubRepo, rel.TagName, assetName, // используем точную версию из API
	)

	// Путь к запущенному бинарю
	exePath, err := os.Executable()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "не удалось определить путь к бинарю: "+err.Error())
		return
	}

	// Скачиваем напрямую поверх текущего бинаря
	// (файл открыт на запись пока процесс держит его на чтение — на Linux это ок)
	if err := downloadFile(r.Context(), downloadURL, exePath+".new"); err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка скачивания: "+err.Error())
		return
	}

	// Атомарная замена через rename; если разные устройства — копируем вручную
	if err := os.Rename(exePath+".new", exePath); err != nil {
		if copyErr := copyFile(exePath+".new", exePath); copyErr != nil {
			os.Remove(exePath + ".new")
			writeErr(w, http.StatusInternalServerError, "ошибка замены бинаря: "+err.Error()+" / "+copyErr.Error())
			return
		}
		os.Remove(exePath + ".new")
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "обновление установлено, перезапуск...",
		"version": rel.TagName,
	})

	// Graceful exit — systemd/supervisor перезапустит с новым бинарём
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
}

// handlePushNodeUpdate устанавливает желаемую версию для всех нод.
// Ноды получат её в следующем heartbeat и обновятся сами.
func (s *Server) handlePushNodeUpdate(w http.ResponseWriter, r *http.Request) {
	rel, err := fetchLatestRelease(r.Context())
	if err != nil {
		writeErr(w, http.StatusBadGateway, "не удалось получить версию: "+err.Error())
		return
	}
	if err := s.db.SetSetting("desired_node_version", rel.TagName); err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка сохранения: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"version": rel.TagName,
		"message": "ноды обновятся в течение 10 секунд",
	})
}

func copyFile(src, dst string) error {
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

func downloadFile(ctx context.Context, url, destPath string) error {
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
		return fmt.Errorf("HTTP %d при скачивании %s", resp.StatusCode, url)
	}
	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(destPath)
	}
	return err
}

package hysteria

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"hysteria2-web/internal/models"
)

// PollTrafficForever каждые 10 секунд опрашивает Traffic Stats API hysteria2
// и добавляет дельты в traffic_used_bytes пользователей.
func (m *Manager) PollTrafficForever(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if m.IsRunning() {
				m.pollOnce()
			}
		}
	}
}

func (m *Manager) pollOnce() {
	secret, err := m.db.GetSetting(models.SettingStatsSecret)
	if err != nil || secret == "" {
		return
	}

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:9999/traffic?clear=1", nil)
	if err != nil {
		return
	}
	// Hysteria2 Traffic Stats API: Authorization без "Bearer"
	req.Header.Set("Authorization", secret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var traffic map[string]struct {
		TX int64 `json:"tx"`
		RX int64 `json:"rx"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&traffic); err != nil {
		return
	}

	for name, usage := range traffic {
		delta := usage.TX + usage.RX
		if delta <= 0 {
			continue
		}
		if err := m.db.Exec(
			"UPDATE users SET traffic_used_bytes = traffic_used_bytes + ? WHERE name = ?",
			delta, name,
		).Error; err != nil {
			log.Printf("stats: обновление трафика %q: %v", name, err)
		}
	}
}

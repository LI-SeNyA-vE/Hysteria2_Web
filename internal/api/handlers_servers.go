package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"hysteria2-web/internal/models"
)

var processStart = time.Now()

func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	var servers []models.Server
	if err := s.db.Order("id").Find(&servers).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка БД")
		return
	}
	dtos := make([]serverDTO, len(servers))
	for i, srv := range servers {
		dtos[i] = toServerDTO(srv)
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (s *Server) handleGetServerLogs(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	var srv models.Server
	if err := s.db.First(&srv, id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "сервер не найден")
		return
	}

	var lines []string
	if srv.Role == models.RoleMain || srv.Role == models.RoleMainNode1 {
		if s.manager != nil {
			lines = s.manager.LogBuf().Lines()
		}
	} else {
		if s.registry != nil {
			lines = s.registry.GetNodeLogs(srv.Name)
		}
	}
	if lines == nil {
		lines = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"lines": lines})
}

func (s *Server) handleGetNodeConfig(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	var srv models.Server
	if err := s.db.First(&srv, id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "сервер не найден")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"bandwidthUp":   srv.BandwidthUp,
		"bandwidthDown": srv.BandwidthDown,
		"masqueradeUrl": srv.MasqueradeURL,
	})
}

func (s *Server) handleUpdateNodeConfig(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	var req struct {
		BandwidthUp   string `json:"bandwidthUp"`
		BandwidthDown string `json:"bandwidthDown"`
		MasqueradeURL string `json:"masqueradeUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if err := s.db.Model(&models.Server{}).Where("id = ?", id).Updates(map[string]any{
		"bandwidth_up":   req.BandwidthUp,
		"bandwidth_down": req.BandwidthDown,
		"masquerade_url": req.MasqueradeURL,
	}).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	var req struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if err := s.db.Model(&models.Server{}).Where("id = ?", id).Update("display_name", req.DisplayName).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	if err := s.db.Delete(&models.Server{}, id).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	var total, active int64
	s.db.Model(&models.User{}).Count(&total)
	s.db.Model(&models.User{}).Where("is_active = ?", true).Count(&active)

	var trafficSum int64
	s.db.Model(&models.User{}).Select("COALESCE(SUM(traffic_used_bytes), 0)").Scan(&trafficSum)

	var hySt hysteriaStatusDTO
	if s.manager != nil {
		st := s.manager.Status()
		hySt = hysteriaStatusDTO{
			Installed: st.Installed,
			Running:   st.Running,
			Version:   st.Version,
			Port:      st.Port,
		}
	}

	writeJSON(w, http.StatusOK, dashboardStatsDTO{
		TotalUsers:     int(total),
		ActiveUsers:    int(active),
		TotalTrafficGb: float64(trafficSum) / GiB,
		Uptime:         formatUptime(time.Since(processStart)),
		Hysteria:       hySt,
	})
}

func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dд %dч %dм", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dч %dм", hours, minutes)
	}
	return fmt.Sprintf("%dм", minutes)
}

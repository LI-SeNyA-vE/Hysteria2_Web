package api

import (
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

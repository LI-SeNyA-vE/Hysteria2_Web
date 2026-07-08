package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"

	"hysteria2-web/internal/models"
	subutil "hysteria2-web/internal/sub"
)

func (s *Server) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	var subs []models.Subscription
	if err := s.db.Order("id").Find(&subs).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка БД")
		return
	}

	var users []models.User
	s.db.Select("id, name").Find(&users)
	nameByID := make(map[uint]string, len(users))
	for _, u := range users {
		nameByID[u.ID] = u.Name
	}

	dtos := make([]subscriptionDTO, len(subs))
	for i, sub := range subs {
		dtos[i] = toSubscriptionDTO(sub, nameByID[sub.UserID])
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (s *Server) handleCreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if req.UserID == 0 {
		writeErr(w, http.StatusBadRequest, "поле userId обязательно")
		return
	}

	var user models.User
	if err := s.db.First(&user, req.UserID).Error; err != nil {
		writeErr(w, http.StatusNotFound, "пользователь не найден")
		return
	}

	name := req.Name
	if name == "" {
		name = user.Name
	}

	sub := models.Subscription{
		UserID: req.UserID,
		Token:  generateSubToken(),
		Name:   name,
	}
	if err := s.db.Create(&sub).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка создания подписки")
		return
	}
	writeJSON(w, http.StatusCreated, toSubscriptionDTO(sub, user.Name))
}

func (s *Server) handleDeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	if err := s.db.Delete(&models.Subscription{}, id).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleGetSub — GET /sub/{token} (публичный, без авторизации).
// Возвращает base64 из hysteria2:// URI для VPN-клиентов.
func (s *Server) handleGetSub(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var sub models.Subscription
	if err := s.db.Where("token = ?", token).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
		} else {
			writeErr(w, http.StatusInternalServerError, "ошибка БД")
		}
		return
	}

	var user models.User
	if err := s.db.First(&user, sub.UserID).Error; err != nil || !user.IsActive {
		http.NotFound(w, r)
		return
	}

	if user.ServerID == 0 {
		writeErr(w, http.StatusUnprocessableEntity, "пользователь не привязан к серверу")
		return
	}

	var server models.Server
	if err := s.db.First(&server, user.ServerID).Error; err != nil {
		writeErr(w, http.StatusUnprocessableEntity, "сервер пользователя не найден")
		return
	}

	obfsPassword, _ := s.db.GetSetting(models.SettingObfsPassword)

	uri := subutil.BuildURI(subutil.URIConfig{
		UserName:     user.Name,
		UserPassword: user.Password,
		PublicIP:     server.PublicIP,
		Hy2Port:      server.Hy2Port,
		ObfsPassword: obfsPassword,
		CertSHA256:   server.CertSHA256,
		Label:        sub.Name,
	})

	now := time.Now()
	s.db.Model(&sub).Update("last_accessed_at", &now)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(subutil.EncodeBase64([]string{uri})))
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"hysteria2-web/internal/models"
)

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := s.db.Order("id").Find(&users).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка БД")
		return
	}
	dtos := make([]userDTO, len(users))
	for i, u := range users {
		dtos[i] = toUserDTO(u)
	}
	writeJSON(w, http.StatusOK, dtos)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if req.Name == "" {
		writeErr(w, http.StatusBadRequest, "поле name обязательно")
		return
	}

	user := models.User{
		Name:              req.Name,
		UUID:              generateUUID(),
		Password:          generateUserPassword(12),
		TrafficLimitBytes: int64(req.TrafficLimitGb * GiB),
		IsActive:          true,
		ServerID:          req.ServerID,
	}
	if req.ExpireDays != nil && *req.ExpireDays > 0 {
		t := time.Now().Add(time.Duration(*req.ExpireDays) * 24 * time.Hour)
		user.ExpireAt = &t
	}

	if err := s.db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeErr(w, http.StatusConflict, "пользователь с таким именем уже существует")
			return
		}
		writeErr(w, http.StatusInternalServerError, "ошибка создания пользователя")
		return
	}
	writeJSON(w, http.StatusCreated, toUserDTO(user))
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}

	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		writeErr(w, http.StatusNotFound, "пользователь не найден")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.TrafficLimitGb != nil {
		user.TrafficLimitBytes = int64(*req.TrafficLimitGb * GiB)
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.ExpireAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpireAt)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "неверный формат expireAt (RFC3339)")
			return
		}
		user.ExpireAt = &t
	}

	if err := s.db.Save(&user).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка сохранения")
		return
	}
	writeJSON(w, http.StatusOK, toUserDTO(user))
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "неверный id")
		return
	}
	if err := s.db.Delete(&models.User{}, id).Error; err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка удаления")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

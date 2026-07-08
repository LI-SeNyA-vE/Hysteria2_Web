package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if !s.auth.CheckPassword(req.Password) {
		writeErr(w, http.StatusUnauthorized, "неверный пароль")
		return
	}
	token, err := s.auth.IssueToken()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "ошибка генерации токена")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token})
}

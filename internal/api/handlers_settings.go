package api

import "net/http"

func (s *Server) handleGetSettings(w http.ResponseWriter, _ *http.Request) {
	token, err := s.auth.NodeToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "ошибка получения токена"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"nodeToken": token})
}

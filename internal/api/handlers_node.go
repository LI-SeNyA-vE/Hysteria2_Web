package api

import (
	"encoding/json"
	"net/http"

	"hysteria2-web/internal/cluster"
)

func (s *Server) handleNodeRegister(w http.ResponseWriter, r *http.Request) {
	if s.registry == nil {
		writeErr(w, http.StatusNotImplemented, "registry недоступен для этой роли")
		return
	}
	var req cluster.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if req.Name == "" || req.Role == "" {
		writeErr(w, http.StatusBadRequest, "name и role обязательны")
		return
	}
	desired, err := s.registry.Register(req)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, desired)
}

func (s *Server) handleNodeHeartbeat(w http.ResponseWriter, r *http.Request) {
	if s.registry == nil {
		writeErr(w, http.StatusNotImplemented, "registry недоступен для этой роли")
		return
	}
	var req cluster.HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	desired, err := s.registry.Heartbeat(req)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, desired)
}

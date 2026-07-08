package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) hysteriaGuard(w http.ResponseWriter) bool {
	if s.manager == nil {
		writeErr(w, http.StatusNotImplemented, "hysteria manager недоступен для этой роли")
		return false
	}
	return true
}

func (s *Server) handleHysteriaStatus(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	st := s.manager.Status()
	writeJSON(w, http.StatusOK, hysteriaStatusDTO{
		Installed: st.Installed,
		Running:   st.Running,
		Version:   st.Version,
		Port:      st.Port,
	})
}

func (s *Server) handleHysteriaGetConfig(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	c := s.manager.GetConfig()
	writeJSON(w, http.StatusOK, hysteriaConfigDTO{
		Port:          c.Port,
		ObfsPassword:  c.ObfsPassword,
		MasqueradeURL: c.MasqueradeURL,
		CertSHA256:    c.CertSHA256,
	})
}

func (s *Server) handleHysteriaUpdateConfig(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	var req struct {
		ObfsPassword  string `json:"obfsPassword"`
		MasqueradeURL string `json:"masqueradeUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "неверный JSON")
		return
	}
	if err := s.manager.UpdateConfig(req.ObfsPassword, req.MasqueradeURL); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHysteriaInstall(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	if err := s.manager.Install(r.Context()); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHysteriaStart(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	if err := s.manager.Start(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHysteriaStop(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	if err := s.manager.Stop(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHysteriaReload(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	if err := s.manager.ReloadConfig(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHysteriaCertRegen(w http.ResponseWriter, r *http.Request) {
	if !s.hysteriaGuard(w) {
		return
	}
	pin, err := s.manager.RegenerateCert()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"sha256": pin})
}

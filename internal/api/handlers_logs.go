package api

import "net/http"

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")

	var lines []string
	switch source {
	case "hysteria":
		if s.manager != nil {
			lines = s.manager.LogBuf().Lines()
		}
	default: // "panel"
		if s.panelLog != nil {
			lines = s.panelLog.Lines()
		}
	}
	if lines == nil {
		lines = []string{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"lines": lines})
}

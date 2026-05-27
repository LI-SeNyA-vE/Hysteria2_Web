package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"hysteria2-web/internal/service"

	"github.com/go-chi/chi/v5"
)

type SubHandler struct {
	sub    *service.SubscriptionService
	logger *slog.Logger
}

func NewSubHandler(sub *service.SubscriptionService, logger *slog.Logger) *SubHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &SubHandler{sub: sub, logger: logger}
}

func (h *SubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, subscriptionToken(r))
}

func (h *SubHandler) ServeQuery(w http.ResponseWriter, r *http.Request) {
	h.serve(w, r, strings.TrimSpace(r.URL.Query().Get("token")))
}

func subscriptionToken(r *http.Request) string {
	token := chi.URLParam(r, "token")
	if decoded, err := url.PathUnescape(token); err == nil {
		token = decoded
	}
	return strings.TrimSpace(token)
}

func (h *SubHandler) serve(w http.ResponseWriter, r *http.Request, token string) {
	body, err := h.sub.BuildSubscription(r.Context(), token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSubscriptionNotFound):
			h.logger.Warn("subscription not found",
				"sub_token", token,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			http.Error(w, "404: token not found.\n\n"+
				"URL must be /sub/{SubToken}, not /sub/{username}.\n"+
				"Get the correct URL from panel: item 10 (QR subscription).\n", http.StatusNotFound)
		case errors.Is(err, service.ErrSubscriptionForbidden):
			http.Error(w, "Forbidden", http.StatusForbidden)
		case errors.Is(err, service.ErrSubscriptionNoURIs):
			h.logger.Warn("subscription has no URIs", "sub_token", token)
			http.Error(w, "No connection URIs available", http.StatusServiceUnavailable)
		default:
			h.logger.Error("subscription handler failed", "sub_token", token, "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Profile-Update-Interval", "24")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

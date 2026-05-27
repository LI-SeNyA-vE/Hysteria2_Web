package httpapi

import (
	"log/slog"
	"net/http"

	"hysteria2-web/internal/app"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(a *app.App, logger *slog.Logger) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	sub := NewSubHandler(a.SubSvc, logger)
	r.Get("/sub/{token}", sub.ServeHTTP)
	r.Get("/sub", sub.ServeQuery)
	return r
}

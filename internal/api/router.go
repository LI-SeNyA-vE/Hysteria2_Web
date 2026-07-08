package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"hysteria2-web/internal/auth"
	"hysteria2-web/internal/cluster"
	"hysteria2-web/internal/db"
	"hysteria2-web/internal/hysteria"
)

// Server — HTTP-сервер API со всеми зависимостями.
type Server struct {
	auth     *auth.Auth
	db       *db.DB
	dev      bool
	manager  *hysteria.Manager // nil для ролей без hysteria (pure main)
	registry *cluster.Registry // nil для ролей node (без БД)
}

func NewServer(a *auth.Auth, d *db.DB, dev bool, mgr *hysteria.Manager, reg *cluster.Registry) *Server {
	return &Server{auth: a, db: d, dev: dev, manager: mgr, registry: reg}
}

// Handler строит chi-роутер для всей панели.
func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if s.dev {
		r.Use(corsDevMiddleware)
	}

	// Публичные маршруты.
	r.Post("/api/login", s.handleLogin)

	// Защищённые маршруты (JWT).
	r.Group(func(r chi.Router) {
		r.Use(jwtMiddleware(s.auth))

		r.Get("/api/users", s.handleListUsers)
		r.Post("/api/users", s.handleCreateUser)
		r.Put("/api/users/{id}", s.handleUpdateUser)
		r.Delete("/api/users/{id}", s.handleDeleteUser)

		r.Get("/api/servers", s.handleListServers)
		r.Delete("/api/servers/{id}", s.handleDeleteServer)
		r.Get("/api/stats", s.handleGetStats)

		r.Get("/api/subscriptions", s.handleListSubscriptions)
		r.Post("/api/subscriptions", s.handleCreateSubscription)
		r.Delete("/api/subscriptions/{id}", s.handleDeleteSubscription)

		r.Get("/api/hysteria/status", s.handleHysteriaStatus)
		r.Get("/api/hysteria/config", s.handleHysteriaGetConfig)
		r.Put("/api/hysteria/config", s.handleHysteriaUpdateConfig)
		r.Post("/api/hysteria/install", s.handleHysteriaInstall)
		r.Post("/api/hysteria/start", s.handleHysteriaStart)
		r.Post("/api/hysteria/stop", s.handleHysteriaStop)
		r.Post("/api/hysteria/reload-config", s.handleHysteriaReload)
		r.Post("/api/hysteria/cert/regenerate", s.handleHysteriaCertRegen)

		r.Get("/api/settings", s.handleGetSettings)

		r.Get("/api/update/check", s.handleCheckUpdate)
		r.Post("/api/update/apply", s.handleApplyUpdate)
	})

	// Нодовые маршруты (X-Node-Token, без JWT).
	r.Group(func(r chi.Router) {
		r.Use(nodeTokenMiddleware(s.auth))
		r.Post("/api/node/register", s.handleNodeRegister)
		r.Post("/api/node/heartbeat", s.handleNodeHeartbeat)
	})

	// GET /sub/{token} — публичный, без auth.
	r.Get("/sub/{token}", s.handleGetSub)

	// SPA — всё остальное.
	r.NotFound(spaHandler().ServeHTTP)

	return r
}

func (s *Server) stub(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"message":"не реализовано"}`))
}

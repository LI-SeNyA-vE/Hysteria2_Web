package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"hysteria2-web/internal/auth"
)

type contextKey string

const contextKeyAuth contextKey = "auth"

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

// jwtMiddleware проверяет Bearer-токен в заголовке Authorization.
func jwtMiddleware(a *auth.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeErr(w, http.StatusUnauthorized, "требуется авторизация")
				return
			}
			if err := a.Verify(token); err != nil {
				writeErr(w, http.StatusUnauthorized, "невалидный токен")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// nodeTokenMiddleware проверяет заголовок X-Node-Token.
func nodeTokenMiddleware(a *auth.Auth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("X-Node-Token")
			if token == "" {
				writeErr(w, http.StatusUnauthorized, "отсутствует X-Node-Token")
				return
			}
			expected, err := a.NodeToken()
			if err != nil || token != expected {
				writeErr(w, http.StatusForbidden, "неверный токен ноды")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// corsDevMiddleware разрешает все origins только в dev-режиме (Vite proxy).
func corsDevMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Node-Token")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

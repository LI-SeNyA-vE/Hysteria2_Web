package service_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/domain/user"
	"hysteria2-web/internal/httpapi"
	"hysteria2-web/internal/repository"
	"hysteria2-web/internal/service"

	"github.com/go-chi/chi/v5"
)

func TestBuildSubscriptionSuccess(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/server/status":
			_ = json.NewEncoder(w).Encode(blitz.ServerStatusResponse{OnlineUsers: 0})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/users/alice/uri":
			_ = json.NewEncoder(w).Encode(blitz.UserURIResponse{
				Username: "alice",
				Nodes: []blitz.NodeURI{
					{Name: "main", URI: "hy2://alice@server1:443/?insecure=1"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	})

	db := setupTestDB(t)
	serverSvc, _, serverID := setupTestServer(t, db, handler)
	userRepo := repository.NewUserRepository(db)
	subSvc := service.NewSubscriptionService(userRepo, serverSvc, nil)

	token := "test-sub-token"
	expires := time.Now().UTC().Add(24 * time.Hour)
	if err := userRepo.Create(&user.User{
		ServerID:     serverID,
		Username:     "alice",
		AuthPassword: "pass",
		SubToken:     token,
		TrafficLimit: 10,
		IsActive:     true,
		ExpiresAt:    expires,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	body, err := subSvc.BuildSubscription(context.Background(), token)
	if err != nil {
		t.Fatalf("BuildSubscription() error = %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}
	if !strings.Contains(string(decoded), "hy2://alice@server1:443/?insecure=1#test") {
		t.Fatalf("unexpected subscription body: %s", decoded)
	}
}

func TestBuildSubscriptionForbiddenInactive(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	subSvc := service.NewSubscriptionService(userRepo, service.NewServerService(repository.NewServerRepository(db), blitz.NewRegistry()), nil)

	token := "inactive-token"
	if err := userRepo.Create(&user.User{
		ServerID:     1,
		Username:     "bob",
		AuthPassword: "pass",
		SubToken:     token,
		TrafficLimit: 10,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := userRepo.Deactivate(1, "bob"); err != nil {
		t.Fatalf("deactivate user: %v", err)
	}

	_, err := subSvc.BuildSubscription(context.Background(), token)
	if err != service.ErrSubscriptionForbidden {
		t.Fatalf("BuildSubscription() error = %v, want ErrSubscriptionForbidden", err)
	}
}

func TestSubHTTPEndpoint(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/server/status":
			_ = json.NewEncoder(w).Encode(blitz.ServerStatusResponse{OnlineUsers: 0})
		case r.URL.Path == "/api/v1/users/alice/uri":
			_ = json.NewEncoder(w).Encode(blitz.UserURIResponse{
				Username: "alice",
				Nodes:    []blitz.NodeURI{{Name: "main", URI: "hy2://alice@host:443/"}},
			})
		default:
			http.NotFound(w, r)
		}
	})

	db := setupTestDB(t)
	serverSvc, _, serverID := setupTestServer(t, db, handler)
	userRepo := repository.NewUserRepository(db)
	subSvc := service.NewSubscriptionService(userRepo, serverSvc, nil)

	token := "http-token"
	if err := userRepo.Create(&user.User{
		ServerID:     serverID,
		Username:     "alice",
		AuthPassword: "pass",
		SubToken:     token,
		TrafficLimit: 0,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/sub/{token}", httpapi.NewSubHandler(subSvc, nil).ServeHTTP)

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/sub/" + token)
	if err != nil {
		t.Fatalf("GET /sub/{token} error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(string(body)); err != nil {
		t.Fatalf("response is not valid base64: %v", err)
	}
}

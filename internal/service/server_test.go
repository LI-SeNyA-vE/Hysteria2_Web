package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/repository"
	"hysteria2-web/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestServerServiceCreateDuplicate(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/server/status" {
			_ = json.NewEncoder(w).Encode(blitz.ServerStatusResponse{OnlineUsers: 0})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&server.Server{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	serverRepo := repository.NewServerRepository(db)
	registry := blitz.NewRegistry()
	serverSvc := service.NewServerService(serverRepo, registry)

	if _, err := serverSvc.CreateServer(context.Background(), "node1", srv.URL, "key"); err != nil {
		t.Fatalf("first CreateServer() error = %v", err)
	}
	_, err = serverSvc.CreateServer(context.Background(), "node1", srv.URL, "key")
	if err != service.ErrServerExists {
		t.Fatalf("error = %v, want ErrServerExists", err)
	}
}

func TestServerServiceGetClient(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/server/status" {
			_ = json.NewEncoder(w).Encode(blitz.ServerStatusResponse{OnlineUsers: 5})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&server.Server{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	serverRepo := repository.NewServerRepository(db)
	registry := blitz.NewRegistry()
	serverSvc := service.NewServerService(serverRepo, registry)

	created, err := serverSvc.CreateServer(context.Background(), "node1", srv.URL, "key")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}

	client, err := serverSvc.GetClient(created.ID)
	if err != nil {
		t.Fatalf("GetClient() error = %v", err)
	}

	status, err := client.GetServerStatus(context.Background())
	if err != nil {
		t.Fatalf("GetServerStatus() error = %v", err)
	}
	if status.OnlineUsers != 5 {
		t.Fatalf("online users = %d, want 5", status.OnlineUsers)
	}
}

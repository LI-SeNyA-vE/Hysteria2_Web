package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/domain/user"
	"hysteria2-web/internal/repository"
	"hysteria2-web/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&server.Server{}, &user.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func setupTestServer(t *testing.T, db *gorm.DB, handler http.Handler) (*service.ServerService, *service.BlitzService, uint) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	serverRepo := repository.NewServerRepository(db)
	userRepo := repository.NewUserRepository(db)
	registry := blitz.NewRegistry()
	serverSvc := service.NewServerService(serverRepo, registry)
	blitzSvc := service.NewBlitzService(registry, userRepo, nil)

	created, err := serverSvc.CreateServer(context.Background(), "test", srv.URL, "key")
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	return serverSvc, blitzSvc, created.ID
}

func TestAddUser(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	addCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/server/version":
			_ = json.NewEncoder(w).Encode(blitz.VersionInfoResponse{CurrentVersion: "0.2.0"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/users/":
			mu.Lock()
			addCalled = true
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "ok"})
		default:
			http.NotFound(w, r)
		}
	})

	db := setupTestDB(t)
	_, svc, serverID := setupTestServer(t, db, handler)

	if err := svc.AddUser(context.Background(), serverID, "alice", "pass", 10, 30); err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	mu.Lock()
	if !addCalled {
		mu.Unlock()
		t.Fatal("blitz AddUser was not called")
	}
	mu.Unlock()

	repo := repository.NewUserRepository(db)
	u, err := repo.GetByUsername(serverID, "alice")
	if err != nil {
		t.Fatalf("GetByUsername() error = %v", err)
	}
	if u == nil || !u.IsActive || u.TrafficLimit != 10 {
		t.Fatalf("unexpected user: %+v", u)
	}
}

func TestAddUserAlreadyExists(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	registry := blitz.NewRegistry()
	svc := service.NewBlitzService(registry, userRepo, nil)

	if err := userRepo.Create(&user.User{
		ServerID:     1,
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 10,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	err := svc.AddUser(context.Background(), 1, "alice", "pass", 10, 30)
	if err == nil {
		t.Fatal("expected ErrUserExists")
	}
	if err != service.ErrUserExists {
		t.Fatalf("error = %v, want ErrUserExists", err)
	}
}

func TestKickUser(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	deleteCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/server/version":
			_ = json.NewEncoder(w).Encode(blitz.VersionInfoResponse{CurrentVersion: "0.2.0"})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/users/alice":
			mu.Lock()
			deleteCalled = true
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "ok"})
		default:
			http.NotFound(w, r)
		}
	})

	db := setupTestDB(t)
	_, svc, serverID := setupTestServer(t, db, handler)

	repo := repository.NewUserRepository(db)
	if err := repo.Create(&user.User{
		ServerID:     serverID,
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 10,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := svc.KickUser(context.Background(), serverID, "alice"); err != nil {
		t.Fatalf("KickUser() error = %v", err)
	}

	mu.Lock()
	if !deleteCalled {
		mu.Unlock()
		t.Fatal("blitz RemoveUser was not called")
	}
	mu.Unlock()

	u, _ := repo.GetByUsername(serverID, "alice")
	if u == nil || u.IsActive {
		t.Fatalf("expected deactivated user, got %+v", u)
	}
}

func TestSyncTrafficDeltaAndKick(t *testing.T) {
	t.Parallel()

	const bytesPerGB = 1024 * 1024 * 1024
	upload := int64(bytesPerGB)
	download := int64(0)

	var mu sync.Mutex
	deleteCalled := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/server/version":
			_ = json.NewEncoder(w).Encode(blitz.VersionInfoResponse{CurrentVersion: "0.2.0"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/users/":
			users := []blitz.UserInfo{
				{
					Username:      "alice",
					UploadBytes:   &upload,
					DownloadBytes: &download,
				},
			}
			_ = json.NewEncoder(w).Encode(users)
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/v1/users/"):
			mu.Lock()
			deleteCalled = true
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "ok"})
		default:
			http.NotFound(w, r)
		}
	})

	db := setupTestDB(t)
	_, svc, serverID := setupTestServer(t, db, handler)

	repo := repository.NewUserRepository(db)
	if err := repo.Create(&user.User{
		ServerID:     serverID,
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 1,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := svc.SyncTraffic(context.Background()); err != nil {
		t.Fatalf("SyncTraffic() error = %v", err)
	}

	u, _ := repo.GetByUsername(serverID, "alice")
	if u == nil {
		t.Fatal("user not found")
	}
	if u.TrafficUsed < 1 {
		t.Fatalf("traffic_used = %d, want >= 1", u.TrafficUsed)
	}
	if u.IsActive {
		t.Fatal("expected user to be kicked after limit exceeded")
	}

	mu.Lock()
	if !deleteCalled {
		mu.Unlock()
		t.Fatal("expected kick (delete) to be called")
	}
	mu.Unlock()
}

func TestSyncTrafficDoesNotDoubleCount(t *testing.T) {
	t.Parallel()

	upload := int64(512 * 1024 * 1024)
	download := int64(512 * 1024 * 1024)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/server/version":
			_ = json.NewEncoder(w).Encode(blitz.VersionInfoResponse{CurrentVersion: "0.2.0"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/users/":
			users := []blitz.UserInfo{
				{
					Username:      "alice",
					UploadBytes:   &upload,
					DownloadBytes: &download,
				},
			}
			_ = json.NewEncoder(w).Encode(users)
		default:
			http.NotFound(w, r)
		}
	})

	db := setupTestDB(t)
	_, svc, serverID := setupTestServer(t, db, handler)

	repo := repository.NewUserRepository(db)
	if err := repo.Create(&user.User{
		ServerID:     serverID,
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 100,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := svc.SyncTraffic(context.Background()); err != nil {
		t.Fatalf("first SyncTraffic() error = %v", err)
	}
	u1, _ := repo.GetByUsername(serverID, "alice")
	if u1 == nil {
		t.Fatal("user not found")
	}

	if err := svc.SyncTraffic(context.Background()); err != nil {
		t.Fatalf("second SyncTraffic() error = %v", err)
	}
	u2, _ := repo.GetByUsername(serverID, "alice")
	if u2.TrafficUsed != u1.TrafficUsed {
		t.Fatalf("traffic_used changed on second sync: %d -> %d", u1.TrafficUsed, u2.TrafficUsed)
	}
}

func TestCreateAndDeleteServer(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/server/version" {
			_ = json.NewEncoder(w).Encode(blitz.VersionInfoResponse{CurrentVersion: "0.2.0"})
			return
		}
		http.NotFound(w, r)
	})

	db := setupTestDB(t)
	serverSvc, _, _ := setupTestServer(t, db, handler)

	servers, err := serverSvc.ListServers(context.Background())
	if err != nil {
		t.Fatalf("ListServers() error = %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("servers count = %d, want 1", len(servers))
	}

	if err := serverSvc.DeleteServer(context.Background(), servers[0].ID); err != nil {
		t.Fatalf("DeleteServer() error = %v", err)
	}

	servers, err = serverSvc.ListServers(context.Background())
	if err != nil {
		t.Fatalf("ListServers() error = %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("servers count = %d, want 0", len(servers))
	}
}

type mockBlitzClient struct{}

func (m *mockBlitzClient) AddUser(_ context.Context, _ blitz.AddUserRequest) error {
	return nil
}

func (m *mockBlitzClient) RemoveUser(_ context.Context, _ string) error {
	return nil
}

func (m *mockBlitzClient) ListUsers(_ context.Context) ([]blitz.UserInfo, error) {
	return nil, nil
}

var _ blitz.UserClient = (*mockBlitzClient)(nil)

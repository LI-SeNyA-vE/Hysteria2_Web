package service_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"hysteria2-web/internal/blitz"
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
	if err := db.AutoMigrate(&user.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAddUser(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	addCalled := false

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/users/" {
			mu.Lock()
			addCalled = true
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	client := blitz.NewClient(srv.URL, "key")
	svc := service.NewBlitzService(client, repo, nil)

	if err := svc.AddUser(context.Background(), "alice", "pass", 10, 30); err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if !addCalled {
		t.Fatal("blitz AddUser was not called")
	}

	u, err := repo.GetByUsername("alice")
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
	repo := repository.NewUserRepository(db)
	svc := service.NewBlitzService(&mockBlitzClient{}, repo, nil)

	if err := repo.Create(&user.User{
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 10,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	err := svc.AddUser(context.Background(), "alice", "pass", 10, 30)
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/v1/users/alice" {
			mu.Lock()
			deleteCalled = true
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	if err := repo.Create(&user.User{
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 10,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	client := blitz.NewClient(srv.URL, "key")
	svc := service.NewBlitzService(client, repo, nil)

	if err := svc.KickUser(context.Background(), "alice"); err != nil {
		t.Fatalf("KickUser() error = %v", err)
	}

	mu.Lock()
	if !deleteCalled {
		mu.Unlock()
		t.Fatal("blitz RemoveUser was not called")
	}
	mu.Unlock()

	u, _ := repo.GetByUsername("alice")
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
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
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	if err := repo.Create(&user.User{
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 1,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	client := blitz.NewClient(srv.URL, "key")
	svc := service.NewBlitzService(client, repo, nil)

	if err := svc.SyncTraffic(context.Background()); err != nil {
		t.Fatalf("SyncTraffic() error = %v", err)
	}

	u, _ := repo.GetByUsername("alice")
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

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/users/" {
			http.NotFound(w, r)
			return
		}
		users := []blitz.UserInfo{
			{
				Username:      "alice",
				UploadBytes:   &upload,
				DownloadBytes: &download,
			},
		}
		_ = json.NewEncoder(w).Encode(users)
	}))
	defer srv.Close()

	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	if err := repo.Create(&user.User{
		Username:     "alice",
		AuthPassword: "pass",
		TrafficLimit: 100,
		IsActive:     true,
	}); err != nil {
		t.Fatalf("create user: %v", err)
	}

	client := blitz.NewClient(srv.URL, "key")
	svc := service.NewBlitzService(client, repo, nil)

	if err := svc.SyncTraffic(context.Background()); err != nil {
		t.Fatalf("first SyncTraffic() error = %v", err)
	}
	u1, _ := repo.GetByUsername("alice")
	if u1 == nil {
		t.Fatal("user not found")
	}

	if err := svc.SyncTraffic(context.Background()); err != nil {
		t.Fatalf("second SyncTraffic() error = %v", err)
	}
	u2, _ := repo.GetByUsername("alice")
	if u2.TrafficUsed != u1.TrafficUsed {
		t.Fatalf("traffic_used changed on second sync: %d -> %d", u1.TrafficUsed, u2.TrafficUsed)
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

var _ blitz.Client = (*mockBlitzClient)(nil)

func TestBlitzClientAddUserError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		body, _ := io.ReadAll(r.Body)
		_ = body
		_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "invalid user"})
	}))
	defer srv.Close()

	client := blitz.NewClient(srv.URL, "key")
	password := "pass"
	err := client.AddUser(context.Background(), blitz.AddUserRequest{
		Username:       "alice",
		Password:       &password,
		TrafficLimit:   1,
		ExpirationDays: 1,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid user") {
		t.Fatalf("error = %v", err)
	}
}

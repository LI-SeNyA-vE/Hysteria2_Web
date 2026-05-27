package blitz_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hysteria2-web/internal/blitz"
)

func TestAddUser(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/users/" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "test-key" {
			t.Fatalf("authorization header = %q, want test-key", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var req blitz.AddUserRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("unmarshal body: %v", err)
		}
		if req.Username != "alice" || req.TrafficLimit != 10 || req.ExpirationDays != 30 {
			t.Fatalf("unexpected body: %+v", req)
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "User added"})
	}))
	defer srv.Close()

	client := blitz.NewClient(srv.URL, "test-key")
	password := "secret"
	err := client.AddUser(context.Background(), blitz.AddUserRequest{
		Username:       "alice",
		Password:       &password,
		TrafficLimit:   10,
		ExpirationDays: 30,
	})
	if err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}
}

func TestRemoveUser(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/v1/users/bob" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "User removed successfully"})
	}))
	defer srv.Close()

	client := blitz.NewClient(srv.URL, "test-key")
	if err := client.RemoveUser(context.Background(), "bob"); err != nil {
		t.Fatalf("RemoveUser() error = %v", err)
	}
}

func TestRemoveUserNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(blitz.DetailResponse{Detail: "User not found"})
	}))
	defer srv.Close()

	client := blitz.NewClient(srv.URL, "test-key")
	err := client.RemoveUser(context.Background(), "missing")
	if !errors.Is(err, blitz.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	upload := int64(100)
	download := int64(200)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/users/" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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

	client := blitz.NewClient(srv.URL, "test-key")
	users, err := client.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 1 || users[0].Username != "alice" {
		t.Fatalf("unexpected users: %+v", users)
	}
}

func TestGetServerStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/server/status" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(blitz.ServerStatusResponse{
			Uptime:      "1h",
			OnlineUsers: 3,
		})
	}))
	defer srv.Close()

	client := blitz.NewClient(srv.URL, "test-key")
	status, err := client.GetServerStatus(context.Background())
	if err != nil {
		t.Fatalf("GetServerStatus() error = %v", err)
	}
	if status.OnlineUsers != 3 {
		t.Fatalf("online users = %d, want 3", status.OnlineUsers)
	}
}

func TestAddUserError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
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

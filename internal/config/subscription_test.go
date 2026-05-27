package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSubscriptionURL(t *testing.T) {
	Set(Config{SubDomain: "https://panel.example.com", SubPath: "sub"})
	if got := SubscriptionURL("abc"); got != "https://panel.example.com/sub/abc" {
		t.Fatalf("SubscriptionURL() = %q", got)
	}
}

func TestSubscriptionURLCustomPath(t *testing.T) {
	Set(Config{SubDomain: "https://panel.example.com", SubPath: "subtoken"})
	if got := SubscriptionURL("abc"); got != "https://panel.example.com/subtoken/abc" {
		t.Fatalf("SubscriptionURL() = %q", got)
	}
}

func TestSubscriptionPublicBaseFromHTTPAddr(t *testing.T) {
	Set(Config{HTTPAddr: "0.0.0.0:8787"})
	if got := SubscriptionPublicBase(); got != "http://127.0.0.1:8787" {
		t.Fatalf("SubscriptionPublicBase() = %q", got)
	}
}

func TestUsingLocalSubscriptionURL(t *testing.T) {
	Set(Config{})
	if !UsingLocalSubscriptionURL() {
		t.Fatal("expected local URL mode")
	}
	Set(Config{SubDomain: "http://10.0.0.1:8080"})
	if UsingLocalSubscriptionURL() {
		t.Fatal("expected public URL mode")
	}
}

func TestLoadCreatesDefaultFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "panel.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPath != "./panel.db" {
		t.Fatalf("DBPath = %q", cfg.DBPath)
	}
	if cfg.SubPath != "sub" {
		t.Fatalf("SubPath = %q", cfg.SubPath)
	}
	if cfg.SyncInterval.String() != "30s" {
		t.Fatalf("SyncInterval = %s", cfg.SyncInterval)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "panel.json")
	content := `{
  "db_path": "./custom.db",
  "log_path": "./custom.log",
  "http_addr": "0.0.0.0:9999",
  "sync_interval": "1m",
  "sub_domain": "https://vpn.example.com",
  "sub_path": "subtoken"
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DBPath != "./custom.db" {
		t.Fatalf("DBPath = %q", cfg.DBPath)
	}
	if cfg.HTTPAddr != "0.0.0.0:9999" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.SyncInterval != time.Minute {
		t.Fatalf("SyncInterval = %s", cfg.SyncInterval)
	}
	if got := cfg.SubscriptionURL("tok"); got != "https://vpn.example.com/subtoken/tok" {
		t.Fatalf("SubscriptionURL() = %q", got)
	}
}

func TestLoadLegacySubPublicURLAsPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "panel.json")
	content := `{"sub_public_url":"subtoken"}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SubPath != "subtoken" {
		t.Fatalf("SubPath = %q", cfg.SubPath)
	}
}

func TestLoadLegacySubPublicURLAsDomain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "panel.json")
	content := `{"sub_public_url":"https://vpn.example.com"}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SubDomain != "https://vpn.example.com" {
		t.Fatalf("SubDomain = %q", cfg.SubDomain)
	}
}

func TestLoadInvalidSyncInterval(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "panel.json")
	if err := os.WriteFile(path, []byte(`{"sync_interval":"not-a-duration"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid sync_interval")
	}
}

func TestLoadInvalidSubPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "panel.json")
	if err := os.WriteFile(path, []byte(`{"sub_path":"sub/extra"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for invalid sub_path")
	}
}

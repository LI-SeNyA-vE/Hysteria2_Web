package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultPath = "panel.json"

const defaultSubPath = "sub"

type Config struct {
	DBPath       string        `json:"db_path"`
	LogPath      string        `json:"log_path"`
	HTTPAddr     string        `json:"http_addr"`
	SyncInterval time.Duration `json:"-"`
	SubDomain    string        `json:"sub_domain"`
	SubPath      string        `json:"sub_path"`
}

type fileConfig struct {
	DBPath       string `json:"db_path"`
	LogPath      string `json:"log_path"`
	HTTPAddr     string `json:"http_addr"`
	SyncInterval string `json:"sync_interval"`
	SubDomain    string `json:"sub_domain"`
	SubPath      string `json:"sub_path"`
}

var global Config
var configPath = DefaultPath

func Default() Config {
	return Config{
		DBPath:       "./panel.db",
		LogPath:      "./panel.log",
		HTTPAddr:     "0.0.0.0:8787",
		SyncInterval: 30 * time.Second,
		SubPath:      defaultSubPath,
	}
}

func Get() Config {
	if global == (Config{}) {
		return Default()
	}
	return global
}

func Set(cfg Config) {
	global = cfg
}

func SetConfigPath(path string) {
	if path != "" {
		configPath = path
	}
}

func ConfigPath() string {
	return configPath
}

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath
	}
	configPath = path
	cfg := Default()
	created := false

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return Config{}, fmt.Errorf("read config %s: %w", path, err)
		}
		if err := Save(path, cfg); err != nil {
			return Config{}, err
		}
		created = true
	} else {
		var raw fileConfig
		if err := json.Unmarshal(data, &raw); err != nil {
			return Config{}, fmt.Errorf("parse config %s: %w", path, err)
		}
		if err := applyFile(&cfg, raw); err != nil {
			return Config{}, fmt.Errorf("config %s: %w", path, err)
		}
	}

	global = cfg
	if created {
		abs, _ := filepath.Abs(path)
		fmt.Fprintf(os.Stderr, "Создан файл конфигурации: %s\n", abs)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultPath
	}
	data, err := json.MarshalIndent(cfg.toFile(), "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

func (c Config) toFile() fileConfig {
	subPath := c.SubPath
	if subPath == "" {
		subPath = defaultSubPath
	}
	return fileConfig{
		DBPath:       c.DBPath,
		LogPath:      c.LogPath,
		HTTPAddr:     c.HTTPAddr,
		SyncInterval: c.SyncInterval.String(),
		SubDomain:    c.SubDomain,
		SubPath:      subPath,
	}
}

func applyFile(cfg *Config, raw fileConfig) error {
	if raw.DBPath != "" {
		cfg.DBPath = raw.DBPath
	}
	if raw.LogPath != "" {
		cfg.LogPath = raw.LogPath
	}
	if raw.HTTPAddr != "" {
		cfg.HTTPAddr = raw.HTTPAddr
	}
	if raw.SubDomain != "" {
		cfg.SubDomain = strings.TrimRight(strings.TrimSpace(raw.SubDomain), "/")
	}
	if raw.SubPath != "" {
		path, err := normalizeSubPath(raw.SubPath)
		if err != nil {
			return err
		}
		cfg.SubPath = path
	}

	if cfg.SubPath == "" {
		cfg.SubPath = defaultSubPath
	}

	intervalStr := raw.SyncInterval
	if intervalStr == "" {
		intervalStr = "30s"
	}
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return fmt.Errorf("parse sync_interval %q: %w", intervalStr, err)
	}
	cfg.SyncInterval = interval
	return nil
}

func normalizeSubPath(path string) (string, error) {
	p := strings.Trim(strings.TrimSpace(path), "/")
	if p == "" {
		return defaultSubPath, nil
	}
	if strings.Contains(p, "/") || strings.Contains(p, "..") {
		return "", fmt.Errorf("invalid sub_path %q", path)
	}
	return p, nil
}

func Prepare(cfg Config) (Config, error) {
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)
	cfg.LogPath = strings.TrimSpace(cfg.LogPath)
	cfg.HTTPAddr = strings.TrimSpace(cfg.HTTPAddr)
	cfg.SubDomain = strings.TrimRight(strings.TrimSpace(cfg.SubDomain), "/")

	if cfg.DBPath == "" {
		return Config{}, fmt.Errorf("db_path не может быть пустым")
	}
	if cfg.LogPath == "" {
		return Config{}, fmt.Errorf("log_path не может быть пустым")
	}
	if cfg.HTTPAddr == "" {
		return Config{}, fmt.Errorf("http_addr не может быть пустым")
	}
	if cfg.SyncInterval <= 0 {
		return Config{}, fmt.Errorf("sync_interval должен быть больше 0")
	}

	path, err := normalizeSubPath(cfg.SubPath)
	if err != nil {
		return Config{}, err
	}
	cfg.SubPath = path
	return cfg, nil
}

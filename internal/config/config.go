package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	BlitzBaseURL string
	BlitzAPIKey  string
	DefaultName  string
	DBPath       string
	LogPath      string
	HTTPAddr     string
	SyncInterval time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		BlitzBaseURL: os.Getenv("BLITZ_BASE_URL"),
		BlitzAPIKey:  os.Getenv("BLITZ_API_KEY"),
		DefaultName:  envOrDefault("DEFAULT_SERVER_NAME", "default"),
		DBPath:       envOrDefault("DB_PATH", "./panel.db"),
		LogPath:      envOrDefault("LOG_PATH", "./panel.log"),
		HTTPAddr:     envOrDefault("HTTP_ADDR", "0.0.0.0:8787"),
	}

	intervalStr := envOrDefault("SYNC_INTERVAL", "30s")
	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		return Config{}, fmt.Errorf("parse SYNC_INTERVAL: %w", err)
	}
	cfg.SyncInterval = interval

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func EnvOrDefault(key, fallback string) string {
	return envOrDefault(key, fallback)
}

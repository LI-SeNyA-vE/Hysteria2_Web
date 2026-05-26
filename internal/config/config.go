package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	BlitzBaseURL string
	BlitzAPIKey  string
	DBPath       string
	SyncInterval time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		BlitzBaseURL: os.Getenv("BLITZ_BASE_URL"),
		BlitzAPIKey:  os.Getenv("BLITZ_API_KEY"),
		DBPath:       envOrDefault("DB_PATH", "./panel.db"),
	}

	if cfg.BlitzBaseURL == "" {
		return Config{}, fmt.Errorf("BLITZ_BASE_URL is required")
	}
	if cfg.BlitzAPIKey == "" {
		return Config{}, fmt.Errorf("BLITZ_API_KEY is required")
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

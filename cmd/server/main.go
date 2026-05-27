package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hysteria2-web/internal/app"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/httpapi"
	applog "hysteria2-web/internal/log"
)

func main() {
	configPath := flag.String("config", config.DefaultPath, "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	fileLogger, closeLog, err := applog.Open(cfg.LogPath)
	if err != nil {
		log.Fatalf("open log file: %v", err)
	}
	defer closeLog.Close()
	slog.SetDefault(fileLogger)

	a, err := app.OpenWithLogger(cfg.DBPath, fileLogger)
	if err != nil {
		log.Fatalf("open app: %v", err)
	}
	defer a.Close()

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.BlitzSvc.StartTrafficSyncWorker(workerCtx, cfg.SyncInterval)
	fileLogger.Info("traffic sync worker started", "interval", cfg.SyncInterval, "log_path", cfg.LogPath)

	_, listenAddr, err := httpapi.Start(cfg.HTTPAddr, a, fileLogger)
	if err != nil {
		log.Fatalf("http server failed: %v (измените http_addr в %s)", err, *configPath)
	}
	fileLogger.Info("http server started", "addr", listenAddr)

	<-waitForShutdown()
	cancel()
	fileLogger.Info("shutting down")
}

func waitForShutdown() <-chan struct{} {
	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		close(done)
	}()
	return done
}

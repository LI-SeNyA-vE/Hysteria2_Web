package main

import (
	"context"
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
	cfg, err := config.Load()
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

	ctx := context.Background()
	if err = bootstrapDefaultServer(ctx, cfg, a); err != nil {
		log.Fatalf("bootstrap default server: %v", err)
	}

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.BlitzSvc.StartTrafficSyncWorker(workerCtx, cfg.SyncInterval)
	fileLogger.Info("traffic sync worker started", "interval", cfg.SyncInterval, "log_path", cfg.LogPath)

	listenAddr, err := httpapi.Start(cfg.HTTPAddr, a, fileLogger)
	if err != nil {
		log.Fatalf("http server failed: %v (try HTTP_ADDR=0.0.0.0:8787)", err)
	}
	fileLogger.Info("http server started", "addr", listenAddr)

	<-waitForShutdown()
	cancel()
	fileLogger.Info("shutting down")
}

func bootstrapDefaultServer(ctx context.Context, cfg config.Config, a *app.App) error {
	servers, err := a.ServerRepo.List()
	if err != nil {
		return err
	}
	if len(servers) > 0 {
		return nil
	}
	if cfg.BlitzBaseURL == "" || cfg.BlitzAPIKey == "" {
		slog.Warn("no servers in database and BLITZ_BASE_URL/BLITZ_API_KEY not set; add servers via panel")
		return nil
	}

	_, err = a.ServerSvc.CreateServer(ctx, cfg.DefaultName, cfg.BlitzBaseURL, cfg.BlitzAPIKey)
	if err != nil {
		return err
	}
	slog.Info("default blitz server created", "name", cfg.DefaultName)
	return nil
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

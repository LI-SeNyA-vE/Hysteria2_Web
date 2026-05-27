package cli

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hysteria2-web/internal/app"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/httpapi"
	applog "hysteria2-web/internal/log"
)

func RunServe(cfg config.Config) {
	config.Set(cfg)

	fileLogger, closeLog, err := applog.Open(cfg.LogPath)
	if err != nil {
		slog.Error("open log file", "err", err)
		os.Exit(1)
	}
	defer closeLog.Close()
	slog.SetDefault(fileLogger)

	a, err := app.OpenWithLogger(cfg.DBPath, fileLogger)
	if err != nil {
		fileLogger.Error("open app", "err", err)
		os.Exit(1)
	}
	defer a.Close()

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.BlitzSvc.StartTrafficSyncWorker(workerCtx, cfg.SyncInterval)
	fileLogger.Info("traffic sync worker started", "interval", cfg.SyncInterval, "log_path", cfg.LogPath)

	httpServer, listenAddr, err := httpapi.Start(cfg.HTTPAddr, a, fileLogger)
	if err != nil {
		fileLogger.Error("http server failed to start", "addr", cfg.HTTPAddr, "err", err)
		os.Exit(1)
	}
	defer httpServer.Stop()

	fileLogger.Info("panel serve started", "http", listenAddr, "sub", cfg.SubscriptionPublicBase())
	slog.Info("служба запущена", "http", listenAddr, "health", cfg.LocalHealthURL())

	<-waitForShutdown()
	cancel()
	fileLogger.Info("shutting down")
}

func serviceRunning(cfg config.Config) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(cfg.LocalHealthURL())
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
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

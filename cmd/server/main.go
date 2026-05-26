package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/domain/user"
	applog "hysteria2-web/internal/log"
	"hysteria2-web/internal/repository"
	"hysteria2-web/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err = db.AutoMigrate(&server.Server{}, &user.User{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	serverRepo := repository.NewServerRepository(db)
	userRepo := repository.NewUserRepository(db)
	registry := blitz.NewRegistry()

	serverSvc := service.NewServerService(serverRepo, registry)
	blitzSvc := service.NewBlitzService(registry, userRepo, fileLogger)

	ctx := context.Background()
	if err = bootstrapDefaultServer(ctx, cfg, serverSvc, serverRepo); err != nil {
		log.Fatalf("bootstrap default server: %v", err)
	}
	if err = serverSvc.LoadRegistry(ctx); err != nil {
		log.Fatalf("load blitz registry: %v", err)
	}

	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blitzSvc.StartTrafficSyncWorker(workerCtx, cfg.SyncInterval)
	fileLogger.Info("traffic sync worker started", "interval", cfg.SyncInterval, "log_path", cfg.LogPath)

	<-waitForShutdown()
	cancel()
	fileLogger.Info("shutting down")
}

func bootstrapDefaultServer(ctx context.Context, cfg config.Config, serverSvc *service.ServerService, serverRepo server.Repository) error {
	servers, err := serverRepo.List()
	if err != nil {
		return err
	}
	if len(servers) > 0 {
		return nil
	}
	if cfg.BlitzBaseURL == "" || cfg.BlitzAPIKey == "" {
		slog.Warn("no servers in database and BLITZ_BASE_URL/BLITZ_API_KEY not set; add servers via ServerService.CreateServer")
		return nil
	}

	_, err = serverSvc.CreateServer(ctx, cfg.DefaultName, cfg.BlitzBaseURL, cfg.BlitzAPIKey)
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

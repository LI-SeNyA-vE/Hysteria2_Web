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
	"hysteria2-web/internal/domain/user"
	"hysteria2-web/internal/repository"
	"hysteria2-web/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	if err = db.AutoMigrate(&user.User{}); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	blitzClient := blitz.NewClient(cfg.BlitzBaseURL, cfg.BlitzAPIKey)
	svc := service.NewBlitzService(blitzClient, userRepo, slog.Default())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc.StartTrafficSyncWorker(ctx, cfg.SyncInterval)
	slog.Info("traffic sync worker started", "interval", cfg.SyncInterval)

	<-waitForShutdown()
	cancel()
	slog.Info("shutting down")
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

package app

import (
	"context"
	"fmt"
	"os"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/domain/server"
	"hysteria2-web/internal/domain/user"
	"hysteria2-web/internal/repository"
	"hysteria2-web/internal/service"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log/slog"
)

type App struct {
	DB         *gorm.DB
	ServerSvc  *service.ServerService
	BlitzSvc   *service.BlitzService
	SubSvc     *service.SubscriptionService
	ServerRepo *repository.ServerRepository
	UserRepo   *repository.UserRepository
}

func Open(dbPath string) (*App, error) {
	return OpenWithLogger(dbPath, nil)
}

func OpenWithLogger(dbPath string, panelLogger *slog.Logger) (*App, error) {
	if dbPath == "" {
		dbPath = config.Default().DBPath
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(&server.Server{}, &user.User{}); err != nil {
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	serverRepo := repository.NewServerRepository(db)
	userRepo := repository.NewUserRepository(db)
	registry := blitz.NewRegistry()

	serverSvc := service.NewServerService(serverRepo, registry)
	if panelLogger == nil {
		panelLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	blitzSvc := service.NewBlitzService(registry, userRepo, panelLogger)
	subSvc := service.NewSubscriptionService(userRepo, serverSvc, panelLogger)

	if err := serverSvc.LoadRegistry(context.Background()); err != nil {
		return nil, err
	}

	if _, err := blitzSvc.BackfillSubTokens(); err != nil {
		return nil, fmt.Errorf("backfill sub tokens: %w", err)
	}

	return &App{
		DB:         db,
		ServerSvc:  serverSvc,
		BlitzSvc:   blitzSvc,
		SubSvc:     subSvc,
		ServerRepo: serverRepo,
		UserRepo:   userRepo,
	}, nil
}

func (a *App) Close() error {
	sqlDB, err := a.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

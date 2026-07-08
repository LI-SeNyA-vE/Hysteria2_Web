// Package db открывает соединение с БД и предоставляет KV-хелперы.
// По умолчанию используется SQLite (dsn пустой); при непустом dsn — PostgreSQL.
package db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"hysteria2-web/internal/models"
)

// silentNotFoundLogger — GORM-логгер, который молчит о "record not found"
// (это нормальная ветка кода, не ошибка).
type silentNotFoundLogger struct{ logger.Interface }

func (l silentNotFoundLogger) Trace(_ context.Context, _ time.Time, fc func() (string, int64), err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	l.Interface.Trace(context.Background(), time.Now(), fc, err)
}

// DB оборачивает *gorm.DB с удобными хелперами для таблицы Setting.
type DB struct {
	*gorm.DB
}

// Open подключается к БД и выполняет AutoMigrate.
// Если dsn пустой — используется SQLite в dataDir/panel.db.
// Если dsn непустой — используется PostgreSQL.
func Open(dsn, dataDir string) (*DB, error) {
	var dialector gorm.Dialector
	if dsn == "" {
		dialector = sqlite.Open(filepath.Join(dataDir, "panel.db"))
	} else {
		dialector = postgres.Open(dsn)
	}
	g, err := gorm.Open(dialector, &gorm.Config{
		Logger: silentNotFoundLogger{logger.Default.LogMode(logger.Error)},
	})
	if err != nil {
		return nil, fmt.Errorf("подключение к БД: %w", err)
	}
	if err := g.AutoMigrate(
		&models.Server{},
		&models.User{},
		&models.Subscription{},
		&models.Setting{},
	); err != nil {
		return nil, fmt.Errorf("миграция: %w", err)
	}
	return &DB{g}, nil
}

// GetSetting возвращает значение настройки; ("", nil) если ключа нет.
func (d *DB) GetSetting(key string) (string, error) {
	var s models.Setting
	err := d.Where("key = ?", key).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return s.Value, nil
}

// SetSetting создаёт или обновляет настройку.
func (d *DB) SetSetting(key, value string) error {
	return d.Save(&models.Setting{Key: key, Value: value}).Error
}

// GetSettingOrDefault возвращает значение либо def, если ключ пуст/отсутствует.
func (d *DB) GetSettingOrDefault(key, def string) (string, error) {
	v, err := d.GetSetting(key)
	if err != nil {
		return "", err
	}
	if v == "" {
		return def, nil
	}
	return v, nil
}

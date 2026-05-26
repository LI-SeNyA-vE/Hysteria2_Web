package repository

import (
	"errors"
	"fmt"

	"hysteria2-web/internal/domain/server"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

func (r *ServerRepository) Create(s *server.Server) error {
	if err := r.db.Create(s).Error; err != nil {
		return fmt.Errorf("create server: %w", err)
	}
	return nil
}

func (r *ServerRepository) GetByID(id uint) (*server.Server, error) {
	var s server.Server
	err := r.db.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get server by id: %w", err)
	}
	return &s, nil
}

func (r *ServerRepository) GetByName(name string) (*server.Server, error) {
	var s server.Server
	err := r.db.Where("name = ?", name).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get server by name: %w", err)
	}
	return &s, nil
}

func (r *ServerRepository) List() ([]server.Server, error) {
	var servers []server.Server
	if err := r.db.Order("name asc").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	return servers, nil
}

func (r *ServerRepository) ListActive() ([]server.Server, error) {
	var servers []server.Server
	if err := r.db.Where("is_active = ?", true).Order("name asc").Find(&servers).Error; err != nil {
		return nil, fmt.Errorf("list active servers: %w", err)
	}
	return servers, nil
}

func (r *ServerRepository) Update(s *server.Server) error {
	if err := r.db.Save(s).Error; err != nil {
		return fmt.Errorf("update server: %w", err)
	}
	return nil
}

func (r *ServerRepository) Delete(id uint) error {
	result := r.db.Delete(&server.Server{}, id)
	if result.Error != nil {
		return fmt.Errorf("delete server: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete server: not found")
	}
	return nil
}

package server

import "time"

type Server struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex;not null"`
	BaseURL   string `gorm:"not null"`
	APIKey    string `gorm:"not null"`
	IsActive  bool   `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	Create(s *Server) error
	GetByID(id uint) (*Server, error)
	GetByName(name string) (*Server, error)
	List() ([]Server, error)
	ListActive() ([]Server, error)
	Update(s *Server) error
	Delete(id uint) error
}

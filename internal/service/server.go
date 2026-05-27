package service

import (
	"context"
	"errors"
	"fmt"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/domain/server"
)

var (
	ErrServerExists   = errors.New("server already exists")
	ErrServerNotFound = errors.New("server not found")
)

type ServerService struct {
	repo     server.Repository
	registry *blitz.Registry
}

func NewServerService(repo server.Repository, registry *blitz.Registry) *ServerService {
	return &ServerService{
		repo:     repo,
		registry: registry,
	}
}

func (s *ServerService) CreateServer(ctx context.Context, name, baseURL, apiKey string) (*server.Server, error) {
	existing, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrServerExists
	}

	srv := &server.Server{
		Name:     name,
		BaseURL:  baseURL,
		APIKey:   apiKey,
		IsActive: true,
	}
	if err := s.repo.Create(srv); err != nil {
		return nil, err
	}

	client := blitz.NewClient(baseURL, apiKey)
	if _, err := client.GetServerStatus(ctx); err != nil {
		_ = s.repo.Delete(srv.ID)
		return nil, fmt.Errorf("verify blitz connection: %w", err)
	}

	s.registry.Register(*srv)
	return srv, nil
}

func (s *ServerService) DeleteServer(ctx context.Context, id uint) error {
	srv, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if srv == nil {
		return ErrServerNotFound
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.registry.Unregister(id)
	return nil
}

func (s *ServerService) ListServers(ctx context.Context) ([]server.Server, error) {
	return s.repo.List()
}

func (s *ServerService) GetServer(ctx context.Context, id uint) (*server.Server, error) {
	srv, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if srv == nil {
		return nil, ErrServerNotFound
	}
	return srv, nil
}

func (s *ServerService) GetClient(serverID uint) (*blitz.HTTPClient, error) {
	return s.registry.Get(serverID)
}

func (s *ServerService) LoadRegistry(ctx context.Context) error {
	servers, err := s.repo.ListActive()
	if err != nil {
		return err
	}
	s.registry.LoadAll(servers)
	return nil
}

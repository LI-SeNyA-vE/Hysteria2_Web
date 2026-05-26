package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/domain/user"
)

const bytesPerGB = 1024 * 1024 * 1024

var ErrUserExists = errors.New("user already exists")

type BlitzService struct {
	registry *blitz.Registry
	repo     user.Repository
	logger   *slog.Logger
}

func NewBlitzService(registry *blitz.Registry, repo user.Repository, logger *slog.Logger) *BlitzService {
	if logger == nil {
		logger = slog.Default()
	}
	return &BlitzService{
		registry: registry,
		repo:     repo,
		logger:   logger,
	}
}

func (s *BlitzService) Client(serverID uint) (*blitz.HTTPClient, error) {
	return s.registry.Get(serverID)
}

func (s *BlitzService) AddUser(ctx context.Context, serverID uint, username, password string, trafficLimitGB, expirationDays int) error {
	existing, err := s.repo.GetByUsername(serverID, username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserExists
	}

	client, err := s.registry.Get(serverID)
	if err != nil {
		return err
	}

	req := blitz.AddUserRequest{
		Username:       username,
		Password:       &password,
		TrafficLimit:   trafficLimitGB,
		ExpirationDays: expirationDays,
		Unlimited:      false,
	}
	s.logger.Info("blitz add user request",
		"server_id", serverID,
		"username", username,
		"traffic_limit_gb", trafficLimitGB,
		"expiration_days", expirationDays,
	)
	if err := client.AddUser(ctx, req); err != nil {
		s.logger.Error("blitz add user failed",
			"server_id", serverID,
			"username", username,
			"err", err,
		)
		return fmt.Errorf("add user to blitz: %w", err)
	}
	s.logger.Info("user added in blitz", "server_id", serverID, "username", username)

	u := &user.User{
		ServerID:       serverID,
		Username:       username,
		AuthPassword:   password,
		TrafficLimit:   trafficLimitGB,
		TrafficUsed:    0,
		IsActive:       true,
		ExpirationDays: expirationDays,
	}
	if err := s.repo.Create(u); err != nil {
		s.logger.Error("persist user failed",
			"server_id", serverID,
			"username", username,
			"err", err,
		)
		return fmt.Errorf("persist user: %w", err)
	}
	s.logger.Info("user saved locally", "server_id", serverID, "username", username)
	return nil
}

func (s *BlitzService) KickUser(ctx context.Context, serverID uint, username string) error {
	client, err := s.registry.Get(serverID)
	if err != nil {
		return err
	}

	err = client.RemoveUser(ctx, username)
	if err != nil && !errors.Is(err, blitz.ErrNotFound) {
		return fmt.Errorf("remove user from blitz: %w", err)
	}

	if err := s.repo.Deactivate(serverID, username); err != nil {
		return fmt.Errorf("deactivate user locally: %w", err)
	}
	return nil
}

func (s *BlitzService) SyncTraffic(ctx context.Context) error {
	activeUsers, err := s.repo.ListActive()
	if err != nil {
		return err
	}

	usersByServer := make(map[uint][]user.User)
	for _, u := range activeUsers {
		usersByServer[u.ServerID] = append(usersByServer[u.ServerID], u)
	}

	for serverID, users := range usersByServer {
		if err := s.syncTrafficForServer(ctx, serverID, users); err != nil {
			return err
		}
	}
	return nil
}

func (s *BlitzService) SyncTrafficForServer(ctx context.Context, serverID uint) error {
	users, err := s.repo.ListActiveByServer(serverID)
	if err != nil {
		return err
	}
	return s.syncTrafficForServer(ctx, serverID, users)
}

func (s *BlitzService) syncTrafficForServer(ctx context.Context, serverID uint, users []user.User) error {
	if len(users) == 0 {
		return nil
	}

	client, err := s.registry.Get(serverID)
	if err != nil {
		return err
	}

	blitzUsers, err := client.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("list blitz users for server %d: %w", serverID, err)
	}

	trafficByUsername := make(map[string]int64, len(blitzUsers))
	for _, bu := range blitzUsers {
		trafficByUsername[bu.Username] = totalBytes(bu.UploadBytes, bu.DownloadBytes)
	}

	for _, u := range users {
		currentTotal, ok := trafficByUsername[u.Username]
		if !ok {
			s.logger.Warn("blitz user not found during traffic sync",
				"server_id", serverID,
				"username", u.Username,
			)
			continue
		}

		delta := currentTotal - u.LastBlitzTotalBytes
		if delta < 0 {
			delta = currentTotal
		}

		pending := u.PendingBytes + delta
		trafficUsed := u.TrafficUsed + int(pending/bytesPerGB)
		pending %= bytesPerGB

		update := user.TrafficUpdate{
			TrafficUsed:         trafficUsed,
			PendingBytes:        pending,
			LastBlitzTotalBytes: currentTotal,
		}
		if err := s.repo.UpdateTraffic(serverID, u.Username, update); err != nil {
			return fmt.Errorf("update traffic for %s: %w", u.Username, err)
		}

		if u.TrafficLimit > 0 && trafficUsed >= u.TrafficLimit {
			s.logger.Info("traffic limit exceeded, kicking user",
				"server_id", serverID,
				"username", u.Username,
				"traffic_used_gb", trafficUsed,
				"traffic_limit_gb", u.TrafficLimit,
			)
			if err := s.KickUser(ctx, serverID, u.Username); err != nil {
				return fmt.Errorf("kick user %s: %w", u.Username, err)
			}
		}
	}
	return nil
}

func (s *BlitzService) StartTrafficSyncWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := s.SyncTraffic(ctx); err != nil {
					s.logger.Error("traffic sync failed", "err", err)
				}
			}
		}
	}()
}

func totalBytes(upload, download *int64) int64 {
	var total int64
	if upload != nil {
		total += *upload
	}
	if download != nil {
		total += *download
	}
	return total
}

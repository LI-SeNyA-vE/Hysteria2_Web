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
	blitz  blitz.Client
	repo   user.Repository
	logger *slog.Logger
}

func NewBlitzService(client blitz.Client, repo user.Repository, logger *slog.Logger) *BlitzService {
	if logger == nil {
		logger = slog.Default()
	}
	return &BlitzService{
		blitz:  client,
		repo:   repo,
		logger: logger,
	}
}

func (s *BlitzService) AddUser(ctx context.Context, username, password string, trafficLimitGB, expirationDays int) error {
	if expirationDays <= 0 {
		expirationDays = 30
	}

	existing, err := s.repo.GetByUsername(username)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrUserExists
	}

	req := blitz.AddUserRequest{
		Username:       username,
		Password:       &password,
		TrafficLimit:   trafficLimitGB,
		ExpirationDays: expirationDays,
		Unlimited:      false,
	}
	if err := s.blitz.AddUser(ctx, req); err != nil {
		return fmt.Errorf("add user to blitz: %w", err)
	}

	u := &user.User{
		Username:       username,
		AuthPassword:   password,
		TrafficLimit:   trafficLimitGB,
		TrafficUsed:    0,
		IsActive:       true,
		ExpirationDays: expirationDays,
	}
	if err := s.repo.Create(u); err != nil {
		return fmt.Errorf("persist user: %w", err)
	}
	return nil
}

func (s *BlitzService) KickUser(ctx context.Context, username string) error {
	err := s.blitz.RemoveUser(ctx, username)
	if err != nil && !errors.Is(err, blitz.ErrNotFound) {
		return fmt.Errorf("remove user from blitz: %w", err)
	}

	if err := s.repo.Deactivate(username); err != nil {
		return fmt.Errorf("deactivate user locally: %w", err)
	}
	return nil
}

func (s *BlitzService) SyncTraffic(ctx context.Context) error {
	blitzUsers, err := s.blitz.ListUsers(ctx)
	if err != nil {
		return fmt.Errorf("list blitz users: %w", err)
	}

	trafficByUsername := make(map[string]int64, len(blitzUsers))
	for _, bu := range blitzUsers {
		trafficByUsername[bu.Username] = totalBytes(bu.UploadBytes, bu.DownloadBytes)
	}

	activeUsers, err := s.repo.ListActive()
	if err != nil {
		return err
	}

	for _, u := range activeUsers {
		currentTotal, ok := trafficByUsername[u.Username]
		if !ok {
			s.logger.Warn("blitz user not found during traffic sync", "username", u.Username)
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
		if err := s.repo.UpdateTraffic(u.Username, update); err != nil {
			return fmt.Errorf("update traffic for %s: %w", u.Username, err)
		}

		if trafficUsed >= u.TrafficLimit {
			s.logger.Info("traffic limit exceeded, kicking user",
				"username", u.Username,
				"traffic_used_gb", trafficUsed,
				"traffic_limit_gb", u.TrafficLimit,
			)
			if err := s.KickUser(ctx, u.Username); err != nil {
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

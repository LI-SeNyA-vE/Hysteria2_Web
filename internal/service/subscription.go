package service

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"time"

	"hysteria2-web/internal/blitz"
	"hysteria2-web/internal/domain/user"
)

var (
	ErrSubscriptionNotFound  = errors.New("subscription not found")
	ErrSubscriptionForbidden = errors.New("subscription forbidden")
	ErrSubscriptionNoURIs    = errors.New("subscription has no connection URIs")
)

type SubscriptionService struct {
	users   user.Repository
	servers *ServerService
	logger  *slog.Logger
}

func NewSubscriptionService(users user.Repository, servers *ServerService, logger *slog.Logger) *SubscriptionService {
	if logger == nil {
		logger = slog.Default()
	}
	return &SubscriptionService{
		users:   users,
		servers: servers,
		logger:  logger,
	}
}

func (s *SubscriptionService) BuildSubscription(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", ErrSubscriptionNotFound
	}

	records, err := s.users.ListBySubToken(token)
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", ErrSubscriptionNotFound
	}

	now := time.Now().UTC()
	for i := range records {
		if !user.IsSubscriptionValid(&records[i], now) {
			s.logger.Warn("subscription rejected",
				"sub_token", token,
				"username", records[i].Username,
				"server_id", records[i].ServerID,
				"is_active", records[i].IsActive,
				"expires_at", records[i].ExpiresAt,
				"traffic_used_gb", records[i].TrafficUsed,
				"traffic_limit_gb", records[i].TrafficLimit,
			)
			return "", ErrSubscriptionForbidden
		}
	}

	username := records[0].Username
	servers, err := s.servers.ListServers(ctx)
	if err != nil {
		return "", err
	}

	var lines []string

	for _, srv := range servers {
		if !srv.IsActive {
			continue
		}

		local, err := s.users.GetByUsername(srv.ID, username)
		if err != nil {
			return "", err
		}
		if local == nil || !local.IsActive {
			continue
		}

		client, err := s.servers.GetClient(srv.ID)
		if err != nil {
			s.logger.Warn("subscription skip server",
				"server_id", srv.ID,
				"server_name", srv.Name,
				"err", err,
			)
			continue
		}

		uri, err := client.ShowUserURI(ctx, username)
		if err != nil {
			s.logger.Warn("subscription uri fetch failed",
				"server_id", srv.ID,
				"server_name", srv.Name,
				"username", username,
				"err", err,
			)
			continue
		}
		if uri.Error != nil && strings.TrimSpace(*uri.Error) != "" {
			s.logger.Warn("subscription uri error from blitz",
				"server_id", srv.ID,
				"server_name", srv.Name,
				"username", username,
				"err", *uri.Error,
			)
			continue
		}

		for _, line := range blitz.CollectRelabeledHy2URIs(uri, srv.Name) {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		return "", ErrSubscriptionNoURIs
	}

	payload := strings.Join(lines, "\n")
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	s.logger.Info("subscription built",
		"sub_token", token,
		"username", username,
		"uri_count", len(lines),
	)
	return encoded, nil
}

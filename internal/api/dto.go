package api

import (
	"time"

	"hysteria2-web/internal/models"
)

// GiB — делитель для конвертации байтов в гигабайты, показываемые фронтом.
const GiB = 1 << 30

// serverDTO зеркалит frontend/src/types Server (camelCase JSON).
type serverDTO struct {
	ID         uint   `json:"id"`
	Role       string `json:"role"`
	Name       string `json:"name"`
	PublicIP   string `json:"publicIp"`
	PanelURL   string `json:"panelUrl"`
	Hy2Port    int    `json:"hy2Port"`
	Hy2Version string `json:"hy2Version"`
	CreatedAt  string `json:"createdAt"`
}

func toServerDTO(s models.Server) serverDTO {
	return serverDTO{
		ID:         s.ID,
		Role:       s.Role,
		Name:       s.Name,
		PublicIP:   s.PublicIP,
		PanelURL:   s.PanelURL,
		Hy2Port:    s.Hy2Port,
		Hy2Version: s.Hy2Version,
		CreatedAt:  s.CreatedAt.Format(time.RFC3339),
	}
}

// userDTO зеркалит frontend/src/types User.
type userDTO struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	UUID           string  `json:"uuid"`
	Password       string  `json:"password"`
	TrafficLimitGb float64 `json:"trafficLimitGb"`
	TrafficUsedGb  float64 `json:"trafficUsedGb"`
	ExpireAt       *string `json:"expireAt"`
	IsActive       bool    `json:"isActive"`
	ServerID       uint    `json:"serverId"`
	CreatedAt      string  `json:"createdAt"`
}

func toUserDTO(u models.User) userDTO {
	var expireAt *string
	if u.ExpireAt != nil {
		s := u.ExpireAt.Format(time.RFC3339)
		expireAt = &s
	}
	return userDTO{
		ID:             u.ID,
		Name:           u.Name,
		UUID:           u.UUID,
		Password:       u.Password,
		TrafficLimitGb: float64(u.TrafficLimitBytes) / GiB,
		TrafficUsedGb:  float64(u.TrafficUsedBytes) / GiB,
		ExpireAt:       expireAt,
		IsActive:       u.IsActive,
		ServerID:       u.ServerID,
		CreatedAt:      u.CreatedAt.Format(time.RFC3339),
	}
}

// createUserRequest зеркалит CreateUserRequest.
type createUserRequest struct {
	Name           string  `json:"name"`
	TrafficLimitGb float64 `json:"trafficLimitGb"`
	ExpireDays     *int    `json:"expireDays"`
	ServerID       uint    `json:"serverId"`
}

// updateUserRequest — частичное обновление (Partial<User>). Только не-nil поля применяются.
type updateUserRequest struct {
	Name           *string  `json:"name"`
	TrafficLimitGb *float64 `json:"trafficLimitGb"`
	IsActive       *bool    `json:"isActive"`
	ExpireAt       *string  `json:"expireAt"`
}

// subscriptionDTO зеркалит Subscription.
type subscriptionDTO struct {
	ID             uint    `json:"id"`
	UserID         uint    `json:"userId"`
	UserName       string  `json:"userName"`
	Token          string  `json:"token"`
	Name           string  `json:"name"`
	LastAccessedAt *string `json:"lastAccessedAt"`
	CreatedAt      string  `json:"createdAt"`
}

func toSubscriptionDTO(s models.Subscription, userName string) subscriptionDTO {
	var last *string
	if s.LastAccessedAt != nil {
		v := s.LastAccessedAt.Format(time.RFC3339)
		last = &v
	}
	return subscriptionDTO{
		ID:             s.ID,
		UserID:         s.UserID,
		UserName:       userName,
		Token:          s.Token,
		Name:           s.Name,
		LastAccessedAt: last,
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
	}
}

type createSubscriptionRequest struct {
	UserID uint   `json:"userId"`
	Name   string `json:"name"`
}

// hysteriaStatusDTO зеркалит HysteriaStatus.
type hysteriaStatusDTO struct {
	Installed bool   `json:"installed"`
	Running   bool   `json:"running"`
	Version   string `json:"version"`
	Port      int    `json:"port"`
}

// hysteriaConfigDTO зеркалит HysteriaConfig.
type hysteriaConfigDTO struct {
	Port          int    `json:"port"`
	ObfsPassword  string `json:"obfsPassword"`
	MasqueradeURL string `json:"masqueradeUrl"`
	CertSHA256    string `json:"certSha256"`
	BandwidthUp   string `json:"bandwidthUp"`
	BandwidthDown string `json:"bandwidthDown"`
	SNI           string `json:"sni"`
}

// dashboardStatsDTO зеркалит DashboardStats.
type dashboardStatsDTO struct {
	TotalUsers     int               `json:"totalUsers"`
	ActiveUsers    int               `json:"activeUsers"`
	TotalTrafficGb float64           `json:"totalTrafficGb"`
	Uptime         string            `json:"uptime"`
	Hysteria       hysteriaStatusDTO `json:"hysteria"`
}

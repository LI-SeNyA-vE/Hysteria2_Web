package user

import "time"

type User struct {
	ID                  uint   `gorm:"primaryKey"`
	ServerID            uint   `gorm:"not null;uniqueIndex:idx_server_username"`
	Username            string `gorm:"not null;uniqueIndex:idx_server_username"`
	AuthPassword        string `gorm:"not null"`
	SubToken            string `gorm:"index"`
	TrafficLimit        int    `gorm:"not null"`
	TrafficUsed         int    `gorm:"not null;default:0"`
	IsActive            bool   `gorm:"not null;default:true"`
	LastBlitzTotalBytes int64  `gorm:"not null;default:0"`
	PendingBytes        int64  `gorm:"not null;default:0"`
	ExpirationDays      int    `gorm:"not null;default:30"`
	ExpiresAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type TrafficUpdate struct {
	TrafficUsed         int
	PendingBytes        int64
	LastBlitzTotalBytes int64
}

type Repository interface {
	Create(u *User) error
	GetByUsername(serverID uint, username string) (*User, error)
	ListBySubToken(token string) ([]User, error)
	GetSubTokenByUsername(username string) (string, error)
	ListAll() ([]User, error)
	ListByUsername(username string) ([]User, error)
	UpdateSubToken(serverID uint, username, token string) error
	ListActive() ([]User, error)
	ListActiveByServer(serverID uint) ([]User, error)
	Deactivate(serverID uint, username string) error
	UpdateTraffic(serverID uint, username string, update TrafficUpdate) error
}

func IsSubscriptionValid(u *User, now time.Time) bool {
	if u == nil || !u.IsActive {
		return false
	}
	if !u.ExpiresAt.IsZero() && now.After(u.ExpiresAt) {
		return false
	}
	if u.TrafficLimit > 0 && u.TrafficUsed >= u.TrafficLimit {
		return false
	}
	return true
}

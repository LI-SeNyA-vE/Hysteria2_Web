package user

import "time"

type User struct {
	ID                  uint   `gorm:"primaryKey"`
	Username            string `gorm:"uniqueIndex;not null"`
	AuthPassword        string `gorm:"not null"`
	TrafficLimit        int    `gorm:"not null"`
	TrafficUsed         int    `gorm:"not null;default:0"`
	IsActive            bool   `gorm:"not null;default:true"`
	LastBlitzTotalBytes int64  `gorm:"not null;default:0"`
	PendingBytes        int64  `gorm:"not null;default:0"`
	ExpirationDays      int    `gorm:"not null;default:30"`
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
	GetByUsername(username string) (*User, error)
	ListActive() ([]User, error)
	Deactivate(username string) error
	UpdateTraffic(username string, update TrafficUpdate) error
}

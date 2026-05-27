package repository

import (
	"errors"
	"fmt"

	"hysteria2-web/internal/domain/user"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u *user.User) error {
	if err := r.db.Select(
		"ServerID", "Username", "AuthPassword", "SubToken",
		"TrafficLimit", "TrafficUsed", "IsActive",
		"LastBlitzTotalBytes", "PendingBytes", "ExpirationDays", "ExpiresAt",
	).Create(u).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByUsername(serverID uint, username string) (*user.User, error) {
	var u user.User
	err := r.db.Where("server_id = ? AND username = ?", serverID, username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetBySubToken(token string) (*user.User, error) {
	var u user.User
	err := r.db.Where("sub_token = ?", token).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by sub token: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) ListBySubToken(token string) ([]user.User, error) {
	var users []user.User
	if err := r.db.Where("sub_token = ?", token).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users by sub token: %w", err)
	}
	return users, nil
}

func (r *UserRepository) GetSubTokenByUsername(username string) (string, error) {
	var u user.User
	err := r.db.Where("username = ? AND sub_token <> ''", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get sub token by username: %w", err)
	}
	return u.SubToken, nil
}

func (r *UserRepository) ListByUsername(username string) ([]user.User, error) {
	var users []user.User
	if err := r.db.Where("username = ?", username).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users by username: %w", err)
	}
	return users, nil
}

func (r *UserRepository) UpdateSubToken(serverID uint, username, token string) error {
	err := r.db.Model(&user.User{}).
		Where("server_id = ? AND username = ?", serverID, username).
		Update("sub_token", token).Error
	if err != nil {
		return fmt.Errorf("update sub token: %w", err)
	}
	return nil
}

func (r *UserRepository) ListAll() ([]user.User, error) {
	var users []user.User
	if err := r.db.Order("server_id asc, username asc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) ListActive() ([]user.User, error) {
	var users []user.User
	if err := r.db.Where("is_active = ?", true).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list active users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) ListActiveByServer(serverID uint) ([]user.User, error) {
	var users []user.User
	if err := r.db.Where("server_id = ? AND is_active = ?", serverID, true).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list active users by server: %w", err)
	}
	return users, nil
}

func (r *UserRepository) Deactivate(serverID uint, username string) error {
	result := r.db.Model(&user.User{}).
		Where("server_id = ? AND username = ?", serverID, username).
		Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("deactivate user: %w", result.Error)
	}
	return nil
}

func (r *UserRepository) DeactivateAllByServer(serverID uint) error {
	result := r.db.Model(&user.User{}).
		Where("server_id = ?", serverID).
		Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("deactivate users by server: %w", result.Error)
	}
	return nil
}

func (r *UserRepository) UpdateTraffic(serverID uint, username string, update user.TrafficUpdate) error {
	err := r.db.Model(&user.User{}).
		Where("server_id = ? AND username = ?", serverID, username).
		Updates(map[string]interface{}{
			"traffic_used":           update.TrafficUsed,
			"pending_bytes":          update.PendingBytes,
			"last_blitz_total_bytes": update.LastBlitzTotalBytes,
		}).Error
	if err != nil {
		return fmt.Errorf("update traffic: %w", err)
	}
	return nil
}

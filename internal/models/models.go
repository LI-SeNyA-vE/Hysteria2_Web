// Package models содержит GORM-модели панели.
package models

import "time"

// Роли серверов в каскадной сети.
const (
	RoleMain      = "main"
	RoleNode1     = "node1"
	RoleNode2     = "node2"
	RoleMainNode1 = "main_node1"
)

// Server — узел сети (сама панель либо зарегистрировавшаяся нода).
type Server struct {
	ID         uint   `gorm:"primaryKey"`
	Role       string `gorm:"index"`
	Name       string `gorm:"uniqueIndex"`
	PublicIP   string `gorm:"column:public_ip"`
	PanelURL   string `gorm:"column:panel_url"`
	Hy2Port    int    `gorm:"column:hy2_port"`
	Hy2Version string `gorm:"column:hy2_version"`
	CertSHA256 string `gorm:"column:cert_sha256"`
	// CascadeTarget — имя (Name) ноды node2, через которую этот node1 выходит в интернет.
	// Пусто = использовать первую доступную node2.
	CascadeTarget string `gorm:"column:cascade_target"`
	// Per-node overrides: пусто = используется глобальное значение из таблицы Setting.
	BandwidthUp   string `gorm:"column:bandwidth_up"`
	BandwidthDown string `gorm:"column:bandwidth_down"`
	MasqueradeURL string `gorm:"column:masquerade_url"`
	LastSeenAt    *time.Time `gorm:"column:last_seen_at"`
	CreatedAt     time.Time
}

// User — клиент VPN. Трафик хранится в байтах, срок — nil = бессрочно.
type User struct {
	ID                uint   `gorm:"primaryKey"`
	Name              string `gorm:"uniqueIndex"`
	UUID              string
	Password          string
	TrafficLimitBytes int64
	TrafficUsedBytes  int64
	ExpireAt          *time.Time `gorm:"column:expire_at"`
	IsActive          bool       `gorm:"column:is_active"`
	ServerID          uint       `gorm:"column:server_id;index"`
	CreatedAt         time.Time
}

// Subscription — токен подписки, привязанный к пользователю.
type Subscription struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"column:user_id;index"`
	Token          string `gorm:"uniqueIndex"`
	Name           string
	LastAccessedAt *time.Time `gorm:"column:last_accessed_at"`
	CreatedAt      time.Time
}

// Setting — key/value хранилище для секретов и глобальных настроек.
type Setting struct {
	Key   string `gorm:"primaryKey"`
	Value string
}

// Ключи таблицы Setting.
const (
	SettingAdminPasswordHash = "admin_password_hash"
	SettingJWTSecret         = "jwt_secret"
	SettingNodeToken         = "node_token"
	SettingObfsPassword      = "obfs_password"
	SettingMasqueradeURL     = "masquerade_url"
	SettingCascadeUser       = "cascade_user"
	SettingCascadePassword   = "cascade_password"
	SettingHy2Running        = "hy2_running"    // желаемое состояние: "1"/"0"
	SettingStatsSecret       = "stats_secret"   // секрет Traffic Stats API hysteria2
	SettingBandwidthUp       = "bandwidth_up"   // лимит исходящего, напр. "100 mbps"
	SettingBandwidthDown     = "bandwidth_down" // лимит входящего, напр. "1 gbps"
	SettingHy2Port           = "hy2_port"       // UDP-порт hysteria2 (override panel.yaml)
	SettingSNI               = "sni"            // SNI для клиентских URI, напр. "yandex.ru"
)

// Package cluster реализует протокол между главной панелью и нодами.
// Ноды POST-ят heartbeats на main и получают DesiredNodeConfig в ответ (pull-модель).
package cluster

// RegisterRequest — нода регистрируется на main при старте.
type RegisterRequest struct {
	Role       string `json:"role"`
	Name       string `json:"name"`
	PublicIP   string `json:"publicIp"`
	Hy2Port    int    `json:"hy2Port"`
	Hy2Version string `json:"hy2Version"`
	CertSHA256 string `json:"certSha256"`
	// CascadeTarget — имя node2, через которую этот node1 выходит в интернет.
	// Пусто = первая доступная node2.
	CascadeTarget string `json:"cascadeTarget,omitempty"`
}

// UsageStat — байты трафика одного пользователя за период (дельта).
type UsageStat struct {
	TX int64 `json:"tx"`
	RX int64 `json:"rx"`
}

// HeartbeatRequest — нода шлёт каждые 10 секунд.
type HeartbeatRequest struct {
	Name         string               `json:"name"`
	Running      bool                 `json:"running"`
	CertSHA256   string               `json:"certSha256"`
	Usage        map[string]UsageStat `json:"usage"`
	Logs         []string             `json:"logs,omitempty"`         // последние строки из logbuf hysteria2
	PanelVersion string               `json:"panelVersion,omitempty"` // текущая версия панели на ноде
}

// NodeServerConfig — параметры, которые main передаёт ноде для генерации server.yaml.
type NodeServerConfig struct {
	ObfsPassword  string            `json:"obfsPassword"`
	MasqueradeURL string            `json:"masqueradeUrl"`
	StatsSecret   string            `json:"statsSecret"`
	Users         map[string]string `json:"users"` // имя → пароль
	BandwidthUp   string            `json:"bandwidthUp,omitempty"`
	BandwidthDown string            `json:"bandwidthDown,omitempty"`
}

// CascadeClientConfig — параметры для client.yaml на node1 (каскад на node2).
type CascadeClientConfig struct {
	ServerAddr   string `json:"serverAddr"`
	UserName     string `json:"userName"`
	Password     string `json:"password"`
	ObfsPassword string `json:"obfsPassword"`
	PinSHA256    string `json:"pinSha256"`
}

// DesiredNodeConfig — желаемое состояние ноды; main отвечает им на register/heartbeat.
// Version меняется только при изменении контента — нода применяет только при version != lastApplied.
type DesiredNodeConfig struct {
	Version              int64                `json:"version"`
	ServerConfig         NodeServerConfig     `json:"serverConfig"`
	CascadeClient        *CascadeClientConfig `json:"cascadeClient,omitempty"`
	Run                  bool                 `json:"run"`
	DesiredPanelVersion  string               `json:"desiredPanelVersion,omitempty"` // если задана — нода должна обновиться
}

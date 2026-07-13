package cluster

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"hysteria2-web/internal/db"
	"hysteria2-web/internal/models"
)

const maxNodeLogLines = 500

// Registry — серверная сторона протокола (работает на main).
type Registry struct {
	db       *db.DB
	logsMu   sync.RWMutex
	nodeLogs map[string][]string // имя ноды → последние строки логов
}

func NewRegistry(d *db.DB) *Registry {
	return &Registry{db: d, nodeLogs: make(map[string][]string)}
}

// Register создаёт/обновляет запись Server и возвращает DesiredNodeConfig.
func (r *Registry) Register(req RegisterRequest) (DesiredNodeConfig, error) {
	now := time.Now()
	var s models.Server
	err := r.db.Where("name = ?", req.Name).First(&s).Error
	if err != nil {
		s = models.Server{
			Name:          req.Name,
			Role:          req.Role,
			PublicIP:      req.PublicIP,
			Hy2Port:       req.Hy2Port,
			Hy2Version:    req.Hy2Version,
			CertSHA256:    req.CertSHA256,
			CascadeTarget: req.CascadeTarget,
			LastSeenAt:    &now,
		}
		if err := r.db.Create(&s).Error; err != nil {
			return DesiredNodeConfig{}, err
		}
	} else {
		updates := map[string]any{
			"role":         req.Role,
			"public_ip":    req.PublicIP,
			"hy2_port":     req.Hy2Port,
			"hy2_version":  req.Hy2Version,
			"cert_sha256":  req.CertSHA256,
			"last_seen_at": &now,
		}
		if req.CascadeTarget != "" {
			updates["cascade_target"] = req.CascadeTarget
		}
		r.db.Model(&s).Updates(updates)
	}
	log.Printf("cluster: нода «%s» зарегистрирована (роль: %s, ip: %s)", req.Name, req.Role, req.PublicIP)
	return r.buildDesiredConfig(req.Role, req.Name), nil
}

// Heartbeat обновляет lastSeenAt, аккумулирует трафик и возвращает DesiredNodeConfig.
func (r *Registry) Heartbeat(req HeartbeatRequest) (DesiredNodeConfig, error) {
	now := time.Now()
	var s models.Server
	if err := r.db.Where("name = ?", req.Name).First(&s).Error; err != nil {
		return DesiredNodeConfig{}, err
	}

	updates := map[string]any{
		"last_seen_at": &now,
		"hy2_running":  req.Running,
	}
	if req.CertSHA256 != "" && req.CertSHA256 != s.CertSHA256 {
		updates["cert_sha256"] = req.CertSHA256
	}
	if req.PanelVersion != "" && req.PanelVersion != s.PanelVersion {
		updates["panel_version"] = req.PanelVersion
	}
	r.db.Model(&s).Updates(updates)

	// Сохраняем логи ноды в памяти (обрезаем до maxNodeLogLines)
	if len(req.Logs) > 0 {
		r.logsMu.Lock()
		lines := req.Logs
		if len(lines) > maxNodeLogLines {
			lines = lines[len(lines)-maxNodeLogLines:]
		}
		r.nodeLogs[req.Name] = lines
		r.logsMu.Unlock()
	}

	// Аккумулируем трафик пользователей
	for name, usage := range req.Usage {
		delta := usage.TX + usage.RX
		if delta <= 0 {
			continue
		}
		if err := r.db.Exec(
			"UPDATE users SET traffic_used_bytes = traffic_used_bytes + ? WHERE name = ?",
			delta, name,
		).Error; err != nil {
			log.Printf("cluster: трафик «%s»: %v", name, err)
		}
	}

	return r.buildDesiredConfig(s.Role, s.Name), nil
}

// buildDesiredConfig формирует конфиг, который нода должна применить.
// serverName — уникальное имя ноды (hostname), нужно для поиска её CascadeTarget.
func (r *Registry) buildDesiredConfig(role, serverName string) DesiredNodeConfig {
	obfs, _ := r.db.GetSetting(models.SettingObfsPassword)
	globalMasq, _ := r.db.GetSettingOrDefault(models.SettingMasqueradeURL, "https://news.ycombinator.com/")
	globalBwUp, _ := r.db.GetSetting(models.SettingBandwidthUp)
	globalBwDown, _ := r.db.GetSetting(models.SettingBandwidthDown)
	statsSecret, _ := r.db.GetSetting(models.SettingStatsSecret)

	// Ищем запись сервера для per-node override'ов
	var srvRecord models.Server
	_ = r.db.Where("name = ?", serverName).First(&srvRecord).Error
	masq := firstNonEmpty(srvRecord.MasqueradeURL, globalMasq)
	bwUp := firstNonEmpty(srvRecord.BandwidthUp, globalBwUp)
	bwDown := firstNonEmpty(srvRecord.BandwidthDown, globalBwDown)

	// Каскадные учётные данные — генерируются один раз и хранятся в Settings.
	cascadeUser, cascadePass := r.ensureCascadeCredentials()

	users := make(map[string]string)
	runNode := true
	var cascadeClient *CascadeClientConfig

	switch role {
	case models.RoleNode2:
		// node2 принимает только системного пользователя-каскада
		if cascadeUser != "" {
			users[cascadeUser] = cascadePass
		} else {
			runNode = false
		}

	case models.RoleNode1, models.RoleMainNode1:
		// node1 / main_node1 принимают всех активных пользователей + каскадный клиент
		var activeUsers []models.User
		r.db.Where("is_active = ?", true).Find(&activeUsers)
		for _, u := range activeUsers {
			users[u.Name] = u.Password
		}
		cascadeClient = r.buildCascadeClientForNode1(srvRecord.CascadeTarget, cascadeUser, cascadePass, obfs)
		// Не запускаем hysteria без юзеров — он отказывается стартовать с пустым userpass
		if len(users) == 0 {
			runNode = false
		}

	default:
		var activeUsers []models.User
		r.db.Where("is_active = ?", true).Find(&activeUsers)
		for _, u := range activeUsers {
			users[u.Name] = u.Password
		}
		if len(users) == 0 {
			runNode = false
		}
	}

	serverCfg := NodeServerConfig{
		ObfsPassword:  obfs,
		MasqueradeURL: masq,
		StatsSecret:   statsSecret,
		Users:         users,
		BandwidthUp:   bwUp,
		BandwidthDown: bwDown,
	}

	desiredVer, _ := r.db.GetSetting(models.SettingDesiredNodeVersion)

	desired := DesiredNodeConfig{
		ServerConfig:        serverCfg,
		CascadeClient:       cascadeClient,
		Run:                 runNode,
		DesiredPanelVersion: desiredVer,
	}
	desired.Version = configVersion(serverCfg, cascadeClient)
	return desired
}

// buildCascadeClientForNode1 возвращает CascadeClientConfig для node1→node2.
// cascadeTarget — конкретное имя node2; если пусто — используется первая доступная.
func (r *Registry) buildCascadeClientForNode1(cascadeTarget, cascadeUser, cascadePass, obfs string) *CascadeClientConfig {
	var node2 models.Server
	var err error
	if cascadeTarget != "" {
		err = r.db.Where("name = ? AND role = ?", cascadeTarget, models.RoleNode2).First(&node2).Error
	} else {
		err = r.db.Where("role = ?", models.RoleNode2).First(&node2).Error
	}
	if err != nil {
		return nil // node2 не найдена
	}
	if node2.CertSHA256 == "" {
		return nil // ещё нет сертификата
	}
	return &CascadeClientConfig{
		ServerAddr:   fmt.Sprintf("%s:%d", node2.PublicIP, node2.Hy2Port),
		UserName:     cascadeUser,
		Password:     cascadePass,
		ObfsPassword: obfs,
		PinSHA256:    node2.CertSHA256,
	}
}

// ensureCascadeCredentials возвращает cascade_user/cascade_password, генерируя их при первом вызове.
func (r *Registry) ensureCascadeCredentials() (user, pass string) {
	user, _ = r.db.GetSetting(models.SettingCascadeUser)
	pass, _ = r.db.GetSetting(models.SettingCascadePassword)
	if user != "" {
		return
	}
	user = "cascade-" + randomHex(6)
	pass = randomHex(24)
	_ = r.db.SetSetting(models.SettingCascadeUser, user)
	_ = r.db.SetSetting(models.SettingCascadePassword, pass)
	return
}

// BuildDesiredConfig публичная обёртка над buildDesiredConfig — используется main_node1.
func (r *Registry) BuildDesiredConfig(role, serverName string) DesiredNodeConfig {
	return r.buildDesiredConfig(role, serverName)
}

// GetNodeLogs возвращает последние логи ноды по её имени.
func (r *Registry) GetNodeLogs(name string) []string {
	r.logsMu.RLock()
	defer r.logsMu.RUnlock()
	src := r.nodeLogs[name]
	if len(src) == 0 {
		return []string{}
	}
	out := make([]string, len(src))
	copy(out, src)
	return out
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// configVersion вычисляет детерминированный хэш конфига — меняется только при изменении контента.
func configVersion(cfg NodeServerConfig, cascade *CascadeClientConfig) int64 {
	var sb strings.Builder
	keys := make([]string, 0, len(cfg.Users))
	for k := range cfg.Users {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(k + "=" + cfg.Users[k] + ";")
	}
	sb.WriteString(cfg.ObfsPassword + "|" + cfg.MasqueradeURL + "|" + cfg.StatsSecret)
	if cascade != nil {
		sb.WriteString("|cascade:" + cascade.ServerAddr + ":" + cascade.UserName)
	}
	h := sha256.Sum256([]byte(sb.String()))
	return int64(binary.LittleEndian.Uint64(h[:8]))
}

// Package hysteria управляет дочерним процессом hysteria2: скачивание,
// генерация конфига, запуск/стоп, опрос трафика.
package hysteria

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"hysteria2-web/internal/db"
	"hysteria2-web/internal/logbuf"
	"hysteria2-web/internal/models"
)

// NodeCascadeClient — параметры для запуска hysteria2 client (каскад node1→node2).
type NodeCascadeClient struct {
	ServerAddr   string // host:port узла node2
	UserName     string
	Password     string
	ObfsPassword string
	PinSHA256    string
}

// NodeConfig — конфиг, который main передаёт ноде (без обращения к БД).
type NodeConfig struct {
	Users         map[string]string
	ObfsPassword  string
	MasqueradeURL string
	StatsSecret   string
	BandwidthUp   string
	BandwidthDown string
	Run           bool
	Cascade       *NodeCascadeClient // не nil только для node1
}

const (
	hysteriaVersion = "v2.9.3"
	hysteriaTag     = "app%2Fv2.9.3" // URL-encoded тег релиза на GitHub
)

// Status — текущее состояние hysteria2.
type Status struct {
	Installed bool
	Running   bool
	Version   string
	Port      int
}

// Config — пользовательски-редактируемые параметры hysteria2.
type Config struct {
	Port          int
	ObfsPassword  string
	MasqueradeURL string
	CertSHA256    string
	BandwidthUp   string
	BandwidthDown string
	SNI           string
}

// effectivePort возвращает порт из БД (если задан) или из конфига панели.
func (m *Manager) effectivePort() int {
	if m.db != nil {
		if v, _ := m.db.GetSetting(models.SettingHy2Port); v != "" {
			var p int
			if _, err := fmt.Sscanf(v, "%d", &p); err == nil && p > 0 {
				return p
			}
		}
	}
	return m.port
}

// Manager управляет жизненным циклом hysteria2 как дочернего процесса.
type Manager struct {
	dataDir        string
	port           int
	db             *db.DB
	logBuf         *logbuf.Buffer
	mu             sync.Mutex
	srv            *supervisor   // server-процесс
	clientSrv      *supervisor   // client-процесс (каскад, только на node1)
	currentCascade *outboundCfg // сохраняется после ApplyNodeConfig, используется при Start/Reload
}

func New(dataDir string, port int, d *db.DB) *Manager {
	return &Manager{dataDir: dataDir, port: port, db: d, logBuf: logbuf.New()}
}

// LogBuf возвращает буфер логов hysteria2 (stdout/stderr дочернего процесса).
func (m *Manager) LogBuf() *logbuf.Buffer { return m.logBuf }

// Install скачивает pinned-бинарь hysteria2, если ещё не установлен.
func (m *Manager) Install(ctx context.Context) error {
	if m.isBinaryInstalled() {
		return nil
	}
	return m.downloadBinary(ctx)
}

// Start генерирует конфиг и запускает hysteria2 под супервизором.
func (m *Manager) Start() error {
	if !m.isBinaryInstalled() {
		return fmt.Errorf("hysteria2 не установлен, нажмите «Установить» сначала")
	}

	if m.db != nil {
		var count int64
		m.db.Model(&models.User{}).Where("is_active = ?", true).Count(&count)
		if count == 0 {
			return fmt.Errorf("нет активных пользователей — создайте хотя бы одного в разделе «Пользователи»")
		}
	}

	if !m.certExists() {
		pin, err := m.generateCert()
		if err != nil {
			return fmt.Errorf("генерация сертификата: %w", err)
		}
		_ = m.db.Model(&models.Server{}).Where("name = ?", "main").Update("cert_sha256", pin).Error
	}

	if err := m.ensureDefaults(); err != nil {
		return err
	}

	if err := m.buildServerYAML(nil); err != nil {
		return fmt.Errorf("генерация server.yaml: %w", err)
	}

	m.mu.Lock()
	if m.srv == nil {
		m.srv = newSupervisor(m.binaryPath(), m.serverYAMLPath(), m.logBuf)
	}
	m.mu.Unlock()

	m.srv.start()
	return m.db.SetSetting(models.SettingHy2Running, "1")
}

// Stop останавливает hysteria2.
func (m *Manager) Stop() error {
	m.mu.Lock()
	s := m.srv
	m.mu.Unlock()
	if s != nil {
		s.stop()
	}
	return m.db.SetSetting(models.SettingHy2Running, "0")
}

// IsRunning возвращает true, если супервизор активен.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.srv != nil && m.srv.running()
}

// Status возвращает текущее состояние.
func (m *Manager) Status() Status {
	return Status{
		Installed: m.isBinaryInstalled(),
		Running:   m.IsRunning(),
		Version:   hysteriaVersion,
		Port:      m.port,
	}
}

// GetConfig возвращает текущий конфиг из БД.
func (m *Manager) GetConfig() Config {
	obfs, _ := m.db.GetSetting(models.SettingObfsPassword)
	masq, _ := m.db.GetSettingOrDefault(models.SettingMasqueradeURL, "https://news.ycombinator.com/")
	bwUp, _ := m.db.GetSetting(models.SettingBandwidthUp)
	bwDown, _ := m.db.GetSetting(models.SettingBandwidthDown)
	sni, _ := m.db.GetSetting(models.SettingSNI)
	var server models.Server
	_ = m.db.Where("name = ?", "main").First(&server).Error
	return Config{
		Port:          m.effectivePort(),
		ObfsPassword:  obfs,
		MasqueradeURL: masq,
		CertSHA256:    server.CertSHA256,
		BandwidthUp:   bwUp,
		BandwidthDown: bwDown,
		SNI:           sni,
	}
}

// UpdateConfig сохраняет параметры конфига в БД (применяется после ReloadConfig).
func (m *Manager) UpdateConfig(port int, obfsPassword, masqURL, bwUp, bwDown, sni string) error {
	if port > 0 {
		if err := m.db.SetSetting(models.SettingHy2Port, fmt.Sprintf("%d", port)); err != nil {
			return err
		}
	}
	if obfsPassword != "" {
		if err := m.db.SetSetting(models.SettingObfsPassword, obfsPassword); err != nil {
			return err
		}
	}
	if masqURL != "" {
		if err := m.db.SetSetting(models.SettingMasqueradeURL, masqURL); err != nil {
			return err
		}
	}
	if err := m.db.SetSetting(models.SettingBandwidthUp, bwUp); err != nil {
		return err
	}
	if err := m.db.SetSetting(models.SettingBandwidthDown, bwDown); err != nil {
		return err
	}
	if err := m.db.SetSetting(models.SettingSNI, sni); err != nil {
		return err
	}
	return nil
}

// ReloadConfig перезаписывает server.yaml и перезапускает процесс.
// Для нод (m.db == nil) ничего не делает — их конфиг управляется main-ом.
func (m *Manager) ReloadConfig() error {
	if m.db == nil {
		return nil
	}
	if m.IsRunning() {
		if err := m.Stop(); err != nil {
			return err
		}
		return m.Start()
	}
	return m.buildServerYAML(nil)
}

// RegenerateCert создаёт новый self-signed сертификат.
func (m *Manager) RegenerateCert() (string, error) {
	pin, err := m.generateCert()
	if err != nil {
		return "", err
	}
	_ = m.db.Model(&models.Server{}).Where("name = ?", "main").Update("cert_sha256", pin).Error
	return pin, nil
}

// ensureDefaults создаёт настройки при первом запуске hysteria.
func (m *Manager) ensureDefaults() error {
	if obfs, _ := m.db.GetSetting(models.SettingObfsPassword); obfs == "" {
		if err := m.db.SetSetting(models.SettingObfsPassword, randomAlnum(30)); err != nil {
			return err
		}
	}
	if secret, _ := m.db.GetSetting(models.SettingStatsSecret); secret == "" {
		if err := m.db.SetSetting(models.SettingStatsSecret, randomHex(16)); err != nil {
			return err
		}
	}
	if masq, _ := m.db.GetSetting(models.SettingMasqueradeURL); masq == "" {
		if err := m.db.SetSetting(models.SettingMasqueradeURL, "https://news.ycombinator.com/"); err != nil {
			return err
		}
	}
	return nil
}

// CertSHA256 читает SHA-256 пин из cert.pem напрямую (без БД — нужно для нод).
func (m *Manager) CertSHA256() string {
	data, err := os.ReadFile(filepath.Join(m.dataDir, "cert.pem"))
	if err != nil {
		return ""
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return ""
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(cert.Raw)
	parts := make([]string, 32)
	for i, b := range hash {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}

// ApplyNodeConfig применяет конфиг от main на ноде (без БД).
// Скачивает бинарь и генерирует сертификат при первом запуске.
// Run=false означает "не запускай сервер", но каскадный клиент всё равно стартует если задан.
func (m *Manager) ApplyNodeConfig(ctx context.Context, nc NodeConfig) error {
	// Для любого запуска нужен бинарь и сертификат
	if nc.Run || nc.Cascade != nil {
		if !m.isBinaryInstalled() {
			if err := m.downloadBinary(ctx); err != nil {
				return fmt.Errorf("скачивание hysteria2: %w", err)
			}
		}
		if !m.certExists() {
			if _, err := m.generateCert(); err != nil {
				return fmt.Errorf("генерация сертификата: %w", err)
			}
		}
	}

	// Сохраняем каскад в менеджере — Start() и ReloadConfig() будут его использовать.
	m.mu.Lock()
	if nc.Cascade != nil {
		m.currentCascade = &outboundCfg{Addr: "127.0.0.1:1080"}
	} else {
		m.currentCascade = nil
	}
	m.mu.Unlock()

	// ── Hysteria2 сервер ─────────────────────────────────────────────────────
	if nc.Run {
		var ob *outboundCfg
		if nc.Cascade != nil {
			ob = &outboundCfg{Addr: "127.0.0.1:1080"}
		}
		if err := m.buildServerYAMLFromConfig(nc.Users, nc.ObfsPassword, nc.MasqueradeURL, nc.StatsSecret, nc.BandwidthUp, nc.BandwidthDown, ob); err != nil {
			return fmt.Errorf("генерация server.yaml: %w", err)
		}
		m.mu.Lock()
		if m.srv != nil {
			m.srv.stop()
		}
		m.srv = newSupervisor(m.binaryPath(), m.serverYAMLPath(), m.logBuf)
		m.mu.Unlock()
		m.srv.start()
	} else {
		m.mu.Lock()
		s := m.srv
		m.srv = nil
		m.mu.Unlock()
		if s != nil {
			s.stop()
		}
	}

	// ── Каскадный клиент (node1 → node2) ─────────────────────────────────────
	// Запускается независимо от Run: клиент нужен даже когда нет юзеров.
	if nc.Cascade != nil {
		if err := m.buildClientYAML(*nc.Cascade); err != nil {
			return fmt.Errorf("генерация client.yaml: %w", err)
		}
		m.mu.Lock()
		if m.clientSrv != nil {
			m.clientSrv.stop()
		}
		m.clientSrv = newClientSupervisor(m.binaryPath(), m.clientYAMLPath(), m.logBuf)
		m.mu.Unlock()
		m.clientSrv.start()
	} else {
		m.mu.Lock()
		cs := m.clientSrv
		m.clientSrv = nil
		m.mu.Unlock()
		if cs != nil {
			cs.stop()
		}
	}
	return nil
}

func (m *Manager) serverYAMLPath() string {
	return filepath.Join(m.dataDir, "server.yaml")
}

func (m *Manager) clientYAMLPath() string {
	return filepath.Join(m.dataDir, "client.yaml")
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

const alnumAlphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomAlnum(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alnumAlphabet[int(b[i])%len(alnumAlphabet)]
	}
	return string(b)
}

package hysteria

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"hysteria2-web/internal/models"
)

// outboundCfg задаёт исходящий socks5-прокси (для каскада node1→node2).
type outboundCfg struct {
	Addr string // 127.0.0.1:1080
}

// buildClientYAML записывает client.yaml для подключения к node2 (каскад).
func (m *Manager) buildClientYAML(cc NodeCascadeClient) error {
	cfg := clientYAML{
		Server: cc.ServerAddr,
		Auth:   cc.UserName + ":" + cc.Password, // hysteria2 client: "user:pass" строка
		Obfs: &obfsYAML{
			Type:       "salamander",
			Salamander: salamaYAML{Password: cc.ObfsPassword},
		},
		TLS: clientTLSYAML{
			PinSHA256: cc.PinSHA256,
			Insecure:  true,
		},
		Socks5: socks5ListenYAML{Listen: "127.0.0.1:1080"},
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(m.clientYAMLPath(), data, 0o600)
}

// ── client.yaml структуры ───────────────────────────────────────────────────

type clientYAML struct {
	Server string           `yaml:"server"`
	Auth   string           `yaml:"auth"` // "user:password"
	Obfs   *obfsYAML        `yaml:"obfs,omitempty"`
	TLS    clientTLSYAML    `yaml:"tls"`
	Socks5 socks5ListenYAML `yaml:"socks5"`
}

type clientTLSYAML struct {
	PinSHA256 string `yaml:"pinSHA256"`
	Insecure  bool   `yaml:"insecure"`
}

type socks5ListenYAML struct {
	Listen string `yaml:"listen"`
}

// buildServerYAMLFromConfig генерирует server.yaml из переданных параметров (без DB).
func (m *Manager) buildServerYAMLFromConfig(users map[string]string, obfsPassword, masqURL, statsSecret string, outbound *outboundCfg) error {
	cfg := buildCfg(m.port,
		filepath.Join(m.dataDir, "cert.pem"),
		filepath.Join(m.dataDir, "key.pem"),
		obfsPassword, masqURL, statsSecret, users, outbound)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(m.serverYAMLPath(), data, 0o600)
}

// buildServerYAML загружает активных пользователей и настройки из БД,
// генерирует server.yaml в dataDir. outbound != nil только для node1 (каскад).
func (m *Manager) buildServerYAML(outbound *outboundCfg) error {
	// Загружаем активных пользователей
	var users []models.User
	_ = m.db.Where("is_active = ?", true).Find(&users).Error
	userMap := make(map[string]string, len(users))
	for _, u := range users {
		userMap[u.Name] = u.Password
	}

	obfs, _ := m.db.GetSetting(models.SettingObfsPassword)
	masq, _ := m.db.GetSettingOrDefault(models.SettingMasqueradeURL, "https://news.ycombinator.com/")
	secret, _ := m.db.GetSetting(models.SettingStatsSecret)

	cfg := buildCfg(m.port, filepath.Join(m.dataDir, "cert.pem"), filepath.Join(m.dataDir, "key.pem"),
		obfs, masq, secret, userMap, outbound)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(m.serverYAMLPath(), data, 0o600)
}

// ── YAML-структуры конфига hysteria2 server ─────────────────────────────────

type serverYAML struct {
	Listen       string         `yaml:"listen"`
	TLS          tlsYAML        `yaml:"tls"`
	Obfs         *obfsYAML      `yaml:"obfs,omitempty"`
	Auth         authYAML       `yaml:"auth"`
	Masquerade   masqYAML       `yaml:"masquerade"`
	TrafficStats statsYAML      `yaml:"trafficStats"`
	Outbounds    []outboundYAML `yaml:"outbounds,omitempty"`
	ACL          *aclYAML       `yaml:"acl,omitempty"`
}

type tlsYAML struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

type obfsYAML struct {
	Type       string     `yaml:"type"`
	Salamander salamaYAML `yaml:"salamander"`
}

type salamaYAML struct {
	Password string `yaml:"password"`
}

type authYAML struct {
	Type     string            `yaml:"type"`
	UserPass map[string]string `yaml:"userpass"`
}

type masqYAML struct {
	Type  string    `yaml:"type"`
	Proxy masqProxy `yaml:"proxy"`
}

type masqProxy struct {
	URL         string `yaml:"url"`
	RewriteHost bool   `yaml:"rewriteHost"`
}

type statsYAML struct {
	Listen string `yaml:"listen"`
	Secret string `yaml:"secret"`
}

type outboundYAML struct {
	Name   string      `yaml:"name"`
	Type   string      `yaml:"type"`
	Socks5 *socks5YAML `yaml:"socks5,omitempty"`
}

type socks5YAML struct {
	Addr string `yaml:"addr"`
}

type aclYAML struct {
	Inline []string `yaml:"inline"`
}

func buildCfg(port int, certPath, keyPath, obfs, masqURL, statsSecret string,
	users map[string]string, out *outboundCfg) serverYAML {

	cfg := serverYAML{
		Listen: fmt.Sprintf(":%d", port),
		TLS:    tlsYAML{Cert: certPath, Key: keyPath},
		Auth: authYAML{
			Type:     "userpass",
			UserPass: users,
		},
		Masquerade: masqYAML{
			Type:  "proxy",
			Proxy: masqProxy{URL: masqURL, RewriteHost: true},
		},
		TrafficStats: statsYAML{
			Listen: "127.0.0.1:9999",
			Secret: statsSecret,
		},
	}

	if obfs != "" {
		cfg.Obfs = &obfsYAML{
			Type:       "salamander",
			Salamander: salamaYAML{Password: obfs},
		}
	}

	if out != nil {
		// Когда задан только один outbound, hysteria2 использует его как дефолтный
		// для всего трафика без ACL-правил.
		cfg.Outbounds = []outboundYAML{{
			Name:   "cascade",
			Type:   "socks5",
			Socks5: &socks5YAML{Addr: out.Addr},
		}}
	}

	return cfg
}

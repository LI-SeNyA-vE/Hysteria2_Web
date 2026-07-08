// Package config загружает конфигурацию панели из YAML + env-оверрайдов.
package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"

	"hysteria2-web/internal/models"
)

// Config — конфигурация одного экземпляра панели.
type Config struct {
	Role               string `yaml:"role"`
	HTTPAddr           string `yaml:"httpAddr"`
	DataDir            string `yaml:"dataDir"`
	PublicIP           string `yaml:"publicIp"`
	Dev                bool   `yaml:"dev"`
	BootstrapNodeToken string `yaml:"bootstrapNodeToken"`
	// CascadeTarget — имя node2 для каскадирования (только для роли node1).
	// Если пусто — используется первая доступная node2.
	CascadeTarget string `yaml:"cascadeTarget"`

	DB struct {
		DSN string `yaml:"dsn"`
	} `yaml:"db"`

	Main struct {
		URL   string `yaml:"url"`
		Token string `yaml:"token"`
	} `yaml:"main"`

	Hy2 struct {
		Port int `yaml:"port"`
	} `yaml:"hy2"`
}

// Load читает YAML (если path не пуст и файл существует), затем накладывает
// env-оверрайды и валидирует по роли. Работа без файла обязательна (Docker).
func Load(path string) (*Config, error) {
	c := &Config{
		Role:     models.RoleMainNode1,
		HTTPAddr: ":8080",
		DataDir:  "./data",
		PublicIP: "127.0.0.1",
	}
	c.Hy2.Port = 443

	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, c); err != nil {
				return nil, fmt.Errorf("парсинг %s: %w", path, err)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("чтение %s: %w", path, err)
		}
	}

	c.applyEnv()

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) applyEnv() {
	if v := os.Getenv("PANEL_ROLE"); v != "" {
		c.Role = v
	}
	if v := os.Getenv("PANEL_HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := os.Getenv("PANEL_DATA_DIR"); v != "" {
		c.DataDir = v
	}
	if v := os.Getenv("PANEL_PUBLIC_IP"); v != "" {
		c.PublicIP = v
	}
	if v := os.Getenv("PANEL_DB_DSN"); v != "" {
		c.DB.DSN = v
	}
	if v := os.Getenv("PANEL_MAIN_URL"); v != "" {
		c.Main.URL = v
	}
	if v := os.Getenv("PANEL_NODE_TOKEN"); v != "" {
		c.Main.Token = v
	}
	if v := os.Getenv("PANEL_HY2_PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Hy2.Port = n
		}
	}
	if v := os.Getenv("PANEL_DEV"); v == "1" || v == "true" {
		c.Dev = true
	}
	if v := os.Getenv("PANEL_BOOTSTRAP_NODE_TOKEN"); v != "" {
		c.BootstrapNodeToken = v
	}
	if v := os.Getenv("PANEL_CASCADE_TARGET"); v != "" {
		c.CascadeTarget = v
	}
}

func (c *Config) validate() error {
	switch c.Role {
	case models.RoleMain, models.RoleMainNode1:
		// DSN опционален: пустой → SQLite в dataDir/panel.db
	case models.RoleNode1, models.RoleNode2:
		if c.Main.URL == "" || c.Main.Token == "" {
			return fmt.Errorf("роль %s требует main.url и main.token", c.Role)
		}
	default:
		return fmt.Errorf("неизвестная роль %q", c.Role)
	}
	return nil
}

// HasDatabase сообщает, поднимает ли этот экземпляр локальную БД.
func (c *Config) HasDatabase() bool {
	return c.Role == models.RoleMain || c.Role == models.RoleMainNode1
}

// RunsHysteria сообщает, управляет ли экземпляр локальным hysteria-процессом.
func (c *Config) RunsHysteria() bool {
	return c.Role == models.RoleNode1 || c.Role == models.RoleNode2 || c.Role == models.RoleMainNode1
}

// IsNode сообщает, что экземпляр работает в роли ноды (регистрируется на main).
func (c *Config) IsNode() bool {
	return c.Role == models.RoleNode1 || c.Role == models.RoleNode2
}

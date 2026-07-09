package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"hysteria2-web/internal/config"
	"hysteria2-web/internal/hysteria"
)

// NodeAgent работает на ноде: регистрируется на main и держит heartbeat-цикл.
type NodeAgent struct {
	cfg            *config.Config
	mgr            *hysteria.Manager
	lastVersion    int64
	lastCertSHA256 string
}

func NewNodeAgent(cfg *config.Config, mgr *hysteria.Manager) *NodeAgent {
	return &NodeAgent{cfg: cfg, mgr: mgr}
}

// Run запускает регистрацию и бесконечный heartbeat (отменяется через ctx).
func (a *NodeAgent) Run(ctx context.Context) {
	// Первая регистрация — повторяем до успеха
	for {
		if err := a.register(ctx); err != nil {
			log.Printf("nodeagent: регистрация: %v, повтор через 5с", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		break
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.heartbeat(ctx); err != nil {
				log.Printf("nodeagent: heartbeat: %v", err)
			}
		}
	}
}

func (a *NodeAgent) register(ctx context.Context) error {
	// Имя ноды: hostname если доступен, иначе роль (для Docker/dev)
	name := nodeName(a.cfg)
	req := RegisterRequest{
		Role:          a.cfg.Role,
		Name:          name,
		PublicIP:      a.cfg.PublicIP,
		Hy2Port:       a.cfg.Hy2.Port,
		Hy2Version:    "v2.9.3",
		CertSHA256:    a.mgr.CertSHA256(),
		CascadeTarget: a.cfg.CascadeTarget,
	}
	desired, err := a.post(ctx, "/api/node/register", req)
	if err != nil {
		return err
	}
	return a.applyIfChanged(ctx, desired)
}

func (a *NodeAgent) heartbeat(ctx context.Context) error {
	req := HeartbeatRequest{
		Name:       nodeName(a.cfg),
		Running:    a.mgr.IsRunning(),
		CertSHA256: a.mgr.CertSHA256(),
		Usage:      map[string]UsageStat{}, // трафик на ноде — TODO: pollOnce без DB
		Logs:       a.mgr.LogBuf().Lines(),
	}
	desired, err := a.post(ctx, "/api/node/heartbeat", req)
	if err != nil {
		return err
	}
	return a.applyIfChanged(ctx, desired)
}

func (a *NodeAgent) applyIfChanged(ctx context.Context, desired DesiredNodeConfig) error {
	if desired.Version == a.lastVersion {
		return nil // конфиг не изменился
	}
	log.Printf("nodeagent: применяю конфиг версии %d", desired.Version)

	nc := hysteria.NodeConfig{
		Users:         desired.ServerConfig.Users,
		ObfsPassword:  desired.ServerConfig.ObfsPassword,
		MasqueradeURL: desired.ServerConfig.MasqueradeURL,
		StatsSecret:   desired.ServerConfig.StatsSecret,
		BandwidthUp:   desired.ServerConfig.BandwidthUp,
		BandwidthDown: desired.ServerConfig.BandwidthDown,
		Run:           desired.Run,
	}
	if desired.CascadeClient != nil {
		nc.Cascade = &hysteria.NodeCascadeClient{
			ServerAddr:   desired.CascadeClient.ServerAddr,
			UserName:     desired.CascadeClient.UserName,
			Password:     desired.CascadeClient.Password,
			ObfsPassword: desired.CascadeClient.ObfsPassword,
			PinSHA256:    desired.CascadeClient.PinSHA256,
		}
	}
	if err := a.mgr.ApplyNodeConfig(ctx, nc); err != nil {
		return fmt.Errorf("применение конфига: %w", err)
	}

	a.lastVersion = desired.Version
	// Обновляем certSHA256 после возможной генерации
	a.lastCertSHA256 = a.mgr.CertSHA256()
	return nil
}

// nodeName возвращает hostname сервера, либо роль как запасной вариант (Docker/dev).
func nodeName(cfg *config.Config) string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return cfg.Role
}

func (a *NodeAgent) post(ctx context.Context, path string, body any) (DesiredNodeConfig, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return DesiredNodeConfig{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		a.cfg.Main.URL+path, bytes.NewReader(data))
	if err != nil {
		return DesiredNodeConfig{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Node-Token", a.cfg.Main.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return DesiredNodeConfig{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return DesiredNodeConfig{}, fmt.Errorf("main вернул %d для %s", resp.StatusCode, path)
	}
	var desired DesiredNodeConfig
	if err := json.NewDecoder(resp.Body).Decode(&desired); err != nil {
		return DesiredNodeConfig{}, err
	}
	return desired, nil
}

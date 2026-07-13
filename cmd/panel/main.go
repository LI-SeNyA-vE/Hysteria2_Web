package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/gorm"

	"io"

	"hysteria2-web/internal/api"
	"hysteria2-web/internal/auth"
	"hysteria2-web/internal/cluster"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/db"
	"hysteria2-web/internal/hysteria"
	"hysteria2-web/internal/logbuf"
	"hysteria2-web/internal/models"
)

func main() {
	cfgPath := flag.String("config", "panel.yaml", "путь к файлу конфигурации")
	flag.Parse()

	// Перехватываем log.Default() в кольцевой буфер (параллельно с os.Stderr).
	panelBuf := logbuf.New()
	log.SetOutput(io.MultiWriter(os.Stderr, panelBuf))

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("конфигурация: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		d   *db.DB
		a   *auth.Auth
		mgr *hysteria.Manager
		reg *cluster.Registry
	)

	if cfg.HasDatabase() {
		d, err = db.Open(cfg.DB.DSN, cfg.DataDir)
		if err != nil {
			log.Fatalf("база данных: %v", err)
		}
		var firstRunPassword string
		a, firstRunPassword, err = auth.Bootstrap(d, cfg.BootstrapNodeToken)
		if err != nil {
			log.Fatalf("auth bootstrap: %v", err)
		}
		if firstRunPassword != "" {
			printFirstRunBanner(cfg.HTTPAddr, firstRunPassword)
		}
		reg = cluster.NewRegistry(d)
	} else {
		a = auth.NewNodeAuth(cfg.Main.Token)
		log.Printf("Нода запущена (роль: %s, main: %s)", cfg.Role, cfg.Main.URL)
	}

	if cfg.RunsHysteria() {
		mgr = hysteria.New(cfg.DataDir, cfg.Hy2.Port, d)
		if d != nil {
			// Авто-старт: запускаем если не было явной остановки пользователем ("0").
			// Пустая строка = первый запуск / переустановка → тоже пробуем стартовать.
			if stopped, _ := d.GetSetting(models.SettingHy2Running); stopped != "0" {
				if err := mgr.Start(); err != nil {
					log.Printf("авто-старт hysteria2 пропущен: %v", err)
				}
			}
			go mgr.PollTrafficForever(ctx)
		}
	}

	if d != nil {
		registerSelf(d, cfg, mgr)
	}

	// Для нод (без БД): NodeAgent регистрируется на main и применяет DesiredNodeConfig.
	if cfg.IsNode() && mgr != nil {
		agent := cluster.NewNodeAgent(cfg, mgr)
		go agent.Run(ctx)
	}

	// Для main_node1: self-apply каскадного конфига (NodeAgent не запускается т.к. нет remote main).
	if cfg.Role == models.RoleMainNode1 && mgr != nil && reg != nil {
		go func() {
			var lastVer int64
			apply := func() {
				desired := reg.BuildDesiredConfig(models.RoleMainNode1, "main")
				if desired.Version == lastVer {
					return
				}
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
				if err := mgr.ApplyNodeConfig(ctx, nc); err != nil {
					log.Printf("main_node1: применение каскадного конфига: %v", err)
					return
				}
				lastVer = desired.Version
			}
			t := time.NewTicker(15 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					apply()
				}
			}
		}()
	}

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      api.NewServer(a, d, cfg.Dev, mgr, reg, panelBuf).Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("HTTP сервер запущен на %s (роль: %s)", cfg.HTTPAddr, cfg.Role)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP сервер: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Завершение работы...")

	if mgr != nil && mgr.IsRunning() {
		_ = mgr.Stop()
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Printf("принудительное завершение: %v", err)
	}
}

func registerSelf(d *db.DB, cfg *config.Config, mgr *hysteria.Manager) {
	now := time.Now()
	hy2Ver := ""
	if mgr != nil {
		hy2Ver = mgr.Status().Version
	}
	var s models.Server
	err := d.Where("name = ?", "main").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s = models.Server{
			Name:       "main",
			Role:       cfg.Role,
			PublicIP:   cfg.PublicIP,
			PanelURL:   "http://localhost" + cfg.HTTPAddr,
			Hy2Port:    cfg.Hy2.Port,
			Hy2Version: hy2Ver,
			LastSeenAt: &now,
		}
		if err := d.Create(&s).Error; err != nil {
			log.Printf("предупреждение: не удалось создать запись сервера: %v", err)
		}
		return
	}
	updates := map[string]any{
		"role":         cfg.Role,
		"public_ip":    cfg.PublicIP,
		"panel_url":    "http://localhost" + cfg.HTTPAddr,
		"hy2_port":     cfg.Hy2.Port,
		"last_seen_at": &now,
	}
	if hy2Ver != "" {
		updates["hy2_version"] = hy2Ver
	}
	d.Model(&s).Updates(updates)
}

func printFirstRunBanner(addr, password string) {
	host := addr
	if host == "" || host[0] == ':' {
		host = "localhost" + addr
	}
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║      HYSTERIA2 PANEL — ПЕРВЫЙ ЗАПУСК     ║")
	fmt.Println("╠══════════════════════════════════════════╣")
	fmt.Printf("║  URL:      http://%s%-*s║\n", host, 22-len(host), "")
	fmt.Printf("║  Пароль:   %-30s║\n", password)
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("  Сохраните пароль — он показывается только один раз.")
	fmt.Println()
}

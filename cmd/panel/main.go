package main

import (
	"flag"
	"fmt"
	"os"

	"hysteria2-web/internal/cli"
	"hysteria2-web/internal/config"
	"hysteria2-web/internal/version"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-h" || args[0] == "-help" || args[0] == "--help" || args[0] == "help") {
		printUsage()
		os.Exit(0)
	}

	command, flagArgs := parseCommand(args)

	fs := flag.NewFlagSet("panel", flag.ExitOnError)
	configPath := fs.String("config", config.DefaultPath, "путь к файлу конфигурации")
	fs.Usage = printUsage
	if err := fs.Parse(flagArgs); err != nil {
		os.Exit(1)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка конфигурации: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "serve":
		cli.RunServe(cfg)
	default:
		cli.RunInteractive(cfg, *configPath)
	}
}

func parseCommand(args []string) (command string, rest []string) {
	if len(args) > 0 && args[0] == "serve" {
		return "serve", args[1:]
	}
	return "admin", args
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Hysteria2 VPN Panel v%s

Использование:
  panel [-config panel.json]           меню администратора
  panel serve [-config panel.json]     служба (HTTP подписок + sync)

Примеры:
  panel serve                          запуск службы
  panel                                открыть меню
  systemctl start hysteria2-panel      автозапуск службы

`, version.Version)
}

package main

import (
	"flag"
	"fmt"
	"os"

	"hysteria2-web/internal/cli"
	"hysteria2-web/internal/config"
)

func main() {
	configPath := flag.String("config", config.DefaultPath, "путь к файлу конфигурации")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка конфигурации: %v\n", err)
		os.Exit(1)
	}

	cli.RunInteractive(cfg, *configPath)
}

package cli

import (
	"bufio"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"hysteria2-web/internal/config"
)

var ErrReloadPanel = errors.New("reload panel")

func interactiveSettings(reader *bufio.Reader, cfg config.Config) error {
	draft := cfg

	for {
		printSettingsMenu(draft)

		choice, err := readLine(reader, "Выберите: ")
		if isInputCancelled(err) {
			return err
		}

		switch choice {
		case "0", "q", "back":
			return nil
		case "s", "save":
			prepared, err := config.Prepare(draft)
			if err != nil {
				return err
			}
			if err := config.Save(config.ConfigPath(), prepared); err != nil {
				return err
			}
			fmt.Println("Настройки сохранены.")
			return ErrReloadPanel
		case "1":
			draft.DBPath, err = readStringWithDefault(reader, "db_path", draft.DBPath)
		case "2":
			draft.LogPath, err = readStringWithDefault(reader, "log_path", draft.LogPath)
		case "3":
			draft.HTTPAddr, err = readStringWithDefault(reader, "http_addr", draft.HTTPAddr)
		case "4":
			draft.SyncInterval, err = readDurationWithDefault(reader, "sync_interval", draft.SyncInterval)
		case "5":
			draft.SubDomain, err = readOptionalClearable(reader, "sub_domain", draft.SubDomain)
		case "6":
			draft.SubPath, err = readStringWithDefault(reader, "sub_path", draft.SubPath)
		default:
			fmt.Println("Неверный выбор.")
			continue
		}

		if isInputCancelled(err) {
			return err
		}
		if err != nil {
			printFail(err)
		}
	}
}

func printSettingsMenu(cfg config.Config) {
	abs, _ := filepath.Abs(config.ConfigPath())
	fmt.Printf("Файл: %s\n\n", abs)
	fmt.Printf("  1. db_path         %s\n", cfg.DBPath)
	fmt.Printf("  2. log_path        %s\n", cfg.LogPath)
	fmt.Printf("  3. http_addr       %s\n", cfg.HTTPAddr)
	fmt.Printf("  4. sync_interval   %s\n", cfg.SyncInterval)
	fmt.Printf("  5. sub_domain      %s\n", displayOptional(cfg.SubDomain, "локальный из http_addr"))
	fmt.Printf("  6. sub_path        %s\n", cfg.SubscriptionPath())
	fmt.Println()
	fmt.Printf("  Подписка: %s\n", cfg.SubscriptionURL("{SubToken}"))
	fmt.Println()
	fmt.Println("  s. Сохранить и перезагрузить")
	fmt.Println("  0. Назад без сохранения")
	fmt.Println()
}

func displayOptional(value, emptyLabel string) string {
	if value == "" {
		return emptyLabel
	}
	return value
}

func readStringWithDefault(reader *bufio.Reader, label, current string) (string, error) {
	raw, err := readLine(reader, fmt.Sprintf("%s [%s]: ", label, current))
	if isInputCancelled(err) {
		return "", err
	}
	if raw == "" {
		return current, nil
	}
	return raw, nil
}

func readOptionalClearable(reader *bufio.Reader, label, current string) (string, error) {
	hint := displayOptional(current, "пусто")
	raw, err := readLine(reader, fmt.Sprintf("%s [%s, '-' очистить]: ", label, hint))
	if isInputCancelled(err) {
		return "", err
	}
	switch raw {
	case "":
		return current, nil
	case "-":
		return "", nil
	default:
		return raw, nil
	}
}

func readDurationWithDefault(reader *bufio.Reader, label string, current time.Duration) (time.Duration, error) {
	raw, err := readLine(reader, fmt.Sprintf("%s [%s]: ", label, current))
	if isInputCancelled(err) {
		return 0, err
	}
	if raw == "" {
		return current, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("неверный интервал: %w", err)
	}
	return d, nil
}

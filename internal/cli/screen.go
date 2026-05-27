package cli

import (
	"fmt"
	"os"

	"hysteria2-web/internal/config"
)

func clearScreen() {
	restoreTerminal()
	fmt.Fprint(os.Stdout, "\033[2J\033[H")
}

func printScreenHeader(title string) {
	termPrintln("========================================")
	termPrintln("         Hysteria2 VPN Panel")
	termPrintln("========================================")
	if title != "" {
		termPrintf("  %s\n", title)
		termPrintln("----------------------------------------")
	}
}

func printMenu(cfg config.Config, serviceOK bool) {
	printScreenHeader("")
	if serviceOK {
		termPrintln("  Служба: работает (HTTP + sync)")
	} else {
		termPrintln("  Служба: не запущена → panel serve")
	}
	termPrintf("  Sync: %s  |  Log: %s\n", cfg.SyncInterval, cfg.LogPath)
	termPrintf("  Sub: %s/%s/{SubToken}\n", cfg.SubscriptionPublicBase(), cfg.SubscriptionPath())
	termPrintln("  (не username! URL — в п. 10)")
	if cfg.UsingLocalSubscriptionURL() {
		termPrintln("  (телефон: sub_domain в п. 11 — http://IP:8787)")
	}
	termPrintln("  Ctrl+C — отмена (в меню — выход)")
	termPrintln("----------------------------------------")
	termPrintln("  1. Список серверов")
	termPrintln("  2. Добавить сервер")
	termPrintln("  3. Удалить сервер")
	termPrintln("  4. Статус сервера")
	termPrintln("  5. Список пользователей")
	termPrintln("  6. Добавить пользователя")
	termPrintln("  7. Kick пользователя")
	termPrintln("  8. URI подключения")
	termPrintln("  9. Синхронизация трафика")
	termPrintln(" 10. QR подписки")
	termPrintln("  11. Настройки")
	termPrintln("  0. Выход")
	termPrintln("========================================")
}

func menuTitle(choice string) string {
	switch choice {
	case "1":
		return "Список серверов"
	case "2":
		return "Добавить сервер"
	case "3":
		return "Удалить сервер"
	case "4":
		return "Статус сервера"
	case "5":
		return "Список пользователей"
	case "6":
		return "Добавить пользователя"
	case "7":
		return "Kick пользователя"
	case "8":
		return "URI подключения"
	case "9":
		return "Синхронизация трафика"
	case "10":
		return "QR подписки"
	case "11":
		return "Настройки"
	default:
		return ""
	}
}

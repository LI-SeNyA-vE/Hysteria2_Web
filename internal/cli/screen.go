package cli

import (
	"fmt"
	"os"

	"hysteria2-web/internal/config"
)

func clearScreen() {
	fmt.Fprint(os.Stdout, "\033[2J\033[H")
}

func printScreenHeader(title string) {
	fmt.Println("========================================")
	fmt.Println("         Hysteria2 VPN Panel")
	fmt.Println("========================================")
	if title != "" {
		fmt.Printf("  %s\n", title)
		fmt.Println("----------------------------------------")
	}
}

func printMenu(cfg config.Config, serviceOK bool) {
	printScreenHeader("")
	if serviceOK {
		fmt.Println("  Служба: работает (HTTP + sync)")
	} else {
		fmt.Println("  Служба: не запущена → panel serve")
	}
	fmt.Printf("  Sync: %s  |  Log: %s\n", cfg.SyncInterval, cfg.LogPath)
	fmt.Printf("  Sub: %s/%s/{SubToken}\n", cfg.SubscriptionPublicBase(), cfg.SubscriptionPath())
	fmt.Println("  (не username! URL — в п. 10)")
	if cfg.UsingLocalSubscriptionURL() {
		fmt.Println("  (телефон: sub_domain в п. 11 — http://IP:8787)")
	}
	fmt.Println("  Ctrl+C — отмена (в меню — выход)")
	fmt.Println("----------------------------------------")
	fmt.Println("  1. Список серверов")
	fmt.Println("  2. Добавить сервер")
	fmt.Println("  3. Удалить сервер")
	fmt.Println("  4. Статус сервера")
	fmt.Println("  5. Список пользователей")
	fmt.Println("  6. Добавить пользователя")
	fmt.Println("  7. Kick пользователя")
	fmt.Println("  8. URI подключения")
	fmt.Println("  9. Синхронизация трафика")
	fmt.Println(" 10. QR подписки")
	fmt.Println(" 11. Настройки")
	fmt.Println("  0. Выход")
	fmt.Println("========================================")
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

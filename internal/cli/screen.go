package cli

import (
	"fmt"
	"os"
	"time"

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

func printMenu(syncInterval time.Duration, logPath, httpAddr string) {
	printScreenHeader("")
	fmt.Printf("  Sync: %s  |  Log: %s\n", syncInterval, logPath)
	fmt.Printf("  Sub: %s/sub/{SubToken}\n", config.SubscriptionPublicBase())
	fmt.Println("  (не username! URL — в п. 10)")
	if config.UsingLocalSubscriptionURL() {
		fmt.Println("  (телефон: export SUB_PUBLIC_URL=http://IP:8080)")
	}
	fmt.Println("  Ctrl+C — прервать ввод")
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
	default:
		return ""
	}
}

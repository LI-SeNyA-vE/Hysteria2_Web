package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"hysteria2-web/internal/app"
	"hysteria2-web/internal/config"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"golang.org/x/term"
)

func printSubscriptionLink(a *app.App, username string) {
	info, err := subscriptionInfoForUser(a, username)
	if err != nil {
		return
	}
	fmt.Printf("Username: %s\n", info.Username)
	fmt.Printf("SubToken: %s\n", info.Token)
	fmt.Printf("Подписка: %s\n", info.URL)
	fmt.Println("(в URL — SubToken, не username)")
}

type subscriptionInfo struct {
	Username string
	Token    string
	URL      string
}

func subscriptionInfoForUser(a *app.App, username string) (subscriptionInfo, error) {
	if _, err := a.BlitzSvc.EnsureSubToken(username); err != nil {
		return subscriptionInfo{}, err
	}
	token, err := a.UserRepo.GetSubTokenByUsername(username)
	if err != nil {
		return subscriptionInfo{}, err
	}
	if token == "" {
		return subscriptionInfo{}, fmt.Errorf("у пользователя %q нет токена подписки", username)
	}
	return subscriptionInfo{
		Username: username,
		Token:    token,
		URL:      config.SubscriptionURL(token),
	}, nil
}

func subscriptionURLForUser(a *app.App, username string) (string, error) {
	info, err := subscriptionInfoForUser(a, username)
	if err != nil {
		return "", err
	}
	return info.URL, nil
}

func interactiveSubscriptionQR(reader *bufio.Reader, a *app.App, _ context.Context) error {
	username, err := readRequired(reader, "Username (a-z, A-Z, 0-9, _)")
	if err != nil {
		return err
	}
	if err := validateUsername(username); err != nil {
		return err
	}

	url, err := subscriptionURLForUser(a, username)
	if err != nil {
		return err
	}
	info, err := subscriptionInfoForUser(a, username)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("Username:  %s\n", info.Username)
	fmt.Printf("SubToken:  %s\n", info.Token)
	fmt.Println()
	fmt.Println("Ссылка подписки (вставьте в Nekobox / Shadowrocket / v2rayNG):")
	fmt.Println(url)
	fmt.Println()
	fmt.Println("⚠ В URL используется SubToken, а не username.")
	subPath := config.SubscriptionPath()
	fmt.Printf("  Неверно: /%s/testuserpanel1\n", subPath)
	fmt.Printf("  Верно:   /%s/%s\n", subPath, info.Token)
	fmt.Println()

	if config.UsingLocalSubscriptionURL() {
		fmt.Println("Телефон не откроет 127.0.0.1 — задайте sub_domain в настройках (п. 11):")
		fmt.Println(`  "sub_domain": "http://IP_ВАШЕГО_СЕРВЕРА:8787"`)
		fmt.Println(`  "sub_path": "sub"`)
		fmt.Println()
	} else {
		fmt.Printf("Публичный адрес: %s\n", config.SubscriptionPublicBase())
		fmt.Println()
	}

	pngPath, err := saveSubscriptionQRPNG(url, username)
	if err != nil {
		return err
	}
	fmt.Printf("QR сохранён: %s\n", pngPath)
	fmt.Println("Откройте файл на экране и отсканируйте камерой или «Импорт по QR» в клиенте.")

	if runtime.GOOS == "darwin" {
		if err := exec.Command("open", pngPath).Run(); err == nil {
			fmt.Println("(файл открыт в Preview)")
		}
	}

	if canPrintTerminalQR() {
		fmt.Println()
		fmt.Println("Превью в терминале:")
		qrterminal.Generate(url, qrterminal.M, os.Stdout)
	} else {
		fmt.Println()
		fmt.Println("Превью в терминале пропущено — используйте PNG-файл выше.")
	}

	fmt.Println()
	fmt.Println("Клиент скачает конфиг по URL и получит hy2:// ссылки со всех серверов.")
	return nil
}

func saveSubscriptionQRPNG(url, username string) (string, error) {
	name := fmt.Sprintf("qr-%s.png", username)
	path := filepath.Join(".", name)
	if err := qrcode.WriteFile(url, qrcode.Medium, 512, path); err != nil {
		return "", fmt.Errorf("сохранение QR: %w", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

func canPrintTerminalQR() bool {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < 80 {
		return false
	}
	return true
}

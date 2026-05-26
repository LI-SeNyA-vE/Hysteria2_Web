package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"hysteria2-web/internal/service"
)

func printStep(format string, args ...any) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("[%s] → %s\n", ts, fmt.Sprintf(format, args...))
}

func printOK(format string, args ...any) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("[%s] ✓ %s\n", ts, fmt.Sprintf(format, args...))
}

func printFail(err error) {
	ts := time.Now().Format("15:04:05")
	msg := humanError(err)
	fmt.Printf("[%s] ✗ %s\n", ts, msg)
}

func humanError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, service.ErrUserExists) {
		return "пользователь уже есть в локальной базе — выберите другое имя или удалите старую запись"
	}
	if errors.Is(err, service.ErrServerExists) {
		return "сервер с таким именем уже существует"
	}
	if errors.Is(err, service.ErrServerNotFound) {
		return "сервер не найден"
	}
	return humanizeAPIError(err.Error())
}

func humanizeAPIError(msg string) string {
	if strings.Contains(msg, "Username can only contain letters, numbers, and underscores") {
		return "имя пользователя: только латинские буквы, цифры и _ (проверьте раскладку EN)"
	}
	if idx := strings.Index(msg, "status 422:"); idx >= 0 {
		return strings.TrimSpace(msg[idx+len("status 422:"):])
	}
	if idx := strings.Index(msg, "status 4"); idx >= 0 {
		if end := strings.Index(msg[idx:], ": "); end >= 0 {
			return strings.TrimSpace(msg[idx+end+2:])
		}
	}
	return msg
}

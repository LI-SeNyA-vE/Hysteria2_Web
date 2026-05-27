package config

import (
	"os"
	"strings"
)

func SubscriptionPublicBase() string {
	if v := strings.TrimSpace(os.Getenv("SUB_PUBLIC_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}

	addr := EnvOrDefault("HTTP_ADDR", "0.0.0.0:8787")
	if strings.HasPrefix(addr, ":") {
		return "http://127.0.0.1" + addr
	}
	if strings.HasPrefix(addr, "0.0.0.0") {
		return "http://" + strings.Replace(addr, "0.0.0.0", "127.0.0.1", 1)
	}
	if strings.Contains(addr, "://") {
		return strings.TrimRight(addr, "/")
	}
	return "http://" + strings.TrimRight(addr, "/")
}

func SubscriptionURL(token string) string {
	return SubscriptionPublicBase() + "/sub/" + token
}

func UsingLocalSubscriptionURL() bool {
	return strings.TrimSpace(os.Getenv("SUB_PUBLIC_URL")) == ""
}

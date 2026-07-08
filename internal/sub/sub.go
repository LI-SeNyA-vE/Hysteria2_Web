// Package sub строит hysteria2:// URI и кодирует их в base64 для клиентов.
package sub

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// URIConfig — параметры для построения одного hysteria2:// URI.
type URIConfig struct {
	UserName     string
	UserPassword string
	PublicIP     string
	Hy2Port      int
	ObfsPassword string // salamander obfs, пусто — если не настроен
	CertSHA256   string // upper-colon hex (AA:BB:...), пусто — если нет сертификата
	Label        string // фрагмент (#имя) — показывается в VPN-клиенте
}

// BuildURI строит hysteria2:// URI по конфигу.
// Формат: hysteria2://user:pass@host:port/?obfs=salamander&obfs-password=X&pinSHA256=X&insecure=1#label
func BuildURI(c URIConfig) string {
	q := url.Values{}
	if c.ObfsPassword != "" {
		q.Set("obfs", "salamander")
		q.Set("obfs-password", c.ObfsPassword)
	}
	if c.CertSHA256 != "" {
		q.Set("pinSHA256", c.CertSHA256)
	}
	q.Set("insecure", "1") // нужно для self-signed даже с pinSHA256 в ряде клиентов

	u := &url.URL{
		Scheme:   "hysteria2",
		User:     url.UserPassword(c.UserName, c.UserPassword),
		Host:     fmt.Sprintf("%s:%d", c.PublicIP, c.Hy2Port),
		Path:     "/",
		RawQuery: q.Encode(),
		Fragment: c.Label,
	}
	return u.String()
}

// EncodeBase64 объединяет URI через \n и кодирует в base64 — формат подписки для клиентов.
func EncodeBase64(uris []string) string {
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(uris, "\n")))
}

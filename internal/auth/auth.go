// Package auth отвечает за bcrypt-пароль админа, JWT-сессии и first-run bootstrap.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"hysteria2-web/internal/db"
	"hysteria2-web/internal/models"
)

const tokenTTL = 7 * 24 * time.Hour

// Auth хранит секреты сессий, загруженные из БД (main) или из конфига (node).
type Auth struct {
	db           *db.DB
	jwtSecret    []byte
	passwordHash string
	nodeToken    string // для роли node: токен берётся из конфига, не из БД
}

// NewNodeAuth создаёт Auth для ноды — без БД, только с токеном для связи с main.
func NewNodeAuth(nodeToken string) *Auth {
	return &Auth{nodeToken: nodeToken}
}

// Bootstrap загружает (или при первом запуске генерирует) секреты панели.
// fixedNodeToken — если не пустой, используется вместо случайно сгенерированного
// (удобно для Docker Compose, чтобы ноды знали токен заранее).
// Возвращает сгенерированный пароль, если это был первый запуск, иначе "".
func Bootstrap(d *db.DB, fixedNodeToken string) (*Auth, string, error) {
	a := &Auth{db: d}

	secret, err := d.GetSetting(models.SettingJWTSecret)
	if err != nil {
		return nil, "", err
	}
	if secret == "" {
		secret = randomHex(32)
		if err := d.SetSetting(models.SettingJWTSecret, secret); err != nil {
			return nil, "", err
		}
	}
	a.jwtSecret = []byte(secret)

	nodeToken, err := d.GetSetting(models.SettingNodeToken)
	if err != nil {
		return nil, "", err
	}
	// fixedNodeToken (из PANEL_BOOTSTRAP_NODE_TOKEN) всегда имеет приоритет —
	// нужно для Docker Compose, где ноды знают токен заранее.
	if fixedNodeToken != "" {
		nodeToken = fixedNodeToken
	}
	if nodeToken == "" {
		nodeToken = randomHex(24)
	}
	if err := d.SetSetting(models.SettingNodeToken, nodeToken); err != nil {
		return nil, "", err
	}

	hash, err := d.GetSetting(models.SettingAdminPasswordHash)
	if err != nil {
		return nil, "", err
	}

	var plainPassword string
	if hash == "" {
		plainPassword = randomPassword(16)
		h, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, "", err
		}
		hash = string(h)
		if err := d.SetSetting(models.SettingAdminPasswordHash, hash); err != nil {
			return nil, "", err
		}
	}
	a.passwordHash = hash

	return a, plainPassword, nil
}

// NodeToken возвращает токен для аутентификации нод.
// Для main — из БД; для node — из конфига.
func (a *Auth) NodeToken() (string, error) {
	if a.nodeToken != "" {
		return a.nodeToken, nil
	}
	return a.db.GetSetting(models.SettingNodeToken)
}

// CheckPassword сверяет пароль с сохранённым bcrypt-хэшем.
func (a *Auth) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.passwordHash), []byte(password)) == nil
}

// IssueToken выпускает HS256 JWT со сроком tokenTTL.
func (a *Auth) IssueToken() (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   "admin",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.jwtSecret)
}

// Verify проверяет подпись и срок токена.
func (a *Auth) Verify(token string) error {
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		return a.jwtSecret, nil
	})
	if err != nil {
		return errors.New("невалидный токен")
	}
	return nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

const passwordAlphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func randomPassword(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = passwordAlphabet[int(b[i])%len(passwordAlphabet)]
	}
	return string(b)
}

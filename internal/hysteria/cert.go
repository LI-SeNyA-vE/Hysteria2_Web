package hysteria

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (m *Manager) certExists() bool {
	_, certErr := os.Stat(filepath.Join(m.dataDir, "cert.pem"))
	_, keyErr := os.Stat(filepath.Join(m.dataDir, "key.pem"))
	return certErr == nil && keyErr == nil
}

// generateCert создаёт self-signed ECDSA P-256 сертификат на 10 лет.
// Возвращает SHA-256 pin в формате "AA:BB:..." (uppercase colon-hex).
func (m *Manager) generateCert() (string, error) {
	if err := os.MkdirAll(m.dataDir, 0o755); err != nil {
		return "", err
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("генерация ключа: %w", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "hysteria2"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return "", fmt.Errorf("создание сертификата: %w", err)
	}

	// SHA-256 пин — верхний регистр через двоеточие
	hash := sha256.Sum256(derBytes)
	parts := make([]string, 32)
	for i, b := range hash {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	pin := strings.Join(parts, ":")

	// Записываем cert.pem
	certFile, err := os.Create(filepath.Join(m.dataDir, "cert.pem"))
	if err != nil {
		return "", fmt.Errorf("запись cert.pem: %w", err)
	}
	defer certFile.Close()
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", err
	}

	// Записываем key.pem
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("сериализация ключа: %w", err)
	}
	keyFile, err := os.Create(filepath.Join(m.dataDir, "key.pem"))
	if err != nil {
		return "", fmt.Errorf("запись key.pem: %w", err)
	}
	defer keyFile.Close()
	if err := pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}); err != nil {
		return "", err
	}

	return pin, nil
}

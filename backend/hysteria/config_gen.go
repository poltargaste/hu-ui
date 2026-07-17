package hysteria

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"hysteria-panel/backend/config"
	"hysteria-panel/backend/database"
)

// HysteriaObfsConfig представляет структуру обфускации для Hysteria 2
type HysteriaObfsConfig struct {
	Type       string                 `yaml:"type"`
	Salamander HysteriaSalamanderObfs `yaml:"salamander"`
}

type HysteriaSalamanderObfs struct {
	Password string `yaml:"password"`
}

// HysteriaConfig представляет структуру конфигурации для ядра Hysteria 2
type HysteriaConfig struct {
	Listen string `yaml:"listen"`
	TLS    struct {
		Cert string `yaml:"cert"`
		Key  string `yaml:"key"`
	} `yaml:"tls"`
	Obfs *HysteriaObfsConfig `yaml:"obfs,omitempty"`
	Auth struct {
		Type     string            `yaml:"type"`
		Userpass map[string]string `yaml:"userpass"`
	} `yaml:"auth"`
	Stats struct {
		Listen string `yaml:"listen"`
		Secret string `yaml:"secret,omitempty"`
	} `yaml:"stats"`
}

// GenerateHysteriaConfig собирает пользователей из БД и генерирует yaml-конфиг для Hysteria 2
func GenerateHysteriaConfig() error {
	cfg := config.GlobalConfig
	if cfg == nil {
		return fmt.Errorf("global config is not loaded")
	}

	// 1. Убеждаемся в наличии TLS-сертификатов. Если файлов нет, создаем самоподписанные.
	certPath := filepath.Join(filepath.Dir(cfg.HysteriaConfig), "server.crt")
	keyPath := filepath.Join(filepath.Dir(cfg.HysteriaConfig), "server.key")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		err := generateSelfSignedCert(certPath, keyPath)
		if err != nil {
			return fmt.Errorf("failed to generate self-signed cert: %w", err)
		}
	}

	// 2. Выбираем из БД только активных пользователей
	var users []database.User
	// Выбираем тех, у кого is_enabled = true и срок действия не истек
	now := time.Now()
	err := database.DB.Where("is_enabled = ?", true).
		Where("expire_date IS NULL OR expire_date > ?", now).
		Find(&users).Error
	if err != nil {
		return fmt.Errorf("failed to fetch active users: %w", err)
	}

	// Формируем карту пользователей
	userpassMap := make(map[string]string)
	for _, u := range users {
		userpassMap[u.Username] = u.AuthValue
	}

	// Создаем структуру конфигурации Hysteria 2
	hCfg := HysteriaConfig{}
	hCfg.Listen = fmt.Sprintf(":%d", cfg.HysteriaPort)
	hCfg.TLS.Cert = certPath
	hCfg.TLS.Key = keyPath

	if cfg.HysteriaObfs != "" {
		hCfg.Obfs = &HysteriaObfsConfig{
			Type: "salamander",
			Salamander: HysteriaSalamanderObfs{
				Password: cfg.HysteriaObfs,
			},
		}
	}

	hCfg.Auth.Type = "userpass"
	hCfg.Auth.Userpass = userpassMap

	// Задаем адрес статистики (будем опрашивать через HTTP JSON)
	// Слушаем локально на порту 60000
	hCfg.Stats.Listen = "127.0.0.1:60000"

	// Сериализуем в YAML
	yamlData, err := yaml.Marshal(&hCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal hysteria config: %w", err)
	}

	// Записываем файл
	err = os.WriteFile(cfg.HysteriaConfig, yamlData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write hysteria config file: %w", err)
	}

	return nil
}

// generateSelfSignedCert создает самоподписанный сертификат и закрытый ключ
func generateSelfSignedCert(certOut, keyOut string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Действителен 1 год

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Hysteria Panel Auto-Generated"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certFile, err := os.Create(certOut)
	if err != nil {
		return err
	}
	defer certFile.Close()

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyFile, err := os.OpenFile(keyOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return nil
}

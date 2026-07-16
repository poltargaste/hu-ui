package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// AppConfig хранит настройки приложения
type AppConfig struct {
	PanelHost      string `json:"panel_host"`
	PanelPort      int    `json:"panel_port"`
	WebBasePath    string `json:"web_base_path"` // Секретный префикс пути панели, например, /xOmIAGB
	DbPath         string `json:"db_path"`
	HysteriaBin    string `json:"hysteria_bin"`
	HysteriaConfig string `json:"hysteria_config"`
	HysteriaPort   int    `json:"hysteria_port"`
	HysteriaObfs   string `json:"hysteria_obfs"`
	JwtSecret      string `json:"jwt_secret"`
}

var GlobalConfig *AppConfig

// LoadConfig считывает конфигурацию из указанного файла JSON
func LoadConfig(configPath string) (*AppConfig, error) {
	config := &AppConfig{
		PanelHost:      "0.0.0.0",
		PanelPort:      8080,
		WebBasePath:    "", // По умолчанию без префикса
		DbPath:         "./hu-ui.db",
		HysteriaBin:    "./bin/hysteria",
		HysteriaConfig: "./hysteria.yaml",
		HysteriaPort:   443,
		HysteriaObfs:   "",
		JwtSecret:      "super-secret-jwt-key",
	}

	if configPath == "" {
		GlobalConfig = config
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файла нет, возвращаем дефолтную конфигурацию
			GlobalConfig = config
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// Гарантируем, что префикс начинается со слеша и не заканчивается им (кроме корня)
	if config.WebBasePath != "" {
		if !strings.HasPrefix(config.WebBasePath, "/") {
			config.WebBasePath = "/" + config.WebBasePath
		}
		config.WebBasePath = strings.TrimSuffix(config.WebBasePath, "/")
	}

	// Делаем пути абсолютными, если необходимо
	if !filepath.IsAbs(config.DbPath) {
		absPath, err := filepath.Abs(config.DbPath)
		if err == nil {
			config.DbPath = absPath
		}
	}

	GlobalConfig = config
	return config, nil
}

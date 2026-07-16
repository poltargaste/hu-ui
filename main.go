package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"hysteria-panel/backend/api"
	"hysteria-panel/backend/config"
	"hysteria-panel/backend/database"
	"hysteria-panel/backend/hysteria"
)

func main() {
	// 1. Парсинг флагов командной строки
	configFlag := flag.String("config", "", "Path to the configuration JSON file")
	flag.Parse()

	log.Println("[INFO] Starting Hysteria 2 Admin Panel...")

	// Определяем путь к конфигурации
	configPath := *configFlag
	if configPath == "" {
		// Пытаемся найти config.json в текущей папке или /etc/hysteria-panel/
		if _, err := os.Stat("config.json"); err == nil {
			configPath = "config.json"
		} else if _, err := os.Stat("/etc/hysteria-panel/config.json"); err == nil {
			configPath = "/etc/hysteria-panel/config.json"
		}
	}

	// 2. Загрузка конфигурации панели
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to load configuration: %v", err)
	}
	log.Printf("[INFO] Configuration loaded. Panel port: %d, Hysteria port: %d", cfg.PanelPort, cfg.HysteriaPort)

	// Определяем директорию конфига для инициализации админа
	configDir := "."
	if configPath != "" {
		configDir = filepath.Dir(configPath)
	}

	// 3. Инициализация базы данных SQLite
	db, err := database.InitDB(cfg.DbPath, configDir)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize database: %v", err)
	}
	_ = db

	// 4. Проверка и скачивание ядра Hysteria 2
	log.Println("[INFO] Checking Hysteria 2 core binary...")
	if err := hysteria.HysteriaMgr.DownloadIfMissing(); err != nil {
		log.Fatalf("[ERROR] Failed to check/download Hysteria 2 binary: %v", err)
	}

	// 5. Запуск VPN ядра Hysteria 2
	log.Println("[INFO] Starting Hysteria 2 core...")
	if err := hysteria.HysteriaMgr.Start(); err != nil {
		log.Printf("[WARNING] Failed to start Hysteria 2 core: %v. Please check if the port %d is free.", err, cfg.HysteriaPort)
	}

	// 6. Запуск тикера сбора статистики (каждые 10 секунд)
	log.Println("[INFO] Starting statistics polling worker...")
	hysteria.StartStatsTicker(10 * time.Second)

	// 7. Настройка API-маршрутов и раздачи статики фронтенда
	router := api.SetupRouter()
	ServeFrontend(router)

	// 8. Запуск HTTP сервера панели
	serverAddr := fmt.Sprintf("%s:%d", cfg.PanelHost, cfg.PanelPort)
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	go func() {
		log.Printf("[SUCCESS] Web Admin Panel is listening on http://%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Failed to listen and serve: %v", err)
		}
	}()

	// 9. Перехват системных сигналов для Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down Hysteria 2 Admin Panel...")

	// Останавливаем процесс ядра Hysteria 2
	if err := hysteria.HysteriaMgr.Stop(); err != nil {
		log.Printf("[WARNING] Error stopping Hysteria 2 during shutdown: %v", err)
	}

	log.Println("[SUCCESS] Hysteria 2 Panel stopped gracefully.")
}

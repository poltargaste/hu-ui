package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
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

	log.Println("[INFO] Starting hu-ui Admin Panel...")

	// Определяем путь к конфигурации
	configPath := *configFlag
	if configPath == "" {
		// Пытаемся найти config.json в текущей папке или /etc/hu-ui/
		if _, err := os.Stat("config.json"); err == nil {
			configPath = "config.json"
		} else if _, err := os.Stat("/etc/hu-ui/config.json"); err == nil {
			configPath = "/etc/hu-ui/config.json"
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

	// Выводим ссылку подключения для дефолтного клиента в лог
	printDefaultClientLink(cfg)

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

	// Формируем красивую ссылку на веб-панель
	serverIP := getLocalIP()
	panelURL := fmt.Sprintf("http://%s:%d", serverIP, cfg.PanelPort)
	if cfg.WebBasePath != "" {
		panelURL = fmt.Sprintf("http://%s:%d%s", serverIP, cfg.PanelPort, cfg.WebBasePath)
	}

	go func() {
		log.Printf("[SUCCESS] Web Admin Panel (hu-ui) is listening on: %s", panelURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Failed to listen and serve: %v", err)
		}
	}()

	// 9. Перехват системных сигналов для Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[INFO] Shutting down hu-ui Admin Panel...")

	// Останавливаем процесс ядра Hysteria 2
	if err := hysteria.HysteriaMgr.Stop(); err != nil {
		log.Printf("[WARNING] Error stopping Hysteria 2 during shutdown: %v", err)
	}

	log.Println("[SUCCESS] hu-ui Panel stopped gracefully.")
}

func printDefaultClientLink(cfg *config.AppConfig) {
	var defaultUser database.User
	err := database.DB.Where("username = ?", "default_client").First(&defaultUser).Error
	if err != nil {
		return
	}

	serverIP := getLocalIP()
	vpnLink := fmt.Sprintf("hysteria2://%s@%s:%d/?insecure=1", defaultUser.AuthValue, serverIP, cfg.HysteriaPort)
	if cfg.HysteriaObfs != "" {
		vpnLink = fmt.Sprintf("%s&obfs=aes-128-gcm&obfs-password=%s", vpnLink, cfg.HysteriaObfs)
	}
	vpnLink = fmt.Sprintf("%s#default_client-Hysteria2", vpnLink)

	log.Println("--------------------------------------------------")
	log.Println("[INFO] DEFAULT VPN CLIENT CREDENTIALS:")
	log.Printf("Username: %s", defaultUser.Username)
	log.Printf("Password: %s", defaultUser.AuthValue)
	log.Printf("Connection Link: %s", vpnLink)
	log.Println("--------------------------------------------------")
}

// getLocalIP возвращает внешний IP-адрес сервера локально, исключая приватные сети и loopback
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "YOUR_SERVER_IP"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				// Исключаем петлю 127.x.x.x и приватные RFC 1918 сети (10.x, 192.168.x, 172.16-31.x)
				if !strings.HasPrefix(ipStr, "127.") &&
					!strings.HasPrefix(ipStr, "10.") &&
					!strings.HasPrefix(ipStr, "192.168.") &&
					!(strings.HasPrefix(ipStr, "172.") && isPrivate172(ipnet.IP)) {
					return ipStr
				}
			}
		}
	}
	return "YOUR_SERVER_IP"
}

func isPrivate172(ip net.IP) bool {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}
	return ipv4[1] >= 16 && ipv4[1] <= 31
}

package database

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB инициализирует подключение к SQLite и запускает миграции
func InitDB(dbPath string, configDir string) (*gorm.DB, error) {
	// Создаем родительские папки для БД, если их нет
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// Автомиграция таблиц
	err = db.AutoMigrate(&Admin{}, &User{}, &UserStats{}, &Setting{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate database schema: %w", err)
	}

	DB = db

	// Инициализация администратора по умолчанию
	if err := seedAdmin(configDir); err != nil {
		log.Printf("Warning: failed to seed admin: %v", err)
	}

	// Инициализация первого пользователя по умолчанию
	if err := seedDefaultUser(); err != nil {
		log.Printf("Warning: failed to seed default user: %v", err)
	}

	return DB, nil
}

type initAdminData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// seedAdmin инициализирует первого администратора
func seedAdmin(configDir string) error {
	var count int64
	if err := DB.Model(&Admin{}).Count(&count).Error; err != nil {
		return err
	}

	// Если администраторы уже существуют, ничего не делаем
	if count > 0 {
		return nil
	}

	// Проверяем наличие временного файла инициализации
	initFilePath := filepath.Join(configDir, ".init_admin")
	var adminUser string
	var adminPass string

	if _, err := os.Stat(initFilePath); err == nil {
		// Читаем файл инициализации
		data, err := os.ReadFile(initFilePath)
		if err == nil {
			var initData initAdminData
			if err := json.Unmarshal(data, &initData); err == nil {
				adminUser = initData.Username
				adminPass = initData.Password
			}
		}
		// Пытаемся удалить файл сразу после чтения
		_ = os.Remove(initFilePath)
	}

	// Если файл не найден или пуст, генерируем стандартные учетные данные
	if adminUser == "" || adminPass == "" {
		adminUser = "admin"
		adminPass = "admin123" // Временный пароль по умолчанию
		log.Printf("[WARNING] No initialization credentials found. Using default: username=%s, password=%s", adminUser, adminPass)
	}

	// Хэшируем пароль
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin := Admin{
		Username:     adminUser,
		PasswordHash: string(hashedBytes),
	}

	if err := DB.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}

	log.Printf("[SUCCESS] Administrator '%s' successfully provisioned.", adminUser)
	return nil
}

// seedDefaultUser создает первого пользователя VPN, если таблица пользователей пуста
func seedDefaultUser() error {
	var count int64
	if err := DB.Model(&User{}).Count(&count).Error; err != nil {
		return err
	}

	// Если пользователи уже существуют, ничего не делаем
	if count > 0 {
		return nil
	}

	// Генерируем случайный пароль клиента
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	clientPass := base64.RawURLEncoding.EncodeToString(b)[:12]

	user := User{
		Username:  "default_client",
		AuthValue: clientPass,
		IsEnabled: true,
	}

	tx := DB.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return err
	}

	stats := UserStats{
		UserID: user.ID,
	}
	if err := tx.Create(&stats).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	log.Printf("[SUCCESS] Default VPN Client created. Username: default_client, Password: %s", clientPass)
	return nil
}

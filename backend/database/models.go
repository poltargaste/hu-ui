package database

import (
	"time"
)

// Admin представляет администратора панели
type Admin struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// User представляет VPN пользователя (клиента Hysteria 2)
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;not null" json:"username"`
	AuthValue    string     `gorm:"uniqueIndex;not null" json:"auth_value"` // Пароль авторизации в Hysteria 2
	IsEnabled    bool       `gorm:"not null;default:true" json:"is_enabled"`
	LimitSpeedTx int64      `gorm:"not null;default:0" json:"limit_speed_tx"` // Лимит отдачи (Tx) в bps (0 - без лимита)
	LimitSpeedRx int64      `gorm:"not null;default:0" json:"limit_speed_rx"` // Лимит загрузки (Rx) in bps (0 - без лимита)
	LimitTraffic int64      `gorm:"not null;default:0" json:"limit_traffic"`  // Лимит трафика в байтах (0 - без лимита)
	ExpireDate   *time.Time `json:"expire_date"`                              // Дата окончания действия аккаунта
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Stats        UserStats  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"stats"`
}

// UserStats хранит данные об использовании трафика
type UserStats struct {
	UserID       uint       `gorm:"primaryKey" json:"user_id"`
	TrafficTx    int64      `gorm:"not null;default:0" json:"traffic_tx"` // Потребленный исходящий трафик в байтах
	TrafficRx    int64      `gorm:"not null;default:0" json:"traffic_rx"` // Потребленный входящий трафик в байтах
	LastActiveAt *time.Time `json:"last_active_at"`                       // Время последней сетевой активности
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Setting представляет глобальные настройки панели и ядра Hysteria 2
type Setting struct {
	Key       string    `gorm:"primaryKey" json:"key"`
	Value     string    `gorm:"not null" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

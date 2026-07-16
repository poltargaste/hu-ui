package api

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hysteria-panel/backend/database"
	"hysteria-panel/backend/hysteria"
)

type CreateUserRequest struct {
	Username     string     `json:"username" binding:"required"`
	AuthValue    string     `json:"auth_value"` // Опционально. Если пусто, сгенерируем
	IsEnabled    bool       `json:"is_enabled"`
	LimitSpeedTx int64      `json:"limit_speed_tx"`
	LimitSpeedRx int64      `json:"limit_speed_rx"`
	LimitTraffic int64      `json:"limit_traffic"`
	ExpireDate   *time.Time `json:"expire_date"`
}

type UpdateUserRequest struct {
	Username     string     `json:"username" binding:"required"`
	AuthValue    string     `json:"auth_value" binding:"required"`
	IsEnabled    bool       `json:"is_enabled"`
	LimitSpeedTx int64      `json:"limit_speed_tx"`
	LimitSpeedRx int64      `json:"limit_speed_rx"`
	LimitTraffic int64      `json:"limit_traffic"`
	ExpireDate   *time.Time `json:"expire_date"`
}

// GetUsers возвращает список всех пользователей вместе со статистикой
func GetUsers(c *gin.Context) {
	var users []database.User
	if err := database.DB.Preload("Stats").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser создает нового VPN пользователя
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request fields"})
		return
	}

	// Если пароль пустой, генерируем случайный
	if req.AuthValue == "" {
		b := make([]byte, 12)
		_, _ = rand.Read(b)
		req.AuthValue = base64.RawURLEncoding.EncodeToString(b)[:12]
	}

	user := database.User{
		Username:     req.Username,
		AuthValue:    req.AuthValue,
		IsEnabled:    req.IsEnabled,
		LimitSpeedTx: req.LimitSpeedTx,
		LimitSpeedRx: req.LimitSpeedRx,
		LimitTraffic: req.LimitTraffic,
		ExpireDate:   req.ExpireDate,
	}

	tx := database.DB.Begin()

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this username or auth value already exists"})
		return
	}

	// Создаем начальную статистику
	stats := database.UserStats{
		UserID: user.ID,
	}
	if err := tx.Create(&stats).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user stats"})
		return
	}

	tx.Commit()

	// Перезапуск Hysteria 2 для применения изменений
	hysteria.ResetSessionStats()
	go func() {
		if err := hysteria.HysteriaMgr.Restart(); err != nil {
			// Логируем ошибку, но API возвращает OK, так как пользователь создан в БД
			_ = err
		}
	}()

	c.JSON(http.StatusCreated, user)
}

// UpdateUser изменяет параметры существующего VPN пользователя
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request fields"})
		return
	}

	var user database.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Username = req.Username
	user.AuthValue = req.AuthValue
	user.IsEnabled = req.IsEnabled
	user.LimitSpeedTx = req.LimitSpeedTx
	user.LimitSpeedRx = req.LimitSpeedRx
	user.LimitTraffic = req.LimitTraffic
	user.ExpireDate = req.ExpireDate

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or auth value already in use"})
		return
	}

	// Перезапуск Hysteria 2
	hysteria.ResetSessionStats()
	go func() {
		_ = hysteria.HysteriaMgr.Restart()
	}()

	c.JSON(http.StatusOK, user)
}

// DeleteUser удаляет пользователя
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user database.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Перезапуск Hysteria 2
	hysteria.ResetSessionStats()
	go func() {
		_ = hysteria.HysteriaMgr.Restart()
	}()

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// ResetUserStats сбрасывает накопленный трафик пользователя
func ResetUserStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var stats database.UserStats
	if err := database.DB.Where("user_id = ?", id).First(&stats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User stats not found"})
		return
	}

	stats.TrafficTx = 0
	stats.TrafficRx = 0
	stats.UpdatedAt = time.Now()

	if err := database.DB.Save(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset stats"})
		return
	}

	// Перезагрузка статистики Hysteria
	hysteria.ResetSessionStats()
	go func() {
		_ = hysteria.HysteriaMgr.Restart()
	}()

	c.JSON(http.StatusOK, gin.H{"message": "User traffic statistics reset successfully"})
}

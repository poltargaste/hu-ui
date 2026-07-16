package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hysteria-panel/backend/database"
	"hysteria-panel/backend/hysteria"
)

type SystemStatsResponse struct {
	HysteriaRunning   bool  `json:"hysteria_running"`
	ActiveConnections int   `json:"active_connections"`
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	TotalTx           int64 `json:"total_tx"`
	TotalRx           int64 `json:"total_rx"`
}

// GetSystemStats собирает и возвращает общую статистику панели
func GetSystemStats(c *gin.Context) {
	isRunning := hysteria.HysteriaMgr.IsRunning()
	connections := 0

	// Если Hysteria запущена, опрашиваем её API для получения активных подключений
	if isRunning {
		client := http.Client{Timeout: 500 * time.Millisecond}
		resp, err := client.Get("http://127.0.0.1:60000/")
		if err == nil {
			var stats hysteria.HysteriaStatsResponse
			if json.NewDecoder(resp.Body).Decode(&stats) == nil {
				connections = stats.Connections
			}
			resp.Body.Close()
		}
	}

	var totalUsers int64
	var activeUsers int64
	database.DB.Model(&database.User{}).Count(&totalUsers)
	database.DB.Model(&database.User{}).Where("is_enabled = ?", true).Count(&activeUsers)

	// Считаем общий трафик
	var traffic struct {
		TotalTx int64
		TotalRx int64
	}
	database.DB.Model(&database.UserStats{}).Select("SUM(traffic_tx) as total_tx, SUM(traffic_rx) as total_rx").Scan(&traffic)

	c.JSON(http.StatusOK, SystemStatsResponse{
		HysteriaRunning:   isRunning,
		ActiveConnections: connections,
		TotalUsers:        totalUsers,
		ActiveUsers:       activeUsers,
		TotalTx:           traffic.TotalTx,
		TotalRx:           traffic.TotalRx,
	})
}

// StartCore запускает ядро Hysteria 2 вручную
func StartCore(c *gin.Context) {
	if err := hysteria.HysteriaMgr.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Hysteria 2 started successfully"})
}

// StopCore останавливает ядро Hysteria 2 вручную
func StopCore(c *gin.Context) {
	if err := hysteria.HysteriaMgr.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Hysteria 2 stopped successfully"})
}

// RestartCore перезапускает ядро Hysteria 2 вручную
func RestartCore(c *gin.Context) {
	if err := hysteria.HysteriaMgr.Restart(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Hysteria 2 restarted successfully"})
}

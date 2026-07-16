package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"hysteria-panel/backend/config"
)

// AuthMiddleware проверяет валидность JWT токена
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.GlobalConfig.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		adminIDFloat, ok := claims["sub"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
			c.Abort()
			return
		}

		c.Set("admin_id", uint(adminIDFloat))
		c.Next()
	}
}

// CORSMiddleware разрешает кросс-доменные запросы для локальной разработки
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SetupRouter настраивает эндпоинты панели с учетом префикса пути
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Включаем CORS
	r.Use(CORSMiddleware())

	basePath := config.GlobalConfig.WebBasePath
	if basePath == "" {
		basePath = "/"
	}

	// Создаем корневую группу с префиксом пути
	baseGroup := r.Group(basePath)
	{
		// Публичные маршруты
		baseGroup.POST("/api/auth/login", Login)

		// Защищенные маршруты
		apiGroup := baseGroup.Group("/api")
		apiGroup.Use(AuthMiddleware())
		{
			// Профиль админа
			apiGroup.POST("/auth/change-password", ChangePassword)

			// Управление VPN пользователями
			apiGroup.GET("/users", GetUsers)
			apiGroup.POST("/users", CreateUser)
			apiGroup.PUT("/users/:id", UpdateUser)
			apiGroup.DELETE("/users/:id", DeleteUser)
			apiGroup.POST("/users/:id/reset", ResetUserStats)

			// Управление системой и статистика
			apiGroup.GET("/system/stats", GetSystemStats)
			apiGroup.POST("/system/core/start", StartCore)
			apiGroup.POST("/system/core/stop", StopCore)
			apiGroup.POST("/system/core/restart", RestartCore)
		}
	}

	return r
}

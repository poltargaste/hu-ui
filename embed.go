package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"hysteria-panel/backend/config"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

// ServeFrontend настраивает раздачу встроенного React-приложения
func ServeFrontend(r *gin.Engine) {
	// Достаем поддерево файлов из frontend/dist
	subFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(subFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		basePath := config.GlobalConfig.WebBasePath

		// Статические ресурсы (JS, CSS, шрифты, картинки) отдаём напрямую,
		// даже если запрос идёт без префикса basePath.
		// Vite генерирует пути вида /assets/index-abc123.js,
		// и браузер запрашивает их без префикса.
		if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/favicon") {
			file, err := subFS.Open(strings.TrimPrefix(path, "/"))
			if err == nil {
				file.Close()
				c.Request.URL.Path = path
				fileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		// Для всех остальных запросов проверяем префикс basePath
		if basePath != "" && basePath != "/" {
			if !strings.HasPrefix(path, basePath) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
				return
			}
			// Убираем префикс из пути для отдачи статики
			path = strings.TrimPrefix(path, basePath)
			if path == "" {
				path = "/"
			}
		}

		// Игнорируем запросы к API
		if strings.HasPrefix(path, "/api") {
			return
		}

		// Проверяем, существует ли физический файл во встроенной ФС
		file, err := subFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			file.Close()
			// Отдаем статический файл (например, js, css, png)
			c.Request.URL.Path = path
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Для всех остальных путей (SPA роутинг) отдаем index.html с внедренной переменной пути
		indexFile, err := subFS.Open("index.html")
		if err == nil {
			defer indexFile.Close()
			content, err := io.ReadAll(indexFile)
			if err == nil {
				// Динамически встраиваем window.basePath для React
				jsInject := fmt.Sprintf("<script>window.basePath = '%s';</script>", basePath)
				html := strings.Replace(string(content), "<div id=\"root\">", jsInject+"<div id=\"root\">", 1)
				c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})
}

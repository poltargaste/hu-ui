package main

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

		// Для всех остальных путей (SPA роутинг) отдаем index.html
		indexFile, err := subFS.Open("index.html")
		if err == nil {
			defer indexFile.Close()
			content, err := io.ReadAll(indexFile)
			if err == nil {
				c.Data(http.StatusOK, "text/html; charset=utf-8", content)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})
}

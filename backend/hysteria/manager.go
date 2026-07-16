package hysteria

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"hysteria-panel/backend/config"
)

type Manager struct {
	cmd       *exec.Cmd
	cmdMutex  sync.Mutex
	ctx       context.Context
	cancelCtx context.CancelFunc
	isRunning bool
}

var HysteriaMgr = &Manager{}

// DownloadIfMissing проверяет наличие бинарника Hysteria 2 и скачивает его при отсутствии
func (m *Manager) DownloadIfMissing() error {
	cfg := config.GlobalConfig
	if cfg == nil {
		return fmt.Errorf("global config is not loaded")
	}

	binPath := cfg.HysteriaBin
	if _, err := os.Stat(binPath); err == nil {
		log.Printf("[INFO] Hysteria 2 binary already exists at: %s", binPath)
		return nil
	}

	// Создаем директорию для бинарника, если ее нет
	binDir := filepath.Dir(binPath)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create binary directory: %w", err)
	}

	// Определяем архитектуру и формируем URL
	var archSuffix string
	switch runtime.GOARCH {
	case "amd64":
		archSuffix = "linux-amd64"
	case "arm64":
		archSuffix = "linux-arm64"
	default:
		// Для локальной разработки на Mac (например, M1/M2 - darwin-arm64, Intel - darwin-amd64)
		if runtime.GOOS == "darwin" {
			archSuffix = "darwin-" + runtime.GOARCH
		} else {
			return fmt.Errorf("unsupported CPU architecture: %s", runtime.GOARCH)
		}
	}

	var downloadOS string
	if runtime.GOOS == "darwin" {
		downloadOS = "darwin"
	} else {
		downloadOS = "linux"
	}

	// URL для скачивания последней версии Hysteria 2
	downloadURL := fmt.Sprintf("https://github.com/apernet/hysteria/releases/latest/download/hysteria-%s-%s", downloadOS, runtime.GOARCH)
	if downloadOS == "linux" {
		downloadURL = fmt.Sprintf("https://github.com/apernet/hysteria/releases/latest/download/hysteria-%s", archSuffix)
	}

	log.Printf("[INFO] Downloading Hysteria 2 binary from: %s", downloadURL)

	// Загружаем бинарник
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to make HTTP GET request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received bad status code: %d", resp.StatusCode)
	}

	// Записываем файл на диск
	out, err := os.OpenFile(binPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create executable file on disk: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write response body to file: %w", err)
	}

	log.Printf("[SUCCESS] Hysteria 2 binary successfully downloaded and saved to %s", binPath)
	return nil
}

// Start запускает процесс Hysteria 2
func (m *Manager) Start() error {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("hysteria is already running")
	}

	cfg := config.GlobalConfig
	if cfg == nil {
		return fmt.Errorf("global config is not loaded")
	}

	// Генерируем конфигурационный файл
	if err := GenerateHysteriaConfig(); err != nil {
		return fmt.Errorf("failed to generate config before starting: %w", err)
	}

	m.ctx, m.cancelCtx = context.WithCancel(context.Background())

	// Настраиваем команду для запуска
	// Запускаем с флагом server и путем к сгенерированному конфигу
	m.cmd = exec.CommandContext(m.ctx, cfg.HysteriaBin, "server", "--config", cfg.HysteriaConfig)

	// Настраиваем вывод логов
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	// Устанавливаем SysProcAttr для корректной отправки сигналов дочерним процессам
	m.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Запуск процесса
	if err := m.cmd.Start(); err != nil {
		m.cancelCtx()
		return fmt.Errorf("failed to start hysteria process: %w", err)
	}

	m.isRunning = true
	log.Printf("[SUCCESS] Hysteria 2 process started with PID %d", m.cmd.Process.Pid)

	// Следим за завершением процесса в фоне
	go func() {
		err := m.cmd.Wait()
		m.cmdMutex.Lock()
		m.isRunning = false
		if err != nil {
			log.Printf("[WARNING] Hysteria 2 process exited with error: %v", err)
		} else {
			log.Printf("[INFO] Hysteria 2 process stopped gracefully")
		}
		m.cmdMutex.Unlock()
	}()

	return nil
}

// Stop останавливает процесс Hysteria 2
func (m *Manager) Stop() error {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()

	if !m.isRunning || m.cmd == nil || m.cmd.Process == nil {
		return nil
	}

	log.Printf("[INFO] Stopping Hysteria 2 process...")

	// Посылаем сигнал SIGTERM группе процессов
	pgid, err := syscall.Getpgid(m.cmd.Process.Pid)
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = m.cmd.Process.Signal(syscall.SIGTERM)
	}

	// Даем немного времени для завершения
	done := make(chan error, 1)
	go func() {
		for {
			m.cmdMutex.Lock()
			running := m.isRunning
			m.cmdMutex.Unlock()
			if !running {
				done <- nil
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		log.Printf("[SUCCESS] Hysteria 2 process stopped")
	case <-time.After(3 * time.Second):
		// Если не остановился по SIGTERM, убиваем жестко
		log.Printf("[WARNING] Hysteria 2 didn't stop in time. Sending SIGKILL...")
		if err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			_ = m.cmd.Process.Signal(syscall.SIGKILL)
		}
		m.cancelCtx()
		m.isRunning = false
	}

	return nil
}

// Restart перезапускает процесс Hysteria 2
func (m *Manager) Restart() error {
	log.Printf("[INFO] Restarting Hysteria 2...")
	if err := m.Stop(); err != nil {
		log.Printf("[WARNING] Error stopping during restart: %v", err)
	}
	// Небольшая пауза для освобождения портов
	time.Sleep(500 * time.Millisecond)
	return m.Start()
}

// IsRunning возвращает статус работы
func (m *Manager) IsRunning() bool {
	m.cmdMutex.Lock()
	defer m.cmdMutex.Unlock()
	return m.isRunning
}

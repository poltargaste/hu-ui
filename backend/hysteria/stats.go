package hysteria

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"hysteria-panel/backend/database"
)

type HysteriaStatsResponse struct {
	Connections int `json:"connections"`
	Traffic     map[string]struct {
		Tx int64 `json:"tx"`
		Rx int64 `json:"rx"`
	} `json:"traffic"`
}

type UserTrafficSession struct {
	Tx int64
	Rx int64
}

var (
	// Храним последнее увиденное состояние трафика для вычисления дельты
	sessionTraffic      = make(map[string]UserTrafficSession)
	sessionTrafficMutex sync.Mutex
)

// ResetSessionStats сбрасывает сессионную статистику в памяти (вызывать при перезапуске Hysteria)
func ResetSessionStats() {
	sessionTrafficMutex.Lock()
	sessionTraffic = make(map[string]UserTrafficSession)
	sessionTrafficMutex.Unlock()
}

// StartStatsTicker запускает фоновый опрос API статистики Hysteria 2
func StartStatsTicker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if HysteriaMgr.IsRunning() {
				if err := PollHysteriaStats(); err != nil {
					log.Printf("[WARNING] Failed to poll hysteria statistics: %v", err)
				}
			}
		}
	}()
}

// PollHysteriaStats выполняет один цикл опроса API статистики Hysteria 2
func PollHysteriaStats() error {
	resp, err := http.Get("http://127.0.0.1:60000/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var stats HysteriaStatsResponse
	if err := json.Unmarshal(body, &stats); err != nil {
		return err
	}

	sessionTrafficMutex.Lock()
	defer sessionTrafficMutex.Unlock()

	db := database.DB
	if db == nil {
		return nil
	}

	needRestart := false
	now := time.Now()

	// 1. Обновляем статистику активных соединений и вычисляем дельту использования
	for authValue, current := range stats.Traffic {
		// Ищем пользователя по auth_value
		var user database.User
		err := db.Preload("Stats").Where("auth_value = ?", authValue).First(&user).Error
		if err != nil {
			// Пользователь не найден в БД (возможно, удален)
			continue
		}

		lastSession := sessionTraffic[authValue]

		// Вычисляем дельту
		var deltaTx int64
		var deltaRx int64

		if current.Tx >= lastSession.Tx {
			deltaTx = current.Tx - lastSession.Tx
		} else {
			deltaTx = current.Tx // Процесс перезапустился
		}

		if current.Rx >= lastSession.Rx {
			deltaRx = current.Rx - lastSession.Rx
		} else {
			deltaRx = current.Rx
		}

		// Сохраняем новое значение для следующего шага
		sessionTraffic[authValue] = UserTrafficSession{
			Tx: current.Tx,
			Rx: current.Rx,
		}

		// Обновляем статистику в БД
		if deltaTx > 0 || deltaRx > 0 {
			user.Stats.TrafficTx += deltaTx
			user.Stats.TrafficRx += deltaRx
			user.Stats.LastActiveAt = &now
			user.Stats.UpdatedAt = now

			db.Save(&user.Stats)

			// Проверяем лимит трафика
			if user.LimitTraffic > 0 && (user.Stats.TrafficTx+user.Stats.TrafficRx) >= user.LimitTraffic {
				user.IsEnabled = false
				db.Save(&user)
				log.Printf("[INFO] User '%s' exceeded traffic limit. Disabling account.", user.Username)
				needRestart = true
			}
		}
	}

	// 2. Периодически также проверяем лимиты по дате окончания (ExpireDate) для всех пользователей
	var expiredUsers []database.User
	err = db.Where("is_enabled = ?", true).
		Where("expire_date IS NOT NULL AND expire_date <= ?", now).
		Find(&expiredUsers).Error

	if err == nil && len(expiredUsers) > 0 {
		for i := range expiredUsers {
			expiredUsers[i].IsEnabled = false
			db.Save(&expiredUsers[i])
			log.Printf("[INFO] User '%s' expired. Disabling account.", expiredUsers[i].Username)
			needRestart = true
		}
	}

	// Если кого-то заблокировали, перегенерируем конфиг и рестартуем Hysteria
	if needRestart {
		// Сбрасываем сессионные данные перед рестартом
		sessionTraffic = make(map[string]UserTrafficSession)
		go func() {
			if err := HysteriaMgr.Restart(); err != nil {
				log.Printf("[ERROR] Failed to restart Hysteria after disabling user: %v", err)
			}
		}()
	}

	return nil
}

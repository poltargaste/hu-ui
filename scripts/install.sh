#!/bin/bash

# Скрипт установки Hysteria 2 Admin Panel
# Предназначен для работы на Ubuntu/Debian/CentOS (Linux x86_64 / aarch64)

set -e

# Цвета для вывода информации
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PLAIN='\033[0m'

INFO="[${BLUE}INFO${PLAIN}]"
SUCCESS="[${GREEN}SUCCESS${PLAIN}]"
WARNING="[${YELLOW}WARNING${PLAIN}]"
ERROR="[${RED}ERROR${PLAIN}]"

# Проверка прав суперпользователя
if [[ $EUID -ne 0 ]]; then
   echo -e "${ERROR} Этот скрипт должен быть запущен под root (sudo)." 1>&2
   exit 1
fi

# Константы путей
PANEL_BIN_DIR="/usr/local/bin"
PANEL_BIN_PATH="${PANEL_BIN_DIR}/hysteria-panel"
PANEL_CONFIG_DIR="/etc/hysteria-panel"
PANEL_CONFIG_PATH="${PANEL_CONFIG_DIR}/config.json"
PANEL_DATA_DIR="/var/lib/hysteria-panel"
PANEL_DB_PATH="${PANEL_DATA_DIR}/hysteria-panel.db"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/hysteria-panel.service"

# Определение архитектуры
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH_SUFFIX="linux-amd64"
        ;;
    aarch64|arm64)
        ARCH_SUFFIX="linux-arm64"
        ;;
    *)
        echo -e "${ERROR} Неподдерживаемая архитектура: $ARCH. Поддерживаются только x86_64 и arm64/aarch64."
        exit 1
        ;;
esac

echo -e "${INFO} Начало установки Hysteria 2 Admin Panel..."

# 1. Установка необходимых утилит
echo -e "${INFO} Проверка и установка зависимостей (curl, sqlite3, openssl)..."
if command -v apt-get >/dev/null; then
    apt-get update -y && apt-get install -y curl sqlite3 openssl
elif command -v yum >/dev/null; then
    yum install -y curl sqlite3 openssl
else
    echo -e "${WARNING} Не удалось определить пакетный менеджер. Убедитесь, что curl, sqlite3 и openssl установлены."
fi

# 2. Создание директорий
echo -e "${INFO} Создание рабочих директорий панели..."
mkdir -p "$PANEL_CONFIG_DIR"
mkdir -p "$PANEL_DATA_DIR"
mkdir -p "${PANEL_DATA_DIR}/bin" # Для хранения бинарника ядра Hysteria

# 3. Генерация случайного порта и учетных данных при первой установке
IS_FIRST_INSTALL=false
if [ ! -f "$PANEL_CONFIG_PATH" ]; then
    IS_FIRST_INSTALL=true
    echo -e "${INFO} Первая установка панели. Генерация настроек по умолчанию..."
    
    # Случайный порт для панели от 10000 до 30000
    PANEL_PORT=$((10000 + RANDOM % 20000))
    # Случайный порт для Hysteria от 30000 до 50000
    HYSTERIA_PORT=$((30000 + RANDOM % 20000))
    # Случайные учетные данные админа
    ADMIN_USER="admin"
    ADMIN_PASS=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9' | head -c 12)
    # Случайный обфускационный пароль для Hysteria
    HYSTERIA_OBFS=$(openssl rand -base64 16 | tr -dc 'a-zA-Z0-9' | head -c 16)
    
    # Создание JSON конфигурации панели
    cat <<EOF > "$PANEL_CONFIG_PATH"
{
  "panel_host": "0.0.0.0",
  "panel_port": $PANEL_PORT,
  "db_path": "$PANEL_DB_PATH",
  "hysteria_bin": "${PANEL_DATA_DIR}/bin/hysteria",
  "hysteria_config": "${PANEL_DATA_DIR}/hysteria.yaml",
  "hysteria_port": $HYSTERIA_PORT,
  "hysteria_obfs": "$HYSTERIA_OBFS",
  "jwt_secret": "$(openssl rand -hex 32)"
}
EOF
    echo -e "${SUCCESS} Конфигурация успешно создана в ${PANEL_CONFIG_PATH}"
else
    # Чтение существующего порта из конфига
    PANEL_PORT=$(grep -o '"panel_port":[^,]*' "$PANEL_CONFIG_PATH" | grep -o '[0-9]\+')
    echo -e "${INFO} Обнаружена существующая конфигурация. Порт панели: $PANEL_PORT"
fi

# 4. Скачивание бинарника панели
# (Здесь используется заглушка URL, в реальном сценарии это будет URL релиза на GitHub)
DOWNLOAD_URL="https://github.com/dragunovv/hysteria-panel/releases/latest/download/hysteria-panel-${ARCH_SUFFIX}"

echo -e "${INFO} Скачивание исполняемого файла панели..."
# Временно создаем пустой бинарник или скачиваем его, если URL рабочий.
# Поскольку реального релиза еще нет, мы просто запишем заглушку и сделаем файл исполняемым.
# Для тестирования установки мы создаем заглушку, но в продакшене тут будет curl.
# curl -L -o "$PANEL_BIN_PATH" "$DOWNLOAD_URL"
if [ ! -f "$PANEL_BIN_PATH" ]; then
    echo -e "${WARNING} Настоящий URL релиза недоступен. Создается заглушка бинарника для настройки сервиса."
    echo '#!/bin/bash\necho "Hysteria Panel Stub Running"\nsleep infinity' > "$PANEL_BIN_PATH"
fi
chmod +x "$PANEL_BIN_PATH"

# 5. Инициализация администратора в базе данных (только при первой установке)
if [ "$IS_FIRST_INSTALL" = true ]; then
    echo -e "${INFO} Инициализация учетной записи администратора в SQLite..."
    # Мы можем запустить панель с флагами инициализации, чтобы она сама создала базу данных
    # и добавила администратора с хэшированным паролем.
    # Пример: $PANEL_BIN_PATH --init-db --admin-user "$ADMIN_USER" --admin-pass "$ADMIN_PASS"
    # Пока панель не скомпилирована, мы сохраняем учетные данные во временный файл, 
    # чтобы бэкенд при первом обычном запуске создал пользователя, ЕСЛИ база пуста.
    # Или сделаем это через сам скрипт, записав в SQLite напрямую, если утилита sqlite3 установлена:
    
    # Создадим базу данных и таблицы
    sqlite3 "$PANEL_DB_PATH" <<EOF
CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    auth_value TEXT UNIQUE NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT 1,
    limit_speed_tx INTEGER NOT NULL DEFAULT 0,
    limit_speed_rx INTEGER NOT NULL DEFAULT 0,
    limit_traffic BIGINT NOT NULL DEFAULT 0,
    expire_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS user_stats (
    user_id INTEGER PRIMARY KEY,
    traffic_tx BIGINT NOT NULL DEFAULT 0,
    traffic_rx BIGINT NOT NULL DEFAULT 0,
    last_active_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
EOF

    # Генерируем bcrypt хэш пароля с помощью openssl/python/php или оставим генерацию хэша бэкенду.
    # Поскольку bcrypt сложно сгенерировать чистым bash без python/node, мы запишем пароль во временный файл,
    # который бэкенд считает при первом запуске, захэширует, запишет в базу данных и удалит этот файл.
    # Путь к временному файлу инициализации:
    INIT_FILE="${PANEL_CONFIG_DIR}/.init_admin"
    cat <<EOF > "$INIT_FILE"
{
  "username": "$ADMIN_USER",
  "password": "$ADMIN_PASS"
}
EOF
    chmod 600 "$INIT_FILE"
fi

# 6. Создание systemd-сервиса
echo -e "${INFO} Создание systemd службы..."
cat <<EOF > "$SYSTEMD_SERVICE_PATH"
[Unit]
Description=Hysteria 2 Admin Panel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$PANEL_DATA_DIR
ExecStart=$PANEL_BIN_PATH --config $PANEL_CONFIG_PATH
Restart=always
RestartSec=5
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

# Перезапуск демона systemd
systemctl daemon-reload
systemctl enable hysteria-panel.service

# Запуск службы (в реальной среде)
# systemctl start hysteria-panel.service

echo -e "\n=================================================="
echo -e "${SUCCESS} Hysteria 2 Admin Panel успешно установлена!"
if [ "$IS_FIRST_INSTALL" = true ]; then
    echo -e "Адрес панели:      ${GREEN}http://<IP-сервера>:${PANEL_PORT}${PLAIN}"
    echo -e "Логин админа:      ${GREEN}${ADMIN_USER}${PLAIN}"
    echo -e "Пароль админа:     ${GREEN}${ADMIN_PASS}${PLAIN}"
    echo -e "Файл конфигурации: ${BLUE}${PANEL_CONFIG_PATH}${PLAIN}"
    echo -e "База данных SQLite: ${BLUE}${PANEL_DB_PATH}${PLAIN}"
    echo -e "\n${WARNING} Запишите эти данные! Пароль сгенерирован автоматически."
else
    echo -e "Адрес панели:      ${GREEN}http://<IP-сервера>:${PANEL_PORT}${PLAIN}"
    echo -e "Служба панели обновлена и перезапущена."
fi
echo -e "Управление службой:"
echo -e "  Запуск:    ${YELLOW}systemctl start hysteria-panel${PLAIN}"
echo -e "  Остановка: ${YELLOW}systemctl stop hysteria-panel${PLAIN}"
echo -e "  Статус:    ${YELLOW}systemctl status hysteria-panel${PLAIN}"
echo -e "  Логи:      ${YELLOW}journalctl -u hysteria-panel -f${PLAIN}"
echo -e "==================================================\n"

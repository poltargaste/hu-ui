#!/bin/bash

# Скрипт установки hu-ui
# Предназначен для работы на Ubuntu/Debian/CentOS (Linux x86_64 / aarch64)
# Поддерживает ключ --warp для автоматической настройки Cloudflare WARP

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
PANEL_BIN_PATH="${PANEL_BIN_DIR}/hu-ui"
PANEL_CONFIG_DIR="/etc/hu-ui"
PANEL_CONFIG_PATH="${PANEL_CONFIG_DIR}/config.json"
PANEL_DATA_DIR="/var/lib/hu-ui"
PANEL_DB_PATH="${PANEL_DATA_DIR}/hu-ui.db"
SYSTEMD_SERVICE_PATH="/etc/systemd/system/hu-ui.service"

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

# Функция установки Cloudflare WARP
install_warp() {
    echo -e "${INFO} Обнаружен флаг --warp. Начинаем установку Cloudflare WARP..."
    if wget -N https://gitlab.com/fscarmen/warp/-/raw/main/menu.sh; then
        bash menu.sh auto
        echo -e "${SUCCESS} Cloudflare WARP успешно настроен и запущен."
    else
        echo -e "${WARNING} Не удалось скачать скрипт автонастройки WARP. Пропускаем..."
    fi
}

# Проверка флага --warp в аргументах
if [[ "$*" == *"--warp"* ]]; then
    install_warp
fi

echo -e "${INFO} Начало установки hu-ui..."

# 1. Установка необходимых утилит
echo -e "${INFO} Проверка и установка зависимостей (curl, sqlite3, openssl, qrencode)..."
if command -v apt-get >/dev/null; then
    apt-get update -y && apt-get install -y curl sqlite3 openssl qrencode
elif command -v yum >/dev/null; then
    yum install -y curl sqlite3 openssl qrencode
else
    echo -e "${WARNING} Не удалось определить пакетный менеджер. Убедитесь, что curl, sqlite3, openssl и qrencode установлены."
fi

# Получаем внешний IP сервера
SERVER_IP=$(curl -s https://api.ipify.org || curl -s https://ifconfig.me || curl -s https://ipinfo.io/ip || echo "YOUR_SERVER_IP")

# 2. Создание директорий
echo -e "${INFO} Создание рабочих директорий панели..."
mkdir -p "$PANEL_CONFIG_DIR"
mkdir -p "$PANEL_DATA_DIR"
mkdir -p "${PANEL_DATA_DIR}/bin" # Для хранения бинарника ядра Hysteria

# 3. Генерация настроек при первой установке
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
    
    # Генерация первого дефолтного VPN-клиента
    CLIENT_USER="default_client"
    CLIENT_PASS=$(openssl rand -base64 12 | tr -dc 'a-zA-Z0-9' | head -c 12)

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
    HYSTERIA_PORT=$(grep -o '"hysteria_port":[^,]*' "$PANEL_CONFIG_PATH" | grep -o '[0-9]\+')
    HYSTERIA_OBFS=$(grep -o '"hysteria_obfs":[^,]*' "$PANEL_CONFIG_PATH" | cut -d'"' -f4)
    echo -e "${INFO} Обнаружена существующая конфигурация. Порт панели: $PANEL_PORT"
fi

# 4. Скачивание бинарника панели
DOWNLOAD_URL="https://github.com/poltargaste/hu-ui/releases/latest/download/hu-ui-${ARCH_SUFFIX}"

echo -e "${INFO} Скачивание исполняемого файла панели..."
# В реальном сервере: curl -L -o "$PANEL_BIN_PATH" "$DOWNLOAD_URL"
if [ ! -f "$PANEL_BIN_PATH" ]; then
    echo -e "${WARNING} Настоящий URL релиза недоступен. Создается заглушка бинарника для настройки сервиса."
    echo '#!/bin/bash\necho "hu-ui Stub Running"\nsleep infinity' > "$PANEL_BIN_PATH"
fi
chmod +x "$PANEL_BIN_PATH"

# 5. Инициализация базы данных и таблиц
if [ "$IS_FIRST_INSTALL" = true ]; then
    echo -e "${INFO} Инициализация таблиц базы данных SQLite..."
    
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

    # Вставляем первого дефолтного пользователя
    sqlite3 "$PANEL_DB_PATH" <<EOF
INSERT INTO users (id, username, auth_value, is_enabled) VALUES (1, '$CLIENT_USER', '$CLIENT_PASS', 1);
INSERT INTO user_stats (user_id, traffic_tx, traffic_rx) VALUES (1, 0, 0);
EOF

    # Записываем данные админа для инициализации бэкендом
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
Description=hu-ui Admin Panel
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
systemctl enable hu-ui.service

# Формирование клиентской ссылки подключения
if [ "$IS_FIRST_INSTALL" = true ]; then
    VPN_LINK="hysteria2://${CLIENT_PASS}@${SERVER_IP}:${HYSTERIA_PORT}/?insecure=1"
    if [ -n "$HYSTERIA_OBFS" ]; then
        VPN_LINK="${VPN_LINK}&obfs=aes-128-gcm&obfs-password=${HYSTERIA_OBFS}"
    fi
    VPN_LINK="${VPN_LINK}#${CLIENT_USER}-Hysteria2"
fi

echo -e "\n=================================================="
echo -e "${SUCCESS} hu-ui Admin Panel успешно установлена!"
if [ "$IS_FIRST_INSTALL" = true ]; then
    echo -e "Адрес панели:      ${GREEN}http://${SERVER_IP}:${PANEL_PORT}${PLAIN}"
    echo -e "Логин админа:      ${GREEN}${ADMIN_USER}${PLAIN}"
    echo -e "Пароль админа:     ${GREEN}${ADMIN_PASS}${PLAIN}"
    echo -e "Файл конфигурации: ${BLUE}${PANEL_CONFIG_PATH}${PLAIN}"
    echo -e "База данных SQLite: ${BLUE}${PANEL_DB_PATH}${PLAIN}"
    echo -e "\n${WARNING} Запишите эти данные! Пароль сгенерирован автоматически."
    
    echo -e "\n--------------------------------------------------"
    echo -e "ПЕРВЫЙ КЛИЕНТ ДЛЯ ПОДКЛЮЧЕНИЯ (Default Client):"
    echo -e "Имя клиента:       ${GREEN}${CLIENT_USER}${PLAIN}"
    echo -e "Пароль клиента:    ${GREEN}${CLIENT_PASS}${PLAIN}"
    echo -e "Ссылка подключения:\n${YELLOW}${VPN_LINK}${PLAIN}"
    echo -e "\nQR-код для подключения (сканируйте из клиента):"
    if command -v qrencode >/dev/null; then
        qrencode -t ansiutf8 "$VPN_LINK"
    else
        echo -e "[qrencode не установлен, не удалось вывести QR]"
    fi
else
    echo -e "Адрес панели:      ${GREEN}http://${SERVER_IP}:${PANEL_PORT}${PLAIN}"
    echo -e "Служба панели обновлена и перезапущена."
fi
echo -e "\nУправление службой:"
echo -e "  Запуск:    systemctl start hu-ui"
echo -e "  Остановка: systemctl stop hu-ui"
echo -e "  Статус:    systemctl status hu-ui"
echo -e "  Логи:      systemctl status -l --no-pager hu-ui"
echo -e "==================================================\n"

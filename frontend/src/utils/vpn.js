/**
 * Генерирует ссылку подключения Hysteria 2
 * 
 * @param {Object} user Объект пользователя из БД
 * @param {Object} systemStats Системная конфигурация (порты, обфускация)
 * @returns {string} Готовая ссылка hysteria2://
 */
export const generateHysteriaUrl = (user, systemStats) => {
  const host = window.location.hostname;
  const port = systemStats?.hysteria_port || 443;
  const auth = user.auth_value;
  const obfs = systemStats?.hysteria_obfs || '';
  const label = encodeURIComponent(`${user.username}@Hysteria2`);

  let url = `hysteria2://${auth}@${host}:${port}/?insecure=1`;

  if (obfs) {
    url += `&obfs=aes-128-gcm&obfs-password=${obfs}`;
  }

  // Конвертируем скорости из bps в Mbps для клиента
  if (user.limit_speed_tx > 0) {
    const upMbps = Math.round(user.limit_speed_tx / 1000000);
    url += `&up=${upMbps > 0 ? upMbps : 1}`;
  }
  
  if (user.limit_speed_rx > 0) {
    const downMbps = Math.round(user.limit_speed_rx / 1000000);
    url += `&down=${downMbps > 0 ? downMbps : 1}`;
  }

  url += `#${label}`;
  return url;
};

/**
 * Форматирует байты в читаемый вид (GB, MB, KB)
 */
export const formatBytes = (bytes, decimals = 2) => {
  if (!bytes || bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

/**
 * Форматирует скорость из bps в Mbps/Kbps
 */
export const formatSpeed = (bps) => {
  if (!bps || bps === 0) return '∞';
  
  const mbps = bps / 1000000;
  if (mbps >= 1) {
    return `${mbps.toFixed(1)} Mbps`;
  }
  
  const kbps = bps / 1000;
  return `${kbps.toFixed(0)} Kbps`;
};

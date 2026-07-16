# hysteria-panel

A simple web admin panel for managing Hysteria 2 VPN servers. Built with Go (Gin) and React (MUI) using SQLite for storage. The frontend static files are compiled into the Go binary using `go:embed`, allowing the panel to be distributed as a single executable.

## Features

- Hysteria 2 client management (CRUD operations)
- Upload/download speed limits per user
- Traffic volume limits (data caps) and account expiration dates
- Background traffic statistics polling via Hysteria 2 HTTP API
- Automatic user disconnection (core reloads) upon reaching traffic caps or expiration
- Self-signed TLS certificate generation on the first run (if custom certs are not provided)
- Shareable connection links (`hysteria2://`) and QR codes generation
- Live control of the Hysteria 2 core process (start, stop, restart) via the web interface
- Automatic download of the Hysteria 2 binary on startup

## Quick Installation

To install the panel on Linux (Ubuntu/Debian/CentOS), run the following command:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/dragunovv/hysteria-panel/main/scripts/install.sh)
```

The script will automatically:
1. Detect architecture and download the compiled panel binary
2. Generate random ports for the web panel and Hysteria 2 core
3. Generate random administrator credentials for the initial login
4. Create the SQLite database and run migration tables
5. Set up and start the systemd service (`hysteria-panel.service`)

## Build from Source

### Requirements
- Go 1.22 or higher
- Node.js 20 or higher

### Steps

1. Build the frontend:
   ```bash
   cd frontend
   npm install
   npm run build
   cd ..
   ```

2. Compile the backend:
   ```bash
   go mod tidy
   go build -o hysteria-panel .
   ```

## Service Management

```bash
systemctl start hysteria-panel    # start panel
systemctl stop hysteria-panel     # stop panel
systemctl restart hysteria-panel  # restart panel
systemctl status hysteria-panel   # service status
journalctl -u hysteria-panel -f   # live logs
```

## Configuration Schema (config.json)

The default configuration file is generated at `/etc/hysteria-panel/config.json`:

```json
{
  "panel_host": "0.0.0.0",
  "panel_port": 12345,
  "db_path": "/var/lib/hysteria-panel/hysteria-panel.db",
  "hysteria_bin": "/var/lib/hysteria-panel/bin/hysteria",
  "hysteria_config": "/var/lib/hysteria-panel/hysteria.yaml",
  "hysteria_port": 34567,
  "hysteria_obfs": "random_obfs_password",
  "jwt_secret": "random_jwt_secret_hash"
}
```

## License

MIT

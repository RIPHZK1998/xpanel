# xPanel - VPN User Management System

A production-ready VPN user management system with xray-core integration. Includes a web-based admin panel and distributed node agents.

## Features

- **User Management**: Registration, authentication, JWT-based auth with refresh tokens
- **Subscription Plans**: Flexible plans with data limits and expiration
- **Multi-Node Support**: Manage multiple xray-core VPN nodes from one panel
- **Traffic Tracking**: Per-user, per-node traffic statistics
- **Protocol Support**: VLESS, VMess, Trojan with TLS/Reality
- **Clean Architecture**: Layered design with clear separation of concerns

## Quick Install

### Panel Server

Install on your main server (requires PostgreSQL and Redis):

```bash
curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/panel/install.sh | sudo bash
```

Or with command-line options:

```bash
curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/panel/install.sh | sudo bash -s -- \
  --db-host localhost \
  --db-user xpanel \
  --db-password YOUR_PASSWORD \
  --db-name xpanel \
  --port 8080
```

### Node Agent

Install on each VPN node server:

```bash
curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/agent/install.sh | sudo bash -s -- \
  --node-id 1 \
  --panel-url https://your-panel.example.com \
  --api-key YOUR_API_KEY
```

Get the API key from Panel Settings page after installation.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      xPanel System                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐         ┌──────────────────────────┐  │
│  │   Admin Panel   │◄───────►│      Node Agent 1        │  │
│  │   (Web UI)      │         │  ┌────────────────────┐  │  │
│  │                 │         │  │    xray-core       │  │  │
│  │  - Users        │         │  │  (VLESS/Reality)   │  │  │
│  │  - Nodes        │         │  └────────────────────┘  │  │
│  │  - Plans        │         └──────────────────────────┘  │
│  │  - Settings     │                                       │
│  │                 │         ┌──────────────────────────┐  │
│  │  PostgreSQL ────┼────────►│      Node Agent 2        │  │
│  │  Redis          │         │  ┌────────────────────┐  │  │
│  └─────────────────┘         │  │    xray-core       │  │  │
│                              │  │  (VLESS/Reality)   │  │  │
│                              │  └────────────────────┘  │  │
│                              └──────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Requirements

### Panel Server
- Ubuntu 20.04+ / Debian 11+
- PostgreSQL 12+
- Redis 6+
- 1GB RAM minimum

### Node Server
- Ubuntu 20.04+ / Debian 11+
- Public IP address
- Port 443 open (or custom port)

## Configuration

### Panel

Configuration file: `/etc/xpanel/.env`

```env
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release

DB_HOST=localhost
DB_PORT=5432
DB_USER=xpanel
DB_PASSWORD=your_password
DB_NAME=xpanel

REDIS_HOST=localhost
REDIS_PORT=6379
```

### Node Agent

Configuration file: `/etc/xpanel-agent/config.yaml`

```yaml
node:
  id: 1
  name: "node-1"

panel:
  url: "https://your-panel.example.com"
  api_key: "your-api-key"
```

## Default Login

After installation, access the panel at `http://YOUR_IP:8080`

- **Email**: admin@xpanel.local
- **Password**: admin123

⚠️ **Change the default password immediately!**

## Service Management

```bash
# Panel
sudo systemctl start xpanel
sudo systemctl stop xpanel
sudo systemctl status xpanel
journalctl -u xpanel -f

# Node Agent
sudo systemctl start xpanel-agent
sudo systemctl stop xpanel-agent
sudo systemctl status xpanel-agent
journalctl -u xpanel-agent -f
```

## Update

To update to the latest version:

```bash
# Panel
curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/panel/install.sh | sudo bash

# Node Agent
curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/agent/install.sh | sudo bash -s -- \
  --node-id YOUR_NODE_ID \
  --panel-url YOUR_PANEL_URL \
  --api-key YOUR_API_KEY
```

## Development

```bash
# Clone repository
git clone https://github.com/RIPHZK1998/xpanel.git
cd xpanel

# Install dependencies
go mod download

# Setup environment
cp .env.example .env
# Edit .env with your settings

# Run panel
go run main.go

# Run agent (in xpanel-agent directory)
cd xpanel-agent
go run main.go -config config.yaml
```

## License

MIT License

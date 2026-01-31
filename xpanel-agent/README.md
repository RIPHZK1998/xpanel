# xpanel-agent

Node agent for xpanel VPN management system. Runs on VPN servers to manage xray-core and communicate with the management panel.

## Features

- **xray-core Management**: Automatically starts, monitors, and restarts xray-core
- **User Synchronization**: Fetches active users from panel and provisions to xray-core
- **Heartbeat Reporting**: Sends node status, system metrics (CPU, memory, uptime)
- **Traffic Reporting**: Collects and reports per-user traffic statistics
- **Auto-provisioning**: Dynamically adds/removes users based on panel state

## Installation

### Prerequisites

- xray-core installed (`/usr/local/bin/xray`)
- SSL certificates (if using TLS)
- Access to management panel

### Build

```bash
cd xpanel-agent
go build -o xpanel-agent
```

### Install

```bash
sudo cp xpanel-agent /usr/local/bin/
sudo mkdir -p /etc/xpanel-agent
sudo cp config.yaml.example /etc/xpanel-agent/config.yaml
```

## Configuration

Edit `/etc/xpanel-agent/config.yaml`:

```yaml
node:
  id: 1                    # Node ID from management panel
  name: "US-West-1"

panel:
  url: "https://panel.example.com"
  api_key: "your-api-key"

xray:
  binary_path: "/usr/local/bin/xray"
  protocol: "vless"
  listen_port: 443
  tls_enabled: true
  cert_file: "/etc/ssl/certs/server.crt"
  key_file: "/etc/ssl/private/server.key"
```

## Usage

### Run directly

```bash
xpanel-agent -config /etc/xpanel-agent/config.yaml
```

### Run as systemd service

Create `/etc/systemd/system/xpanel-agent.service`:

```ini
[Unit]
Description=xpanel Node Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/xpanel-agent -config /etc/xpanel-agent/config.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable xpanel-agent
sudo systemctl start xpanel-agent
sudo systemctl status xpanel-agent
```

## How It Works

1. **Startup**
   - Generates xray-core config
   - Starts xray-core process
   - Performs initial user sync

2. **Heartbeat Loop** (every 30s)
   - Collects system metrics
   - Sends status to panel
   - Panel updates node status

3. **User Sync Loop** (every 5m)
   - Fetches active users from panel
   - Adds new users to xray-core
   - Removes inactive users

4. **Traffic Report Loop** (every 1m)
   - Queries xray-core for user stats
   - Reports to panel
   - Resets counters

## Logs

View logs:

```bash
sudo journalctl -u xpanel-agent -f
```

## Troubleshooting

### Agent won't start

- Check config file syntax
- Verify panel URL is accessible
- Ensure xray binary exists

### Users not provisioning

- Check panel API connectivity
- Verify node ID matches panel
- Check xray-core is running

### Traffic not reporting

- Verify xray API is accessible
- Check xray-core stats are enabled
- Review agent logs

## Architecture

```
xpanel-agent
├── Manages xray-core process
├── Syncs users from panel
├── Reports heartbeat & traffic
└── Auto-restarts on crash

xray-core
├── Handles VPN connections
├── Provides API for user management
└── Tracks traffic statistics
```

## Security

- Use HTTPS for panel communication
- Protect API keys
- Run as dedicated user (not root) in production
- Use firewall to restrict xray API access

## License

Same as xpanel management panel

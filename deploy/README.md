# xPanel Deployment Guide

## Overview

This guide covers deploying xPanel (management panel) and xpanel-agent (VPN node agents) to cloud servers.

## Requirements

### Panel Server
- Ubuntu 22.04 LTS (or compatible)
- PostgreSQL 14+
- Redis 6+
- 1GB+ RAM, 1 vCPU

### Agent Nodes
- Ubuntu 22.04 / Debian 12
- 512MB+ RAM, 1 vCPU
- Port 443 accessible (or configured VPN port)

---

## Quick Start

### Build Releases

```bash
# Build panel
./deploy/panel/build.sh

# Build agent (for amd64 and arm64)
./deploy/agent/build.sh
```

Output: `dist/xpanel-linux-amd64.tar.gz` and agent binaries.

---

## Panel Deployment

### 1. Prepare Server

```bash
# Install dependencies
apt update && apt install -y postgresql redis-server

# Create database
sudo -u postgres createuser xpanel
sudo -u postgres createdb -O xpanel xpanel
sudo -u postgres psql -c "ALTER USER xpanel PASSWORD 'secure_password';"
```

### 2. Install Panel

```bash
# Upload and extract release
tar -xzf xpanel-linux-amd64.tar.gz
cd panel

# Run installer
sudo ./install.sh --binary ./xpanel
```

### 3. Configure

```bash
# Edit configuration
sudo nano /etc/xpanel/.env

# Required changes:
# - DB_PASSWORD (your PostgreSQL password)

# Auto-generated (no manual config needed):
# - JWT_SECRET (stored in database)
# - NODE_API_KEY (view/change from Settings page)
```

### 4. Start Service

```bash
sudo systemctl start xpanel
sudo systemctl enable xpanel

# Check logs
journalctl -u xpanel -f
```

Default login: `admin@xpanel.local` / `admin123`

---

## Agent Deployment

### 1. Create Node in Panel

1. Login to panel → Nodes → Add Node
2. Configure: Name, Address, Port, Protocol (VLESS + Reality recommended)
3. Note the **Node ID**

### 2. Install Agent

```bash
# Upload agent files to VPS
scp xpanel-agent install.sh user@node-server:/tmp/

# SSH to node and install
sudo ./install.sh \
  --binary ./xpanel-agent \
  --node-id 1 \
  --panel-url https://your-panel.com \
  --api-key YOUR_NODE_API_KEY
```

### 3. Verify

```bash
# Start and enable service
sudo systemctl start xpanel-agent
sudo systemctl enable xpanel-agent

# Check logs
journalctl -u xpanel-agent -f

# Expected: "Node sync successful" messages
```

---

## Logs & Monitoring

### View Logs

```bash
# Panel logs
journalctl -u xpanel -f

# Agent logs  
journalctl -u xpanel-agent -f

# Last 100 lines
journalctl -u xpanel -n 100
```

### Log Rotation

Logs are handled by systemd journald with automatic rotation. Configure in `/etc/systemd/journald.conf`:

```ini
[Journal]
SystemMaxUse=500M
MaxRetentionSec=1month
```

---

## Security Checklist

- [ ] Change default admin password
- [ ] Configure firewall (ufw)
- [ ] Use HTTPS for panel (nginx reverse proxy)
- [ ] Use Reality protocol for VPN (no certificates needed)
- [ ] Secure database access

---

## File Locations

| Component | Path |
|-----------|------|
| Panel binary | `/opt/xpanel/xpanel` |
| Panel web | `/opt/xpanel/web/` |
| Panel config | `/etc/xpanel/.env` |
| Agent binary | `/opt/xpanel-agent/xpanel-agent` |
| Agent config | `/etc/xpanel-agent/config.yaml` |
| Xray geodata | `/usr/local/share/xray/` |

---

## Troubleshooting

### Panel won't start
```bash
# Check config
cat /etc/xpanel/.env

# Test database connection
psql -h localhost -U xpanel -d xpanel

# Check Redis
redis-cli ping
```

### Agent can't connect to panel
```bash
# Test connectivity
curl -v https://your-panel.com/health

# Check API key matches
grep NODE_API_KEY /etc/xpanel/.env
grep api_key /etc/xpanel-agent/config.yaml
```

### Users not syncing
```bash
# Check agent logs for errors
journalctl -u xpanel-agent | grep -i error

# Verify node ID matches panel
```

#!/bin/bash
#
# xPanel Agent - Node Agent Installation Script
# Tested on: Ubuntu 22.04 LTS, Debian 12
#
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  xPanel Agent - Node Installer        ${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Configuration
INSTALL_DIR="/opt/xpanel-agent"
CONFIG_DIR="/etc/xpanel-agent"
XRAY_DIR="/usr/local/share/xray"

# Parse arguments
BINARY_PATH=""
PANEL_URL=""
API_KEY=""
NODE_ID=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --binary)
            BINARY_PATH="$2"
            shift 2
            ;;
        --panel-url)
            PANEL_URL="$2"
            shift 2
            ;;
        --api-key)
            API_KEY="$2"
            shift 2
            ;;
        --node-id)
            NODE_ID="$2"
            shift 2
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Validate required arguments
if [ -z "$BINARY_PATH" ]; then
    echo -e "${YELLOW}Usage: $0 --binary /path/to/xpanel-agent --node-id <ID> --panel-url <URL> --api-key <KEY>${NC}"
    exit 1
fi

if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Binary not found: $BINARY_PATH${NC}"
    exit 1
fi

echo -e "\n${GREEN}[1/6] Creating directories...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$XRAY_DIR"

echo -e "\n${GREEN}[2/6] Installing xray-core...${NC}"
if ! command -v xray &> /dev/null; then
    echo "Installing xray-core..."
    # Download latest xray-core
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) XRAY_ARCH="64" ;;
        aarch64) XRAY_ARCH="arm64-v8a" ;;
        *) echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1 ;;
    esac
    
    XRAY_VERSION=$(curl -s https://api.github.com/repos/XTLS/Xray-core/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    XRAY_URL="https://github.com/XTLS/Xray-core/releases/download/${XRAY_VERSION}/Xray-linux-${XRAY_ARCH}.zip"
    
    echo "Downloading xray-core ${XRAY_VERSION}..."
    curl -sL "$XRAY_URL" -o /tmp/xray.zip
    unzip -q /tmp/xray.zip -d /tmp/xray
    
    mv /tmp/xray/xray /usr/local/bin/xray
    chmod +x /usr/local/bin/xray
    
    # Install geoip/geosite
    if [ -f /tmp/xray/geoip.dat ]; then
        mv /tmp/xray/geoip.dat "$XRAY_DIR/"
    fi
    if [ -f /tmp/xray/geosite.dat ]; then
        mv /tmp/xray/geosite.dat "$XRAY_DIR/"
    fi
    
    rm -rf /tmp/xray /tmp/xray.zip
    echo "xray-core installed: $(xray version | head -1)"
else
    echo "xray-core already installed: $(xray version | head -1)"
fi

echo -e "\n${GREEN}[3/6] Installing agent binary...${NC}"
cp "$BINARY_PATH" "$INSTALL_DIR/xpanel-agent"
chmod +x "$INSTALL_DIR/xpanel-agent"
echo "Installed: $INSTALL_DIR/xpanel-agent"

echo -e "\n${GREEN}[4/6] Creating configuration...${NC}"
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    cat > "$CONFIG_DIR/config.yaml" << EOF
# ============================================
# xPanel Agent Configuration
# ============================================

# Node Identification
node:
  id: ${NODE_ID:-1}
  name: "$(hostname)"

# Panel Connection
panel:
  url: "${PANEL_URL:-http://your-panel-server:8080}"
  api_key: "${API_KEY:-YOUR_API_KEY_HERE}"
  timeout: 10s

# Xray API Settings
xray:
  api_address: "127.0.0.1"
  api_port: 10085

# File Paths
files:
  tls_cert: "/etc/xpanel-agent/server.crt"
  tls_key: "/etc/xpanel-agent/server.key"
  geoip: "$XRAY_DIR/geoip.dat"
  geosite: "$XRAY_DIR/geosite.dat"
  xray_config: "/tmp/xray-config.json"

# Sync Intervals
intervals:
  heartbeat: 30s
  user_sync: 1m
  traffic_report: 1m
  activity_report: 30s

# Logging
logging:
  level: "info"
  format: "json"
EOF
    chmod 600 "$CONFIG_DIR/config.yaml"
    echo -e "${YELLOW}IMPORTANT: Edit $CONFIG_DIR/config.yaml with correct settings!${NC}"
else
    echo "Configuration already exists, skipping..."
fi

echo -e "\n${GREEN}[5/6] Installing systemd service...${NC}"
cat > /etc/systemd/system/xpanel-agent.service << 'EOF'
[Unit]
Description=xPanel Node Agent
Documentation=https://github.com/yourorg/xpanel
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/xpanel-agent
ExecStart=/opt/xpanel-agent/xpanel-agent -config /etc/xpanel-agent/config.yaml
Restart=always
RestartSec=10

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=xpanel-agent

# Security (root needed for xray)
NoNewPrivileges=false
PrivateTmp=true

# Resource limits
LimitNOFILE=1048576
LimitNPROC=512

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
echo "Installed: /etc/systemd/system/xpanel-agent.service"

echo -e "\n${GREEN}[6/6] Final setup...${NC}"

# Open common VPN ports (if ufw is installed)
if command -v ufw &> /dev/null; then
    echo "Configuring firewall..."
    ufw allow 443/tcp comment 'xpanel-agent VPN' 2>/dev/null || true
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!               ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: nano $CONFIG_DIR/config.yaml"
echo "     - Set correct node.id (from panel)"
echo "     - Set panel.url (your panel server address)"
echo "     - Set panel.api_key (from panel Settings)"
echo ""
echo "  2. Start service: systemctl start xpanel-agent"
echo "  3. Enable on boot: systemctl enable xpanel-agent"
echo "  4. Check logs: journalctl -u xpanel-agent -f"
echo ""
echo -e "${CYAN}Tips:${NC}"
echo "  - Node configuration (protocol, port, TLS/Reality) is managed from the panel"
echo "  - Reality protocol is recommended (no certificates needed)"
echo "  - For TLS mode, place certificates in $CONFIG_DIR/"

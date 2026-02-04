#!/bin/bash
#
# xPanel Agent - Node Installation Script
# One-line install: curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/agent/install.sh | sudo bash -s -- --node-id 1 --panel-url https://panel.example.com --api-key YOUR_KEY
#
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
REPO_URL="https://github.com/RIPHZK1998/xpanel.git"
INSTALL_DIR="/opt/xpanel-agent"
CONFIG_DIR="/etc/xpanel-agent"
XRAY_DIR="/usr/local/share/xray"
GO_VERSION="1.22.0"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  xPanel Agent - Node Installer        ${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Parse arguments
NODE_ID=""
PANEL_URL=""
API_KEY=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --node-id) NODE_ID="$2"; shift 2 ;;
        --panel-url) PANEL_URL="$2"; shift 2 ;;
        --api-key) API_KEY="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Interactive prompts for missing values
if [ -z "$NODE_ID" ]; then
    read -p "Node ID (from panel): " NODE_ID
    if [ -z "$NODE_ID" ]; then
        echo -e "${RED}Node ID is required${NC}"
        exit 1
    fi
fi

if [ -z "$PANEL_URL" ]; then
    read -p "Panel URL (e.g., https://panel.example.com): " PANEL_URL
    if [ -z "$PANEL_URL" ]; then
        echo -e "${RED}Panel URL is required${NC}"
        exit 1
    fi
fi

if [ -z "$API_KEY" ]; then
    read -sp "API Key (from panel Settings): " API_KEY
    echo
    if [ -z "$API_KEY" ]; then
        echo -e "${RED}API Key is required${NC}"
        exit 1
    fi
fi

echo -e "\n${GREEN}[1/7] Installing dependencies...${NC}"
apt-get update -qq
apt-get install -y -qq git curl wget unzip > /dev/null

echo -e "\n${GREEN}[2/7] Installing Go ${GO_VERSION}...${NC}"
if ! command -v go &> /dev/null; then
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
    echo "Go ${GO_VERSION} installed"
else
    echo "Go already installed: $(go version)"
fi
export PATH=$PATH:/usr/local/go/bin

echo -e "\n${GREEN}[3/7] Installing xray-core...${NC}"
if ! command -v xray &> /dev/null; then
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
    mkdir -p /tmp/xray
    unzip -q /tmp/xray.zip -d /tmp/xray
    
    mv /tmp/xray/xray /usr/local/bin/xray
    chmod +x /usr/local/bin/xray
    
    mkdir -p "$XRAY_DIR"
    [ -f /tmp/xray/geoip.dat ] && mv /tmp/xray/geoip.dat "$XRAY_DIR/"
    [ -f /tmp/xray/geosite.dat ] && mv /tmp/xray/geosite.dat "$XRAY_DIR/"
    
    rm -rf /tmp/xray /tmp/xray.zip
    echo "xray-core installed: $(xray version | head -1)"
else
    echo "xray-core already installed: $(xray version | head -1)"
fi

echo -e "\n${GREEN}[4/7] Cloning/updating repository...${NC}"
TEMP_DIR="/tmp/xpanel-build"
rm -rf "$TEMP_DIR"
git clone --depth 1 "$REPO_URL" "$TEMP_DIR"

echo -e "\n${GREEN}[5/7] Building xPanel Agent...${NC}"
cd "$TEMP_DIR/xpanel-agent"
CGO_ENABLED=0 go build -ldflags="-s -w" -o xpanel-agent .
mkdir -p "$INSTALL_DIR"
mv xpanel-agent "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/xpanel-agent"
rm -rf "$TEMP_DIR"
echo "Built: $INSTALL_DIR/xpanel-agent"

echo -e "\n${GREEN}[6/7] Creating configuration...${NC}"
mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/config.yaml" ]; then
    cat > "$CONFIG_DIR/config.yaml" << EOF
# xPanel Agent Configuration

node:
  id: ${NODE_ID}
  name: "$(hostname)"

panel:
  url: "${PANEL_URL}"
  api_key: "${API_KEY}"
  timeout: 10s

xray:
  api_address: "127.0.0.1"
  api_port: 10085

files:
  tls_cert: "/etc/xpanel-agent/server.crt"
  tls_key: "/etc/xpanel-agent/server.key"
  geoip: "${XRAY_DIR}/geoip.dat"
  geosite: "${XRAY_DIR}/geosite.dat"
  xray_config: "/tmp/xray-config.json"

intervals:
  heartbeat: 30s
  user_sync: 1m
  traffic_report: 1m
  activity_report: 30s

logging:
  level: "info"
  format: "json"
EOF
    chmod 600 "$CONFIG_DIR/config.yaml"
    echo "Created configuration: $CONFIG_DIR/config.yaml"
else
    echo "Configuration already exists, skipping..."
fi

echo -e "\n${GREEN}[7/7] Installing systemd service...${NC}"
cat > /etc/systemd/system/xpanel-agent.service << 'EOF'
[Unit]
Description=xPanel Node Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/etc/xpanel-agent
ExecStart=/opt/xpanel-agent/xpanel-agent -config /etc/xpanel-agent/config.yaml
Restart=always
RestartSec=10

StandardOutput=journal
StandardError=journal
SyslogIdentifier=xpanel-agent

NoNewPrivileges=false
PrivateTmp=true
LimitNOFILE=1048576
LimitNPROC=512

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload

# Open VPN port
if command -v ufw &> /dev/null; then
    ufw allow 443/tcp comment 'xpanel-agent VPN' 2>/dev/null || true
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!               ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Configuration:"
echo "  Node ID: ${NODE_ID}"
echo "  Panel URL: ${PANEL_URL}"
echo ""
echo "Next steps:"
echo "  1. Start service: systemctl start xpanel-agent"
echo "  2. Enable on boot: systemctl enable xpanel-agent"
echo "  3. Check logs: journalctl -u xpanel-agent -f"
echo ""
echo -e "${CYAN}Tips:${NC}"
echo "  - Node settings (protocol, port, TLS/Reality) are managed from the panel"
echo "  - Reality protocol is recommended (no certificates needed)"
echo "  - Edit config: nano $CONFIG_DIR/config.yaml"

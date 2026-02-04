#!/bin/bash
#
# xPanel - Panel Server Installation Script
# One-line install: curl -sSL https://raw.githubusercontent.com/RIPHZK1998/xpanel/main/deploy/panel/install.sh | sudo bash
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
INSTALL_DIR="/opt/xpanel"
CONFIG_DIR="/etc/xpanel"
DATA_DIR="/var/lib/xpanel"
LOG_DIR="/var/log/xpanel"
USER="xpanel"
GROUP="xpanel"
GO_VERSION="1.22.0"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  xPanel - Panel Server Installer      ${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Parse command-line arguments
DB_HOST="localhost"
DB_PORT="5432"
DB_USER=""
DB_PASSWORD=""
DB_NAME="xpanel"
REDIS_HOST="localhost"
REDIS_PORT="6379"
SERVER_PORT="8080"

while [[ $# -gt 0 ]]; do
    case $1 in
        --db-host) DB_HOST="$2"; shift 2 ;;
        --db-port) DB_PORT="$2"; shift 2 ;;
        --db-user) DB_USER="$2"; shift 2 ;;
        --db-password) DB_PASSWORD="$2"; shift 2 ;;
        --db-name) DB_NAME="$2"; shift 2 ;;
        --redis-host) REDIS_HOST="$2"; shift 2 ;;
        --redis-port) REDIS_PORT="$2"; shift 2 ;;
        --port) SERVER_PORT="$2"; shift 2 ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# Interactive prompts for missing required values
if [ -z "$DB_USER" ]; then
    read -p "PostgreSQL username [xpanel]: " DB_USER
    DB_USER=${DB_USER:-xpanel}
fi

if [ -z "$DB_PASSWORD" ]; then
    read -sp "PostgreSQL password: " DB_PASSWORD
    echo
    if [ -z "$DB_PASSWORD" ]; then
        echo -e "${RED}Password cannot be empty${NC}"
        exit 1
    fi
fi

echo -e "\n${GREEN}[1/8] Installing dependencies...${NC}"
apt-get update -qq
apt-get install -y -qq git curl wget unzip > /dev/null

echo -e "\n${GREEN}[2/8] Installing Go ${GO_VERSION}...${NC}"
if ! command -v go &> /dev/null || [[ "$(go version)" != *"go${GO_VERSION}"* ]]; then
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

echo -e "\n${GREEN}[3/8] Creating system user...${NC}"
if ! id "$USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "$USER"
    echo "Created user: $USER"
else
    echo "User $USER already exists"
fi

echo -e "\n${GREEN}[4/8] Cloning/updating repository...${NC}"
if [ -d "$INSTALL_DIR/.git" ]; then
    cd "$INSTALL_DIR"
    git fetch origin
    git reset --hard origin/main
    echo "Updated existing installation"
else
    rm -rf "$INSTALL_DIR"
    git clone "$REPO_URL" "$INSTALL_DIR"
    echo "Cloned repository"
fi

echo -e "\n${GREEN}[5/8] Building xPanel...${NC}"
cd "$INSTALL_DIR"
CGO_ENABLED=0 go build -ldflags="-s -w" -o xpanel .
chmod +x xpanel
echo "Built: $INSTALL_DIR/xpanel"

echo -e "\n${GREEN}[6/8] Setting up directories...${NC}"
mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"
chown -R "$USER:$GROUP" "$DATA_DIR" "$LOG_DIR"

echo -e "\n${GREEN}[7/8] Creating configuration...${NC}"
if [ ! -f "$CONFIG_DIR/.env" ]; then
    cat > "$CONFIG_DIR/.env" << EOF
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=${SERVER_PORT}
SERVER_MODE=release

# Database Configuration
DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=${REDIS_PORT}
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration (auto-generated on first startup)
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168
EOF
    chmod 600 "$CONFIG_DIR/.env"
    chown "$USER:$GROUP" "$CONFIG_DIR/.env"
    echo "Created configuration: $CONFIG_DIR/.env"
else
    echo "Configuration already exists, skipping..."
fi

echo -e "\n${GREEN}[8/8] Installing systemd service...${NC}"
cat > /etc/systemd/system/xpanel.service << 'EOF'
[Unit]
Description=xPanel - VPN User Management Panel
After=network.target postgresql.service redis.service
Wants=postgresql.service redis.service

[Service]
Type=simple
User=xpanel
Group=xpanel
WorkingDirectory=/opt/xpanel
EnvironmentFile=/etc/xpanel/.env
ExecStart=/opt/xpanel/xpanel
Restart=always
RestartSec=10

StandardOutput=journal
StandardError=journal
SyslogIdentifier=xpanel

NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/xpanel /var/log/xpanel /opt/xpanel

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
echo "Installed systemd service"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!               ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Ensure PostgreSQL and Redis are running"
echo "  2. Create database:"
echo "     sudo -u postgres createuser ${DB_USER}"
echo "     sudo -u postgres createdb -O ${DB_USER} ${DB_NAME}"
echo "     sudo -u postgres psql -c \"ALTER USER ${DB_USER} PASSWORD '***';\""
echo ""
echo "  3. Start service: systemctl start xpanel"
echo "  4. Enable on boot: systemctl enable xpanel"
echo "  5. Access panel: http://YOUR_IP:${SERVER_PORT}"
echo ""
echo -e "${CYAN}Default admin login:${NC}"
echo "  Email: admin@xpanel.local"
echo "  Password: admin123"
echo ""
echo -e "${RED}IMPORTANT: Change the default password immediately!${NC}"

#!/bin/bash
#
# xPanel - Panel Server Installation Script
# Tested on: Ubuntu 22.04 LTS
#
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  xPanel - Panel Server Installer      ${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root${NC}"
    exit 1
fi

# Configuration
INSTALL_DIR="/opt/xpanel"
CONFIG_DIR="/etc/xpanel"
DATA_DIR="/var/lib/xpanel"
LOG_DIR="/var/log/xpanel"
USER="xpanel"
GROUP="xpanel"

# Parse arguments
BINARY_PATH=""
SKIP_DB=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --binary)
            BINARY_PATH="$2"
            shift 2
            ;;
        --skip-db)
            SKIP_DB=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Check for binary
if [ -z "$BINARY_PATH" ]; then
    echo -e "${YELLOW}Usage: $0 --binary /path/to/xpanel [--skip-db]${NC}"
    exit 1
fi

if [ ! -f "$BINARY_PATH" ]; then
    echo -e "${RED}Binary not found: $BINARY_PATH${NC}"
    exit 1
fi

echo -e "\n${GREEN}[1/7] Creating system user...${NC}"
if ! id "$USER" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "$USER"
    echo "Created user: $USER"
else
    echo "User $USER already exists"
fi

echo -e "\n${GREEN}[2/7] Creating directories...${NC}"
mkdir -p "$INSTALL_DIR/web"
mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"
chown -R "$USER:$GROUP" "$DATA_DIR" "$LOG_DIR"

echo -e "\n${GREEN}[3/7] Installing binary...${NC}"
cp "$BINARY_PATH" "$INSTALL_DIR/xpanel"
chmod +x "$INSTALL_DIR/xpanel"
echo "Installed: $INSTALL_DIR/xpanel"

echo -e "\n${GREEN}[4/7] Copying web files...${NC}"
if [ -d "$(dirname "$BINARY_PATH")/web" ]; then
    cp -r "$(dirname "$BINARY_PATH")/web/"* "$INSTALL_DIR/web/"
    echo "Copied web files to $INSTALL_DIR/web/"
else
    echo -e "${YELLOW}Warning: Web directory not found, skipping...${NC}"
fi

echo -e "\n${GREEN}[5/7] Installing systemd service...${NC}"
cat > /etc/systemd/system/xpanel.service << 'EOF'
[Unit]
Description=xPanel - VPN User Management Panel
Documentation=https://github.com/yourorg/xpanel
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

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=xpanel

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/xpanel /var/log/xpanel

[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
echo "Installed: /etc/systemd/system/xpanel.service"

echo -e "\n${GREEN}[6/7] Creating configuration...${NC}"
if [ ! -f "$CONFIG_DIR/.env" ]; then
    cat > "$CONFIG_DIR/.env" << 'EOF'
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=release

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=xpanel
DB_PASSWORD=CHANGE_ME_SECURE_PASSWORD
DB_NAME=xpanel
DB_SSLMODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration (auto-generated on first startup)
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168

# Note: NODE_API_KEY is auto-generated on first startup
# View and change it from Settings page in admin panel
EOF
    chmod 600 "$CONFIG_DIR/.env"
    chown "$USER:$GROUP" "$CONFIG_DIR/.env"
    echo -e "${YELLOW}IMPORTANT: Edit $CONFIG_DIR/.env with your settings!${NC}"
else
    echo "Configuration already exists, skipping..."
fi

echo -e "\n${GREEN}[7/7] Post-installation steps...${NC}"

if [ "$SKIP_DB" = false ]; then
    echo ""
    echo -e "${YELLOW}Database Setup Required:${NC}"
    echo "  1. Install PostgreSQL: apt install postgresql"
    echo "  2. Install Redis: apt install redis-server"
    echo "  3. Create database:"
    echo "     sudo -u postgres createuser xpanel"
    echo "     sudo -u postgres createdb -O xpanel xpanel"
    echo "     sudo -u postgres psql -c \"ALTER USER xpanel PASSWORD 'your_password';\""
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Installation Complete!               ${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Next steps:"
echo "  1. Edit configuration: nano $CONFIG_DIR/.env"
echo "  2. Start service: systemctl start xpanel"
echo "  3. Enable on boot: systemctl enable xpanel"
echo "  4. Check logs: journalctl -u xpanel -f"
echo ""
echo "Default admin login:"
echo "  Email: admin@xpanel.local"
echo "  Password: admin123"
echo ""
echo -e "${RED}IMPORTANT: Change the default password immediately!${NC}"

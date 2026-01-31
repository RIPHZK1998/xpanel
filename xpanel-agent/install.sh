#!/bin/bash
# Installation script for xpanel-agent

set -e

echo "Installing xpanel-agent..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo "Please run as root"
  exit 1
fi

# Copy binary
echo "Installing binary..."
cp xpanel-agent /usr/local/bin/
chmod +x /usr/local/bin/xpanel-agent

# Create config directory
echo "Creating config directory..."
mkdir -p /etc/xpanel-agent

# Copy config if it doesn't exist
if [ ! -f /etc/xpanel-agent/config.yaml ]; then
  echo "Installing default config..."
  cp config.yaml.example /etc/xpanel-agent/config.yaml
  echo "Please edit /etc/xpanel-agent/config.yaml with your settings"
fi

# Install systemd service
echo "Installing systemd service..."
cp xpanel-agent.service /etc/systemd/system/
systemctl daemon-reload

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "1. Edit /etc/xpanel-agent/config.yaml"
echo "2. systemctl enable xpanel-agent"
echo "3. systemctl start xpanel-agent"
echo "4. systemctl status xpanel-agent"

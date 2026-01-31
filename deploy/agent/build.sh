#!/bin/bash
#
# Build script for xPanel Agent
# Produces cross-compiled binaries for deployment
#
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")/xpanel-agent"
OUTPUT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")/dist"

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Building xPanel Agent...${NC}"

cd "$PROJECT_ROOT"

# Create output directory
mkdir -p "$OUTPUT_DIR/agent"

# Build for Linux amd64
echo "Building for linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$OUTPUT_DIR/agent/xpanel-agent" .

# Build for Linux arm64 (for ARM servers)
echo "Building for linux/arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "$OUTPUT_DIR/agent/xpanel-agent-arm64" .

# Copy deployment files
cp "$(dirname "$SCRIPT_DIR")/agent/install.sh" "$OUTPUT_DIR/agent/"
chmod +x "$OUTPUT_DIR/agent/install.sh"

# Copy example config
cp config.yaml.example "$OUTPUT_DIR/agent/" 2>/dev/null || true

echo -e "${GREEN}Build complete!${NC}"
echo "Output:"
echo "  - $OUTPUT_DIR/agent/xpanel-agent (amd64)"
echo "  - $OUTPUT_DIR/agent/xpanel-agent-arm64 (arm64)"
echo ""
echo "To deploy:"
echo "  1. Copy files to your VPS node"
echo "  2. Run: sudo ./install.sh --binary ./xpanel-agent --node-id <ID> --panel-url <URL> --api-key <KEY>"

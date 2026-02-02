#!/bin/bash
#
# Build script for xPanel
# Produces cross-compiled binaries for deployment
#
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="${PROJECT_ROOT}/dist"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Building xPanel...${NC}"

cd "$PROJECT_ROOT"

# Create output directory
mkdir -p "$OUTPUT_DIR/panel"

# Build for Linux amd64 (most common server architecture)
echo "Building for linux/amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$OUTPUT_DIR/panel/xpanel" .

# Copy web files
echo "Copying web files..."
cp -r web "$OUTPUT_DIR/panel/"

# Copy deployment files
cp deploy/panel/install.sh "$OUTPUT_DIR/panel/"
chmod +x "$OUTPUT_DIR/panel/install.sh"

# Create tarball
echo "Creating archive..."
cd "$OUTPUT_DIR"
tar -czf xpanel-linux-amd64.tar.gz ls


echo -e "${GREEN}Build complete!${NC}"
echo "Output: $OUTPUT_DIR/xpanel-linux-amd64.tar.gz"
echo ""
echo "To deploy:"
echo "  1. Copy xpanel-linux-amd64.tar.gz to your server"
echo "  2. Extract: tar -xzf xpanel-linux-amd64.tar.gz"
echo "  3. Run: sudo ./panel/install.sh --binary ./panel/xpanel"

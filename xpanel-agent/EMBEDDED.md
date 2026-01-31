# xpanel-agent with Embedded xray-core

Complete guide for using xray-core as an embedded library instead of external binary.

## Benefits of Embedded xray-core

✅ **Single Binary** - No need to install xray-core separately
✅ **Better Control** - Direct API access to xray-core internals
✅ **Easier Deployment** - Just copy one binary file
✅ **Automatic Updates** - Update xray-core by updating the agent
✅ **Simpler Configuration** - No binary_path needed

## Changes from External Binary Version

### 1. Removed Binary Dependency

**Before (External Binary):**
```yaml
xray:
  binary_path: "/usr/local/bin/xray"  # ❌ Not needed
  config_path: "/etc/xpanel-agent/xray-config.json"
```

**After (Embedded Library):**
```yaml
xray:
  config_path: "/etc/xpanel-agent/xray-config.json"  # Optional for debugging
  api_address: "127.0.0.1"
  api_port: 10085
```

### 2. Manager Implementation

**Before:** Used `os/exec` to start xray as external process
**After:** Uses `github.com/xtls/xray-core/core` library directly

```go
// Embedded version
instance, err := core.New(xrayConfig)
instance.Start()
```

### 3. Build Size

- **External binary version:** ~9MB
- **Embedded library version:** ~30MB (includes full xray-core)

## Installation

### Build

```bash
cd xpanel-agent
go mod tidy
go build -o xpanel-agent-embedded main.go
```

### Deploy

```bash
# Copy single binary
sudo cp xpanel-agent-embedded /usr/local/bin/xpanel-agent

# Configure
sudo mkdir -p /etc/xpanel-agent
sudo cp config.yaml.example /etc/xpanel-agent/config.yaml

# Edit config (no binary_path needed!)
sudo nano /etc/xpanel-agent/config.yaml

# Install systemd service
sudo cp xpanel-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable xpanel-agent
sudo systemctl start xpanel-agent
```

## Configuration

Minimal configuration needed:

```yaml
node:
  id: 1
  name: "US-West-1"

panel:
  url: "https://panel.example.com"
  api_key: "your-api-key"

xray:
  # No binary_path needed!
  api_address: "127.0.0.1"
  api_port: 10085
  protocol: "vless"
  listen_port: 443
  tls_enabled: true
  cert_file: "/etc/ssl/certs/server.crt"
  key_file: "/etc/ssl/private/server.key"
```

## Advantages

1. **Simplified Deployment**
   - One binary contains everything
   - No xray-core installation required
   - Easier to distribute

2. **Version Control**
   - xray-core version locked to agent version
   - Consistent behavior across nodes
   - Easier rollbacks

3. **Better Integration**
   - Direct access to xray internals
   - No process management overhead
   - Faster startup

4. **Reduced Dependencies**
   - No external binary dependencies
   - Fewer moving parts
   - Less can go wrong

## Comparison

| Feature | External Binary | Embedded Library |
|---------|----------------|------------------|
| Binary Size | ~9MB | ~30MB |
| xray-core Install | Required | Not Required |
| Deployment | 2 binaries | 1 binary |
| Version Control | Separate | Unified |
| Startup Time | Slower | Faster |
| Memory Usage | Similar | Similar |
| Configuration | More complex | Simpler |

## Recommendation

**Use Embedded Library** for:
- New deployments
- Simplified operations
- Easier updates
- Single binary preference

**Use External Binary** if:
- You need latest xray-core immediately
- You want independent xray updates
- Disk space is critical

## Technical Details

The embedded version uses:
- `github.com/xtls/xray-core/core` - Core xray functionality
- `github.com/xtls/xray-core/infra/conf` - Configuration parsing
- All xray protocols (VLESS, VMess, Trojan, etc.)
- Full stats and API support

## Migration from External to Embedded

1. Stop current agent
2. Replace binary with embedded version
3. Update config.yaml (remove `binary_path`)
4. Restart agent

```bash
sudo systemctl stop xpanel-agent
sudo cp xpanel-agent-embedded /usr/local/bin/xpanel-agent
sudo nano /etc/xpanel-agent/config.yaml  # Remove binary_path line
sudo systemctl start xpanel-agent
```

Done! The agent now runs with embedded xray-core.

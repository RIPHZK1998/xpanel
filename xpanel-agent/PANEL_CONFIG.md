# Panel-Driven vs Local Configuration

## Overview

The node agent can get its configuration from two sources:

1. **Panel-Driven (Recommended)** - Fetch from management panel
2. **Local Configuration** - Read from `config.yaml`

## How It Works

### Panel-Driven Configuration (Recommended)

The agent fetches node settings from the panel on startup:

```
Agent Startup
    ↓
Fetch config from panel: GET /api/v1/node-agent/{node_id}/config
    ↓
Panel returns: protocol, port, TLS settings, etc.
    ↓
Agent uses panel settings to configure xray-core
    ↓
Falls back to local config if panel unreachable
```

**Advantages:**
- ✅ Centralized management
- ✅ Change settings without SSH to nodes
- ✅ Consistent configuration across nodes
- ✅ Easy protocol/port updates

**Configuration Flow:**
1. Admin creates/updates node in panel
2. Panel stores: protocol, port, TLS, SNI, etc.
3. Agent fetches config on startup
4. Agent applies settings to xray-core

### Local Configuration (Fallback)

If panel is unreachable, agent uses `config.yaml`:

```yaml
xray:
  protocol: "vless"
  listen_port: 443
  tls_enabled: true
  sni: "vpn.example.com"
```

## Configuration Priority

```
1. Panel Configuration (if available)
   ↓ (fallback if panel unreachable)
2. Local config.yaml
```

## What's Configured Where

### Managed by Panel

These settings are fetched from panel:

- ✅ `protocol` (vless/vmess/trojan)
- ✅ `port` (VPN listen port)
- ✅ `tls_enabled` (TLS on/off)
- ✅ `sni` (Server Name Indication)
- ✅ `inbound_tag` (xray inbound tag)
- ✅ `api_port` (xray API port)

### Always Local

These settings remain in `config.yaml`:

- `node.id` - Node identifier
- `panel.url` - Panel URL
- `panel.api_key` - Authentication
- `xray.cert_file` - TLS certificate path
- `xray.key_file` - TLS key path
- `intervals.*` - Sync intervals

## Example Workflow

### 1. Admin Creates Node in Panel

```bash
POST /api/v1/admin/nodes
{
  "name": "US-West-1",
  "address": "vpn1.example.com",
  "port": 443,
  "protocol": "vless",
  "tls_enabled": true,
  "sni": "vpn1.example.com"
}
```

### 2. Deploy Agent on Server

```yaml
# config.yaml (minimal)
node:
  id: 1  # From panel

panel:
  url: "https://panel.example.com"
  api_key: "secret-key"

xray:
  # Protocol/port/TLS will be fetched from panel!
  cert_file: "/etc/ssl/certs/server.crt"
  key_file: "/etc/ssl/private/server.key"
```

### 3. Agent Starts

```
[INFO] Starting xpanel-agent...
[INFO] Fetching node configuration from panel...
[INFO] Using panel configuration: protocol=vless, port=443, tls=true
[INFO] Starting embedded xray-core...
[INFO] xpanel-agent started successfully
```

### 4. Change Protocol (No SSH Needed!)

```bash
# Admin updates in panel
PUT /api/v1/admin/nodes/1
{
  "protocol": "vmess",  # Changed from vless
  "port": 8443          # Changed port
}

# Restart agent (or wait for auto-reload)
systemctl restart xpanel-agent

# Agent fetches new config automatically!
```

## Benefits

### Centralized Management
- Change all nodes from panel
- No SSH to individual servers
- Consistent configuration

### Easy Updates
- Switch protocols without editing files
- Change ports dynamically
- Enable/disable TLS from panel

### Disaster Recovery
- Node config backed up in database
- Easy to recreate nodes
- Configuration history

## Implementation Details

### Panel Endpoint

```go
GET /api/v1/node-agent/:node_id/config

Response:
{
  "success": true,
  "data": {
    "node": {
      "id": 1,
      "protocol": "vless",
      "port": 443,
      "tls_enabled": true,
      "sni": "vpn.example.com",
      "inbound_tag": "proxy"
    }
  }
}
```

### Agent Startup Code

```go
// Fetch config from panel
nodeConfig, err := panelClient.FetchNodeConfig(nodeID)
if err != nil {
    logger.Warn("Using local config.yaml")
} else {
    // Use panel config
    xrayConfig.Protocol = nodeConfig.Protocol
    xrayConfig.Port = nodeConfig.Port
    xrayConfig.TLSEnabled = nodeConfig.TLSEnabled
}
```

## Migration Guide

### From Local to Panel-Driven

1. **Create node in panel** with current settings
2. **Simplify config.yaml** - remove protocol/port/TLS
3. **Restart agent** - it will fetch from panel

### Minimal config.yaml

```yaml
node:
  id: 1

panel:
  url: "https://panel.example.com"
  api_key: "your-key"

xray:
  cert_file: "/etc/ssl/certs/server.crt"
  key_file: "/etc/ssl/private/server.key"
  
# Protocol, port, TLS fetched from panel!
```

## Best Practices

1. **Use Panel Configuration** for production
2. **Keep Local Config** as fallback
3. **Document Node Settings** in panel
4. **Test Changes** on one node first
5. **Monitor Agent Logs** for config fetch errors

## Troubleshooting

### Agent uses local config instead of panel

**Check:**
- Panel URL is correct
- API key is valid
- Node ID exists in panel
- Network connectivity to panel

### Config changes not applied

**Solution:**
- Restart agent: `systemctl restart xpanel-agent`
- Check agent logs for errors
- Verify panel config is correct

## Summary

✅ **Panel-driven configuration** is the recommended approach
✅ **Local config** serves as fallback
✅ **Centralized management** makes operations easier
✅ **No SSH needed** to change node settings

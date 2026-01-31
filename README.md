# xpanel - VPN User Management Backend

A production-ready Golang backend service for VPN user management with xray-core integration.

## Features

- **User Management**: Registration, authentication, JWT-based auth with refresh tokens
- **Subscription Plans**: Free, Monthly, and Yearly plans with data limits
- **Multi-Node Support**: Manage multiple xray-core VPN nodes
- **Traffic Tracking**: Per-user, per-node traffic statistics
- **Device Management**: Track and manage user devices/sessions
- **Security**: bcrypt password hashing, JWT tokens, Redis token blacklist, rate limiting
- **Clean Architecture**: Layered design with clear separation of concerns

## Tech Stack

- **Language**: Go 1.25+
- **Framework**: Gin
- **Database**: PostgreSQL (GORM)
- **Cache**: Redis
- **Authentication**: JWT
- **VPN Protocol**: xray-core (VLESS/VMess/Trojan)

## Project Structure

```
xpanel/
├── main.go                    # Application entry point
├── config/                    # Configuration management
├── pkg/                       # Shared utilities
│   ├── jwt/                   # JWT token handling
│   └── response/              # API response helpers
├── internal/
│   ├── models/                # Database models (GORM)
│   ├── repository/            # Data access layer
│   ├── service/               # Business logic
│   ├── handler/               # HTTP handlers
│   ├── middleware/            # HTTP middleware
│   └── xray/                  # xray-core integration
├── .env.example               # Environment variables template
└── schema.sql                 # Database schema
```

## Prerequisites

- Go 1.25 or higher
- PostgreSQL 12+
- Redis 6+
- (Optional) xray-core nodes for actual VPN functionality

## Installation

1. **Clone the repository**
   ```bash
   cd /Users/han/Documents/go-work-dir/wpn
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Setup database**
   ```bash
   # Create PostgreSQL database
   createdb xpanel

   # Run schema
   psql -d xpanel -f schema.sql
   ```

5. **Start Redis**
   ```bash
   redis-server
   ```

## Running the Service

```bash
# Development mode
go run main.go

# Build binary
go build -o xpanel main.go

# Run binary
./xpanel
```

The service will start on `http://localhost:8080` (configurable via `.env`).

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout (requires auth)

### User

- `GET /api/v1/user/profile` - Get user profile
- `GET /api/v1/user/devices` - List user devices
- `DELETE /api/v1/user/devices/:id` - Deactivate device
- `GET /api/v1/user/subscription` - Get subscription details
- `GET /api/v1/user/config` - Get VPN client configuration

### Subscription

- `POST /api/v1/subscription/renew` - Renew/upgrade subscription

### Nodes

- `GET /api/v1/nodes` - List available VPN nodes

### Health Check

- `GET /health` - Service health check

## Configuration

Edit `.env` file with your settings:

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_MODE=debug

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=xpanel
DB_SSLMODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-super-secret-key-min-32-chars
JWT_ACCESS_TTL_MINUTES=15
JWT_REFRESH_TTL_HOURS=168
```

## Example Usage

### Register a new user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword123"
  }'
```

### Login



```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@xpanel.local",
    "password": "admin123"
  }'
```

### Get user profile (with JWT token)

```bash
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Get VPN configuration

```bash
curl -X GET http://localhost:8080/api/v1/user/config \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## xray-core Integration

This backend is designed to work with xray-core nodes. To use the VPN functionality:

1. Setup xray-core on your server nodes
2. Enable xray API (typically on port 10085)
3. Add nodes to the database:

```sql
INSERT INTO nodes (name, address, port, protocol, status, api_endpoint, api_port, tls_enabled, sni, inbound_tag)
VALUES ('US-West-1', 'vpn.example.com', 443, 'vless', 'online', 'vpn.example.com', 10085, true, 'vpn.example.com', 'proxy');
```

4. The backend will automatically provision users to registered nodes

## Security Considerations

- Change `JWT_SECRET` to a strong random string (min 32 characters)
- Use `SERVER_MODE=release` in production
- Enable PostgreSQL SSL (`DB_SSLMODE=require`) in production
- Set Redis password in production
- Run behind Nginx or similar reverse proxy
- Use HTTPS in production
- Implement additional rate limiting at nginx level
- Regular security audits and updates

## License

This is a production-ready template. Adjust licensing as needed.

## Support

For issues or questions, please file an issue in the repository.

# EdgeSync

A secure multi-device sync service built with Go. EdgeSync enables real-time synchronization of encrypted data across multiple devices with presence awareness and conflict resolution.

## Tech Stack

- **Go 1.21+** - Backend runtime
- **PostgreSQL 15** - Primary database
- **Redis 7** - Caching and presence
- **Chi Router** - HTTP routing
- **pgx** - PostgreSQL driver
- **golang-jwt** - JWT authentication

## Project Structure

```
EdgeSync/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point (TODO)
├── internal/
│   ├── config/
│   │   └── config.go            # Environment config (TODO)
│   ├── database/
│   │   ├── postgres.go          # Postgres connection (TODO)
│   │   └── redis.go             # Redis connection (TODO)
│   └── models/
│       ├── account.go           # User account model
│       ├── device.go            # Device model
│       ├── encrypted_state.go   # Encrypted state model
│       └── sync_event.go        # Sync event model
├── migrations/
│   ├── 000001_create_accounts.up.sql
│   ├── 000001_create_accounts.down.sql
│   ├── 000002_create_devices.up.sql
│   ├── 000002_create_devices.down.sql
│   ├── 000003_create_encrypted_states.up.sql
│   ├── 000003_create_encrypted_states.down.sql
│   ├── 000004_create_sync_events.up.sql
│   └── 000004_create_sync_events.down.sql
├── docker-compose.yaml
├── go.mod
├── go.sum
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- golang-migrate CLI

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/prudhvinik1/edgesync.git
   cd edgesync
   ```

2. **Start infrastructure**
   ```bash
   docker-compose up -d
   ```

3. **Verify services are healthy**
   ```bash
   docker-compose ps
   ```

4. **Install migrate CLI** (macOS)
   ```bash
   brew install golang-migrate
   ```

5. **Run database migrations**
   ```bash
   migrate -path ./migrations -database 'postgres://postgres:postgres@localhost:5432/edgesync?sslmode=disable' up
   ```

6. **Install Go dependencies**
   ```bash
   go mod download
   ```

## Database Schema

### accounts
Stores user account information.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) | Unique email address |
| password_hash | VARCHAR(255) | Bcrypt hashed password |
| created_at | TIMESTAMPTZ | Account creation time |
| updated_at | TIMESTAMPTZ | Last update time |

### devices
Registered devices for each account.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| account_id | UUID | Foreign key to accounts |
| name | VARCHAR(255) | Device name |
| device_type | VARCHAR(50) | Type (mobile, desktop, browser) |
| public_key | TEXT | E2E encryption public key |
| last_seen_at | TIMESTAMPTZ | Last presence update |
| revoked_at | TIMESTAMPTZ | When device was revoked |

### encrypted_states
Stores encrypted user data blobs.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| account_id | UUID | Foreign key to accounts |
| key | VARCHAR(255) | State identifier (e.g., "settings") |
| state | BYTEA | Encrypted data blob |
| nonce | BYTEA | Encryption nonce |
| version | BIGINT | Optimistic locking version |

### sync_events
Event log for tracking all sync operations.

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| account_id | UUID | Foreign key to accounts |
| device_id | UUID | Which device triggered event |
| event_type | VARCHAR(50) | create, update, delete |
| state_key | VARCHAR(255) | Affected state key |
| sequence_num | BIGSERIAL | Ordered sequence number |

## Development

### Useful Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Connect to Postgres
docker exec -it edgesync-postgres-1 psql -U postgres -d edgesync

# Migration commands
migrate -path ./migrations -database '...' up      # Apply all
migrate -path ./migrations -database '...' down 1  # Rollback one
migrate -path ./migrations -database '...' version # Check version
```

## License

MIT


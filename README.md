# EdgeSync

A secure multi-device sync service built with Go. EdgeSync enables real-time synchronization of encrypted data across multiple devices with presence awareness and conflict resolution.

## Tech Stack

- **Go 1.21+** - Backend runtime
- **PostgreSQL 15** - Primary database
- **Redis 7** - Session management & presence
- **Chi Router** - HTTP routing
- **pgx** - PostgreSQL driver
- **go-redis** - Redis client
- **golang-jwt** - JWT authentication

## Project Structure

```
EdgeSync/
├── cmd/
│   └── server/
│       └── main.go                  # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go                # Environment configuration
│   ├── database/
│   │   ├── postgres.go              # Postgres connection pool
│   │   └── redis.go                 # Redis client
│   ├── models/
│   │   ├── account.go               # User account
│   │   ├── device.go                # Device
│   │   ├── encrypted_state.go       # Encrypted state blob
│   │   ├── session.go               # JWT session
│   │   ├── sync_event.go            # Sync event log
│   │   └── presence.go              # Device presence
│   └── repositories/
│       ├── interfaces.go            # Repository interfaces
│       ├── account_repo.go          # Account CRUD (Postgres)
│       ├── device_repo.go           # Device CRUD (Postgres)
│       └── session_repo.go          # Session management (Redis)
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
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

## Architecture

### Repository Pattern

EdgeSync uses the repository pattern to separate data access from business logic:

```
HTTP Handlers → Services → Repository Interfaces → Implementations
                                    ↓
                          ┌─────────┴─────────┐
                          │                   │
                    PostgresRepo          RedisRepo
                          │                   │
                          ▼                   ▼
                      PostgreSQL           Redis
```

### Session Storage with Redis

Sessions are stored in Redis with automatic TTL expiration. To enable querying sessions by account, we maintain a secondary index:

```
session:{id}                    → Session JSON (with TTL)
account:{accountID}:sessions    → Set of session IDs
```

**Design Decision:** We use lazy cleanup for expired session references in the secondary index. When listing sessions for an account, we check if each session still exists and remove stale references. This approach is simple, requires no background jobs, and handles the eventual consistency naturally.

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

6. **Create .env file**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

7. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

8. **Test health endpoint**
   ```bash
   curl http://localhost:8080/health
   # Returns: OK
   ```

## Database Schema

### accounts
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| email | VARCHAR(255) | Unique email address |
| password_hash | VARCHAR(255) | Bcrypt hashed password |
| created_at | TIMESTAMPTZ | Account creation time |
| updated_at | TIMESTAMPTZ | Last update time |
| deleted_at | TIMESTAMPTZ | Soft delete timestamp |

### devices
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
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| account_id | UUID | Foreign key to accounts |
| key | VARCHAR(255) | State identifier (e.g., "settings") |
| state | BYTEA | Encrypted data blob |
| nonce | BYTEA | Encryption nonce |
| version | BIGINT | Optimistic locking version |

### sync_events
| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| account_id | UUID | Foreign key to accounts |
| device_id | UUID | Which device triggered event |
| event_type | VARCHAR(50) | create, update, delete |
| state_key | VARCHAR(255) | Affected state key |
| sequence_num | BIGSERIAL | Ordered sequence number |

## Development Progress

### Phase 1: Project Setup ✅
- [x] Go module with Chi router
- [x] Docker Compose (Postgres 15 + Redis 7)
- [x] Database migrations
- [x] Environment configuration
- [x] Health check endpoint

### Phase 2: Models & Repositories (In Progress)
- [x] All domain models
- [x] Repository interfaces
- [x] Account repository (Postgres)
- [x] Device repository (Postgres)
- [x] Session repository (Redis)
- [ ] State repository (Postgres)
- [ ] SyncEvent repository (Postgres)
- [ ] Presence repository (Redis)

### Phase 3-8: Coming Soon
- Authentication & device management
- Encrypted state sync with optimistic locking
- Presence system
- WebSocket & fan-out
- Demo frontend
- Docker & documentation

## Useful Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f

# Connect to Postgres
docker exec -it edgesync-postgres-1 psql -U postgres -d edgesync

# Connect to Redis
docker exec -it edgesync-redis-1 redis-cli

# Migration commands
migrate -path ./migrations -database '...' up        # Apply all
migrate -path ./migrations -database '...' down 1    # Rollback one
migrate -path ./migrations -database '...' version   # Check version
migrate -path ./migrations -database '...' force N   # Force version
```

## License

MIT

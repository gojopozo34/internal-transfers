# Internal Transfers API

A production-ready Go REST API for managing bank account transfers with ACID compliance and row-level locking to ensure data consistency.

## Quick Start

```powershell
# 1. Clone repository
git clone <your-repo-url>
cd internal-transfers

# 2. Install dependencies
go mod tidy

# 3. Setup database (see Setup section below)

# 4. Start server
go run ./cmd/server/main.go

# 5. Run tests (in another terminal)
& ".\test_simple.ps1"
```

This application implements a secure, transactional system for transferring money between accounts. It uses PostgreSQL with serializable transaction isolation and row-level locking (FOR UPDATE) to prevent race conditions and ensure data integrity even under concurrent load.

## Assumptions

This project makes the following assumptions about the deployment environment and usage:

### Environment Assumptions
1. **Operating System**: Windows 10/11, Linux, or macOS with standard shells
2. **PostgreSQL Version**: PostgreSQL 18.x or later (tested with 18.1)
3. **Go Version**: Go 1.24.1 or later
4. **Network**: Localhost deployment (127.0.0.1:8080) for development; production requires HTTPS
5. **Database Access**: Local PostgreSQL instance accessible via TCP/IP

### Security Assumptions
1. **No Authentication**: API endpoints are unauthenticated (add OAuth2/API keys for production)
2. **No Authorization**: All users have full access to all accounts (implement RBAC for production)
3. **Plain TCP Connection**: Database connection uses no encryption (enable SSL/TLS for production)
4. **Single Server**: No load balancing or horizontal scaling (use connection pooling for high concurrency)
5. **Default Credentials**: PostgreSQL user `postgres` with password `password` (use secrets manager in production)

### Data Assumptions
1. **Account IDs**: Strings (VARCHAR), assumed to be unique within system
2. **Balances**: Decimal(18,2), non-negative, no currency specification (single currency assumed)
3. **Transfer Amounts**: Positive decimals only, no partial transactions
4. **Transaction History**: Immutable (no transaction cancellations or reversals)
5. **Audit Logs**: Permanent (logs are never deleted)

### Concurrency Assumptions
1. **Row-Level Locking**: PostgreSQL FOR UPDATE prevents lost updates
2. **Serializable Isolation**: No dirty reads, phantom reads, or non-repeatable reads
3. **Single Process**: Application runs as single Go process (not distributed)
4. **Connection Pool**: Max 25 concurrent connections to PostgreSQL
5. **No Eventual Consistency**: All reads see committed state (strong consistency)

### Operational Assumptions
1. **Logging**: All requests and errors logged to stdout as JSON
2. **Monitoring**: No built-in metrics (integrate Prometheus/Grafana separately)
3. **Graceful Shutdown**: Server handles SIGINT/SIGTERM for clean exit
4. **Health Checks**: Basic `/health` endpoint for liveness checks
5. **No Caching**: Each request queries database directly (add Redis for high-traffic scenarios)

## Key Features

- **Account Management**: Create and retrieve bank accounts with persistent balance tracking
- **Money Transfers**: Atomic transfers between accounts with strict validation
- **Transaction Safety**: SERIALIZABLE isolation level with row-level locking
- **Audit Logging**: Complete audit trail of all account and transaction events
- **Error Handling**: Comprehensive validation and error responses
- **REST API**: Clean HTTP endpoints with JSON payloads
- **Database Persistence**: PostgreSQL backend with proper constraints and indexing

## Architecture

### Technology Stack

- **Language**: Go 1.24.1
- **Framework**: Gorilla Mux (routing)
- **Database**: PostgreSQL 18
- **Logging**: Structured JSON logging with slog

### Core Components

1. **Handlers** (`internal/handler/`)
   - `AccountHandler`: Manages account endpoints
   - `TransactionHandler`: Manages transfer endpoints

2. **Services** (`internal/service/`)
   - `AccountService`: Business logic for account operations
   - `TransactionService`: Orchestrates transfers with transaction management

3. **Repositories** (`internal/repository/`)
   - `AccountRepository`: Account data access with row-level locking
   - `TransactionRepository`: Transaction record persistence
   - `AuditRepository`: Audit log storage

4. **Models** (`internal/models/`)
   - `Account`: Bank account entity
   - `Transaction`: Transfer record
   - `AuditLog`: Event tracking
   - Request/Response DTOs

## Row-Level Locking & Concurrency Control

### Problem: Race Conditions

Without proper locking, concurrent transfers can cause data inconsistency:
```
Thread 1: Read acc1 balance = $1000
Thread 2: Read acc1 balance = $1000
Thread 1: Deduct $500, Write $500
Thread 2: Deduct $300, Write $700  ← WRONG! Should be $200
```

### Solution: PostgreSQL Row-Level Locks

**FOR UPDATE Clause**:
```sql
SELECT id, balance FROM accounts WHERE id = $1 FOR UPDATE
```

This acquires an exclusive row lock that:
- Prevents other transactions from reading or modifying the row during the lock lifetime
- Blocks until other locks are released (preventing dirty reads)
- Ensures atomic read-modify-write operations

**Transaction Flow**:
1. Begin transaction with SERIALIZABLE isolation
2. Lock source account with FOR UPDATE
3. Lock destination account with FOR UPDATE
4. Verify sufficient balance
5. Update both account balances
6. Record transaction
7. Create audit logs
8. Commit (locks released automatically)

This guarantees:
- **Atomicity**: All-or-nothing transfer
- **Consistency**: Balances remain accurate
- **Isolation**: No dirty reads or phantom reads
- **Durability**: Persisted to PostgreSQL

## API Endpoints

### Accounts

#### Create Account
```
POST /accounts
Content-Type: application/json

{
  "id": "acc001",
  "initial_balance": 1000.00
}

Response (201):
{
  "id": "acc001",
  "balance": 1000.00
}
```

**Validation**:
- Account ID must be non-empty
- Initial balance must be >= 0
- Account ID must be unique (409 Conflict if duplicate)

#### Get Account
```
GET /accounts/{id}

Response (200):
{
  "id": "acc001",
  "balance": 1000.00
}
```

**Errors**:
- 404 Not Found if account doesn't exist
- 400 Bad Request if ID is empty

### Transactions

#### Create Transfer
```
POST /transactions
Content-Type: application/json

{
  "source_account_id": "acc001",
  "destination_account_id": "acc002",
  "amount": 250.00
}

Response (201):
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "source_account_id": "acc001",
  "destination_account_id": "acc002",
  "amount": 250.00,
  "created_at": "2025-11-30T18:11:43.156635Z"
}
```

**Validation**:
- Source and destination must be different (400 Bad Request)
- Amount must be > 0 (400 Bad Request)
- Source account must exist (404 Not Found)
- Destination account must exist (404 Not Found)
- Source account must have sufficient balance (400 Bad Request)

**Atomicity Guarantees**:
- Both accounts are updated or neither is updated
- Transaction record is created only if both updates succeed
- All changes are persisted or rolled back as a unit

## Prerequisites

### System Requirements
- **OS**: Windows 10/11, macOS 10.14+, or Linux (Ubuntu 20.04+)
- **Go**: Version 1.24.1 or later ([Download](https://golang.org/dl/))
- **PostgreSQL**: Version 18.x or later ([Download](https://www.postgresql.org/download/))
- **Terminal**: PowerShell (Windows), Bash (macOS/Linux), or equivalent shell
- **Memory**: Minimum 2GB RAM
- **Disk**: Minimum 500MB free space

### Installation Steps

#### 1. Install Go

**Windows**:
- Download installer from [golang.org](https://golang.org/dl/)
- Run installer, accept defaults
- Verify: `go version`

**macOS** (Homebrew):
```bash
brew install go
go version
```

**Linux** (Ubuntu):
```bash
sudo apt-get update
sudo apt-get install golang-go
go version
```

#### 2. Install PostgreSQL

**Windows**:
- Download from [postgresql.org](https://www.postgresql.org/download/windows/)
- Run installer
- **Important**: Remember the postgres user password you set during installation
- Verify: `psql --version`
- Add to PATH if needed: `$env:PATH += ";C:\Program Files\PostgreSQL\18\bin"`

**macOS** (Homebrew):
```bash
brew install postgresql
brew services start postgresql
createdb transfers
psql postgres -c "ALTER USER postgres WITH PASSWORD 'password';"
```

**Linux** (Ubuntu):
```bash
sudo apt-get update
sudo apt-get install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'password';"
```

#### 3. Clone Repository

```powershell
# Windows/PowerShell
git clone <your-repo-url>
cd internal-transfers

# macOS/Linux
git clone <your-repo-url>
cd internal-transfers
```

## Setup Instructions

### Step 1: Create Database

**Windows** (PowerShell):
```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres postgres
```

In the psql prompt, execute:
```sql
CREATE DATABASE transfers;
ALTER USER postgres WITH PASSWORD 'password';
\q
```

**macOS/Linux** (Bash):
```bash
sudo -u postgres psql

-- In psql:
CREATE DATABASE transfers;
ALTER USER postgres WITH PASSWORD 'password';
\q
```

### Step 2: Run Database Migrations

This creates all required tables, constraints, and indexes.

**Windows** (PowerShell):
```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres -h localhost -d transfers -f "C:\path\to\internal-transfers\db\migrations\001_init.sql"
```

**macOS/Linux** (Bash):
```bash
psql -U postgres -h localhost -d transfers -f ./db/migrations/001_init.sql
```

Verify tables were created:
```bash
psql -U postgres -h localhost -d transfers -c "\dt"
```

Expected output:
```
               List of relations
 Schema |    Name    | Type  |  Owner
--------+------------+-------+----------
 public | accounts   | table | postgres
 public | audit_logs | table | postgres
 public | transactions | table | postgres
(3 rows)
```

### Step 3: Install Go Dependencies

```bash
cd internal-transfers
go mod tidy
```

Downloads:
- `github.com/gorilla/mux` v1.8.1
- `github.com/lib/pq` v1.10.9
- `github.com/google/uuid` v1.6.0

### Step 4: Run the Server

**Windows** (PowerShell):
```powershell
cd c:\internal-transfers
go run ./cmd/server/main.go
```

**macOS/Linux** (Bash):
```bash
cd internal-transfers
go run ./cmd/server/main.go
```

Expected output:
```json
{"time":"2025-11-30T18:09:19.867Z","level":"INFO","msg":"connected to database successfully"}
{"time":"2025-11-30T18:09:19.870Z","level":"INFO","msg":"starting server on port 8080"}
```

The server is now listening on `http://localhost:8080`

### Step 5: Run Tests (Optional)

**Windows** (PowerShell - New Terminal):
```powershell
cd c:\internal-transfers
& ".\test_simple.ps1"
```

**macOS/Linux** (Bash - New Terminal):
```bash
cd internal-transfers
bash ./test_simple.ps1
```

### Environment Variables (Optional)

Override defaults by setting environment variables before starting the server:

**Windows** (PowerShell):
```powershell
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "password"
$env:DB_NAME = "transfers"
$env:DB_SSLMODE = "disable"
$env:SERVER_PORT = "8080"
```

**macOS/Linux** (Bash):
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=transfers
export DB_SSLMODE=disable
export SERVER_PORT=8080
```

Then start the server as usual.

## Running Tests

### Automated Test Suite

The test suite comprehensively validates all API endpoints and error scenarios.

**Windows** (PowerShell):
```powershell
# Terminal 1: Start server (if not already running)
cd c:\internal-transfers
go run ./cmd/server/main.go

# Terminal 2: Run tests
cd c:\internal-transfers
& ".\test_simple.ps1"
```

**macOS/Linux** (Bash):
```bash
# Terminal 1: Start server
cd internal-transfers
go run ./cmd/server/main.go

# Terminal 2: Run tests
cd internal-transfers
bash ./test_simple.ps1
```

### Test Coverage

The automated test suite (`test_simple.ps1`) executes 11 comprehensive tests:

| # | Test Case | Endpoint | Expected Status | Purpose |
|---|-----------|----------|-----------------|---------|
| 1 | Health Check | GET /health | 200 | Verify server is running |
| 2 | Create Account 1 | POST /accounts | 201 | Create acc001 with $1000 |
| 3 | Create Account 2 | POST /accounts | 201 | Create acc002 with $500 |
| 4 | Get Account 1 | GET /accounts/acc001 | 200 | Retrieve acc001 details |
| 5 | Get Account 2 | GET /accounts/acc002 | 200 | Retrieve acc002 details |
| 6 | Transfer Money | POST /transactions | 201 | Transfer $250 (acc001→acc002) |
| 7 | Verify Source Balance | GET /accounts/acc001 | 200 | Verify balance = $750 |
| 8 | Verify Dest Balance | GET /accounts/acc002 | 200 | Verify balance = $750 |
| 9 | Duplicate Account | POST /accounts (duplicate) | 409 | Reject duplicate account ID |
| 10 | Insufficient Balance | POST /transactions (high amount) | 400 | Reject insufficient funds |
| 11 | Same Source/Dest | POST /transactions (same account) | 400 | Reject same account transfer |

### Manual Testing

Use curl or PowerShell to test individual endpoints:

**Create Account**:
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"id":"test001","initial_balance":1000}'
```

**Get Account**:
```bash
curl http://localhost:8080/accounts/test001
```

**Transfer Money**:
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{"source_account_id":"test001","destination_account_id":"acc001","amount":500}'
```

**Health Check**:
```bash
curl http://localhost:8080/health
```

### PowerShell Examples

```powershell
# Create account
$body = @{id="ps001"; initial_balance=2000} | ConvertTo-Json
Invoke-WebRequest -Uri "http://localhost:8080/accounts" `
  -Method Post `
  -Headers @{"Content-Type"="application/json"} `
  -Body $body `
  -UseBasicParsing

# Get account
Invoke-WebRequest -Uri "http://localhost:8080/accounts/ps001" `
  -Method Get `
  -UseBasicParsing

# Transfer
$body = @{
  source_account_id="ps001"
  destination_account_id="acc001"
  amount=100
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:8080/transactions" `
  -Method Post `
  -Headers @{"Content-Type"="application/json"} `
  -Body $body `
  -UseBasicParsing
```

## Database Schema

### accounts
```sql
CREATE TABLE accounts (
    id VARCHAR(36) PRIMARY KEY,
    balance DECIMAL(18,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT balance_non_negative CHECK (balance >= 0)
);
```

### transactions
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_account_id VARCHAR(36) NOT NULL REFERENCES accounts(id),
    destination_account_id VARCHAR(36) NOT NULL REFERENCES accounts(id),
    amount DECIMAL(18,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT amount_positive CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (source_account_id != destination_account_id)
);
```

### audit_logs
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(36) NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Error Handling

### HTTP Status Codes

| Code | Scenario | Response |
|------|----------|----------|
| 200 | Success (GET) | Entity data |
| 201 | Success (POST) | Created entity |
| 400 | Validation error | `{"error":"...","message":"..."}` |
| 404 | Account not found | `{"error":"account not found","message":""}` |
| 409 | Duplicate account | `{"error":"account already exists","message":""}` |
| 500 | Server error | `{"error":"internal server error","message":""}` |

### Error Examples

**Insufficient Balance**:
```json
{
  "error": "insufficient balance",
  "message": "source account does not have enough funds for txn"
}
```

**Invalid Account ID**:
```json
{
  "error": "invalid account ID",
  "message": ""
}
```

**Same Source/Destination**:
```json
{
  "error": "same source and destination account",
  "message": "source and destination accounts cannot be the same"
}
```

## Logging

The application uses structured JSON logging for all events:

```json
{"time":"2025-11-30T18:09:19.8676473+05:30","level":"INFO","msg":"connected to database successfully"}
{"time":"2025-11-30T18:09:19.8696167+05:30","level":"INFO","msg":"starting server on port 8080"}
{"time":"2025-11-30T18:11:43.1520000+05:30","level":"INFO","msg":"incoming request","method":"POST","path":"/transactions","status":201,"duration_ms":45}
```

## Troubleshooting

### "psql not found"
Add PostgreSQL to PATH:
```powershell
$env:PATH += ";C:\Program Files\PostgreSQL\18\bin"
```

### "password authentication failed"
Ensure postgres user password is set:
```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres postgres -c "ALTER USER postgres WITH PASSWORD 'password';"
```

### "connection refused"
Verify PostgreSQL service is running:
```powershell
Get-Service postgres* | Start-Service
```

### "database does not exist"
Create the database:
```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres postgres -c "CREATE DATABASE transfers;"
```

### "tables do not exist"
Run migrations:
```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres -h localhost -d transfers -f "C:\internal-transfers\db\migrations\001_init.sql"
```

## Production Deployment

### Recommended Changes

1. **Connection Pooling**: Increase `SetMaxOpenConns` and `SetMaxIdleConns` based on load
2. **Logging Level**: Change to `LevelWarn` to reduce verbosity
3. **Database Credentials**: Use environment variables or secrets manager (never hardcode)
4. **SSL/TLS**: Set `DB_SSLMODE = "require"` and enable HTTPS
5. **Rate Limiting**: Add middleware to prevent abuse
6. **Authentication**: Add API key or OAuth2 authentication
7. **Monitoring**: Integrate with APM tools (Datadog, New Relic, etc.)

### Load Testing

Test concurrent transfers using Apache Bench:
```bash
ab -n 1000 -c 50 -p data.json -T application/json http://localhost:8080/transactions
```

## License

MIT

# Internal Transfers API

A production-ready Go REST API for managing bank account transfers with ACID compliance and row-level locking to ensure data consistency.

## Overview

This application implements a secure, transactional system for transferring money between accounts. It uses PostgreSQL with serializable transaction isolation and row-level locking (FOR UPDATE) to prevent race conditions and ensure data integrity even under concurrent load.

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
- Windows 10/11 or Linux/macOS with PostgreSQL installed
- Go 1.24.1 or later
- PostgreSQL 18.x or later

### Installation

1. **PostgreSQL**
   - Download from [postgresql.org](https://www.postgresql.org/download/)
   - During installation, set postgres user password (used later)

2. **Go**
   - Download from [golang.org](https://golang.org/dl/)

3. **This Repository**
   ```powershell
   git clone <repo-url>
   cd internal-transfers
   ```

## Setup Instructions

### 1. Create Database

Open PowerShell and connect to PostgreSQL:

```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres postgres
```

In psql prompt, run:
```sql
CREATE DATABASE transfers;
ALTER USER postgres WITH PASSWORD 'password';
\q
```

### 2. Run Migrations

```powershell
cd "C:\Program Files\PostgreSQL\18\bin"
.\psql -U postgres -h localhost -d transfers -f "C:\internal-transfers\db\migrations\001_init.sql"
```

This creates:
- `accounts` table with balance constraints
- `transactions` table with amount validation
- `audit_logs` table with JSONB fields
- Performance indexes on foreign keys and timestamps

### 3. Install Dependencies

```powershell
cd c:\internal-transfers
go mod tidy
```

Downloads:
- `github.com/gorilla/mux` - HTTP routing
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/google/uuid` - UUID generation

### 4. Start the Server

```powershell
cd c:\internal-transfers
go run ./cmd/server/main.go
```

Expected output:
```
{"time":"2025-11-30T18:09:19.8676473+05:30","level":"INFO","msg":"connected to database successfully"}
{"time":"2025-11-30T18:09:19.8696167+05:30","level":"INFO","msg":"starting server on port 8080"}
```

Server runs on `http://localhost:8080`

**Environment Variables** (optional):
```powershell
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "password"
$env:DB_NAME = "transfers"
$env:DB_SSLMODE = "disable"
$env:SERVER_PORT = "8080"
```

## Running Tests

### Automated Test Suite

Open a new PowerShell terminal and run:

```powershell
& "c:\internal-transfers\test_simple.ps1"
```

This script executes 11 comprehensive tests:

| # | Test | Expected Result |
|---|------|-----------------|
| 1 | Health Check | 200 OK |
| 2 | Create Account 1 (acc001, $1000) | 201 Created |
| 3 | Create Account 2 (acc002, $500) | 201 Created |
| 4 | Get Account 1 | 200 OK, balance $1000 |
| 5 | Get Account 2 | 200 OK, balance $500 |
| 6 | Transfer $250 (acc001→acc002) | 201 Created |
| 7 | Verify acc001 balance | 200 OK, balance $750 |
| 8 | Verify acc002 balance | 200 OK, balance $750 |
| 9 | Duplicate account (should fail) | 409 Conflict |
| 10 | Insufficient balance (should fail) | 400 Bad Request |
| 11 | Same source/dest (should fail) | 400 Bad Request |

### Manual Testing

Use PowerShell to test individual endpoints:

```powershell
# Create account
$body = @{id="test001"; initial_balance=500} | ConvertTo-Json
Invoke-WebRequest -Uri "http://localhost:8080/accounts" `
  -Method Post `
  -Headers @{"Content-Type"="application/json"} `
  -Body $body `
  -UseBasicParsing

# Get account
Invoke-WebRequest -Uri "http://localhost:8080/accounts/test001" `
  -Method Get `
  -UseBasicParsing

# Transfer money
$body = @{source_account_id="test001"; destination_account_id="acc001"; amount=100} | ConvertTo-Json
Invoke-WebRequest -Uri "http://localhost:8080/transactions" `
  -Method Post `
  -Headers @{"Content-Type"="application/json"} `
  -Body $body `
  -UseBasicParsing

# Health check
Invoke-WebRequest -Uri "http://localhost:8080/health" `
  -Method Get `
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

## Contributing

Pull requests welcome. For major changes, open an issue first.

## Support

For issues or questions, open a GitHub issue or contact the maintainers.

# GopherPay

GopherPay is a payment processing and wallet management system built in Go.  
It focuses on correctness, concurrency safety, and financial data integrity.

The system supports account lifecycle management, asynchronous wallet transfers, transaction auditing, and backpressure handling under high load.

--------

Capstone Project Submitted by – Group 2 (Pune)
  -Aditya Chavan
  -Atharv Joundal
  -Ashlesha Pal
  -Kalpesh Gajare
  -Vedant Jeughale

---------

## Overview

GopherPay is designed as a minimal but realistic fintech backend. It demonstrates:

- ACID-compliant transaction handling  
- Row-level locking to prevent race conditions  
- Asynchronous job processing with worker pools  
- RESTful APIs for wallet operations  
- CLI-based auditing with streaming CSV export  
- Structured logging and request tracing  

This project emphasizes clean architecture, testability, and scalability.

---
## System Architecture

<img width="8192" height="2154" alt="image" src="https://github.com/user-attachments/assets/e348e34b-ed1a-436c-8b51-8492db789162" />


---

## Tech Stack

- **Language:** Go (Golang) v1.25+
- **Database:** PostgreSQL
- **HTTP Layer:** `net/http`
- **Concurrency Model:** Buffered channel + worker pool (10 workers)
- **Logging:** `slog`
- **CLI:** `flag`, `encoding/csv`, `bufio`

---

## Core Features

### 1. Account Management (REST API)

- Create account  
- Retrieve account  
- Update account  
- Delete account  

Balances are stored as **BIGINT** to avoid floating-point precision issues.

---

### 2. Asynchronous Transfers

Transfers are processed via:

- Buffered job queue  
- 10 background workers  
- Immediate `"pending"` response  
- Final status persisted in database  

Each request includes a unique `X-Request-ID` for full traceability.

---

### 3. Transaction Safety

The transfer engine guarantees:

- Atomic transactions using `db.BeginTx`
- Automatic rollback on failure
- Row-level locking using `SELECT ... FOR UPDATE`
- Validation before locking
- Context-aware DB operations

This prevents:
- Double spending  
- Lost updates  
- Race conditions  

---

### 4. Backpressure Protection

When the transfer queue is full:

- API returns **HTTP 429 — Too Many Requests**
- Message: `"transfer queue is full, please retry later"`

This protects the system from overload and memory exhaustion.

---

### 5. Validation Rules

The system enforces strict business rules:

- No negative transfers  
- No self-transfers  
- Insufficient funds detection  
- Destination account must exist  

Failed transactions are logged and stored for audit history.

---

### 6. Audit CLI

GopherPay includes an administrative CLI tool.

Example:

```bash
go run cmd/admin/main.go report --user=ACC1001
```

---

### Project Structure

gopherpay/
│
├── cmd/
│   ├── admin/        # CLI entrypoint
│   └── server/       # HTTP server entrypoint
│
├── internal/
│   ├── api/          # Handlers & middleware
│   ├── billing/      # Reporting logic
│   ├── config/       # Configuration
│   ├── db/           # Database setup
│   ├── logger/       # Logging setup
│   ├── middleware/   # HTTP middleware
│   ├── wallet/       # Core business logic
│   └── worker/       # Worker pool
│
├── migrations/       # SQL schema
├── reports/          # Generated CSV reports
└── README.md

---

### Running the Project

1. Clone Repository

```bash
git clone https://github.com/joundalatharv-max/gopher-pay.git
cd gopher-pay
```

2. Setup PostgreSQL

Create a database:

```bash
CREATE DATABASE gopherpay;
```

Run migrations from /migrations.

3. Configure Environment

Create a .env file:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=gopherpay
4. Run Server
go run cmd/server/main.go
```

Server runs on:

```bash
http://localhost:8080
```

5. Run CLI

```bash
go run cmd/admin/main.go report --user=ACC1001
```

---

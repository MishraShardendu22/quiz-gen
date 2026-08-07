# Quiz Generator AI Platform (Go + SQLite + Next.js)

An enterprise-grade, asynchronous quiz question generation engine built in **Go (Fiber)** with **SQLite**, background workers, LLM-as-Judge duplicate prevention, token usage tracking, and a **Next.js** frontend.

---

## 🏗️ Architecture Overview

```
                          ┌──────────────────────────┐
                          │    Next.js Frontend      │
                          └────────────┬─────────────┘
                                       │ HTTP REST
                                       ▼
                          ┌──────────────────────────┐
                          │      Fiber App API       │
                          └────────────┬─────────────┘
                                       │ Enqueue (Channel)
                                       ▼
                          ┌──────────────────────────┐
                          │    Background Worker     │
                          └─────┬──────────────┬─────┘
                                │              │
           OpenRouter (Resty)   │              │ TopicLockManager
           LLM & Judge Calls    ▼              ▼
                    ┌──────────────┐     ┌──────────────┐
                    │ OpenRouter   │     │ SQLite DB    │
                    │ LLM API      │     │ (WAL Mode)   │
                    └──────────────┘     └──────────────┘
```

- **Backend**: Go 1.22+ with Fiber v2, SQLite (`database/sql`), Resty v2, UUIDv7.
- **Ingestion**: Startup filesystem traversal (`content-pack/`), SHA-256 hash sync, markdown section chunking.
- **Generation & Deduplication**: Asynchronous background worker queue, prompt context injection, LLM-as-Judge duplicate comparison, 5-attempt retry limit.
- **Safety**: Per-topic mutex locking (`TopicLockManager`), WAL mode SQLite PRAGMAs (`journal_mode=WAL; busy_timeout=5000`), atomic database transactions.
- **Frontend**: Next.js 14 (TypeScript, TailwindCSS), high-contrast dark theme, nested retry session visualization.

---

## 🚀 Setup & Execution

### Quick Start (Single Command)
```bash
# Installs dependencies for backend and frontend, then launches both concurrently
make run
```

### Prerequisites
- **Go**: 1.22 or higher
- **Node.js**: 18+ & `npm`
- **OpenRouter API Key**: Set in `.env` or exported in shell.

### 1. Backend Setup
```bash
# Clone and enter project directory
cd quiz-gen

# Create environment configuration
echo 'OPENROUTER_API_KEY="sk-or-v1-your-key-here"' > .env

# Build and run backend (starts on :9000)
go build -o quiz-gen main.go
./quiz-gen
```

### 2. Frontend Setup
```bash
# Enter frontend directory
cd frontend

# Install dependencies
npm install

# Run dev server (starts on :3000)
npm run dev
```

Open `http://localhost:3000` in your browser.

---

## 📡 API Documentation

### `GET /topics`
Retrieves all discovered content topics with document counts.

### `POST /generate`
Enqueues a new quiz generation session asynchronously.
- **Header**: `Idempotency-Key: <unique-key>` (Optional)
- **Request Body**:
```json
{
  "topic_id": "019fcdcd-4cdd-7ef4-8c88-29bfb6ddba5a",
  "requested_count": 5,
  "token_budget": 5000,
  "idempotency_key": "optional-key"
}
```
- **Response**: `202 Accepted`
```json
{
  "code": 202,
  "success": true,
  "message": "Generation started",
  "data": {
    "session_id": "019fcdf1-22c6-75e0-85ae-709ae1a61c31",
    "status": "pending"
  }
}
```

### `GET /sessions`
Retrieves all quiz generation sessions ordered by creation date.

### `GET /sessions/:id`
Retrieves full details for a session including status, counts, token usage, error details, and generated questions.

### `POST /sessions/:id/retry`
Creates a new retry session for a failed session with an updated budget.
- **Request Body**:
```json
{
  "token_budget": 10000
}
```
- **Response**: `202 Accepted`

### `GET /usage`
Fetches overall token usage, prompt/completion token breakdown, estimated cost, and per-session usage.

---

## 🧪 Testing

```bash
# Run unit & package tests
go test ./...

# Run race condition detector
go test -race ./...

# Run static analysis
go vet ./...
```

---

## 📂 Project Structure

```
├── main.go                     # Entry point: DB init, migrations, ingestion, worker, Fiber HTTP
├── controller/                 # HTTP Handlers (Generate, Sessions, Topics, Usage)
├── model/                      # Data models & API DTOs
├── router/                     # Fiber routes & CORS middleware
├── service/
│   ├── db/                     # SQLite initialization, PRAGMAs, migrations & schema
│   ├── judge/                  # LLM-as-Judge duplicate comparison
│   ├── llmretry/               # Resilient LLM execution & JSON cleaning
│   ├── loader/                 # Recursive filesystem traversal & chunking
│   ├── openrouter/             # Singleton Resty client & cost estimation
│   ├── prompt/                 # Prompt template builders
│   ├── storage/                # Hash-based DB synchronization
│   └── worker/                 # Worker queue, TopicLockManager, storage CRUD, usage tracking
├── util/                       # Structured logger & standardized HTTP response builders
├── frontend/                   # Next.js 14 TypeScript Frontend
├── DECISIONS.md                # Technical decisions & tradeoffs
└── FINAL_REVIEW.md             # Engineering review & interview defense guide
```

---

## ⚠️ Known Limitations

1. **Post-Request Token Usage Discovery**: Token consumption is only returned by OpenRouter after a request completes. If an LLM call exceeds the session budget, the usage for that call is recorded, the session is marked `failed` with `"Token budget exhausted after generating X of Y questions."`, and all successfully generated questions are retained without loss.
2. **Single Worker Instance**: In-memory channel queue and `TopicLockManager` operate in a single Go process. Scaling across multiple binary instances would require Redis pub/sub and distributed locks.

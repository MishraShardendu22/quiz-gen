# AI Question-Generation Service

An asynchronous, resilient AI question-generation backend service and frontend dashboard built in **Go (Fiber)** with **SQLite**, background worker queues, per-topic concurrency locking, LLM-as-Judge duplicate prevention, and token budget tracking.

---

## Overview

This service powers automated quiz question generation from corporate training content (`content-pack/`) across multiple topics (`safety`, `service`, `privacy`). It handles slow LLM calls asynchronously via background job processing, guarantees zero semantic duplicate questions across prior sessions, enforces per-session token budgets, and handles malformed LLM outputs and network retries gracefully.

---

## Architecture

```
                          ┌──────────────────────────┐
                          │    Next.js Frontend      │
                          │   (http://localhost:3000)│
                          └────────────┬─────────────┘
                                       │ HTTP REST
                                       ▼
                          ┌──────────────────────────┐
                          │      Fiber App API       │
                          │   (http://localhost:9000)│
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

* **Backend**: Go 1.22+ using Fiber v2, SQLite (`database/sql`), Resty v2, and UUIDv7.
* **Content Ingestion**: Recursive filesystem traversal (`content-pack/`), incremental SHA-256 hash sync, document heading chunking.
* **Asynchronous Queue**: Non-blocking `POST /generate` returns immediately (`202 Accepted`) with a session ID. A configurable worker pool (default `runtime.NumCPU()`) processes queued sessions concurrently across different topics.
* **Concurrency Safety**: Per-topic mutex manager (`TopicLockManager`) ensures sessions for different topics execute concurrently while sessions on the same topic are serialized to prevent duplicate questions and race conditions.
* **Duplicate Prevention**: LLM-as-Judge verification compares new candidate questions against all previously generated questions for that topic.
* **Frontend**: Next.js 14 (TypeScript, TailwindCSS) dashboard for creating sessions, viewing question bank status, and tracking token/cost usage.

---

## Getting Started

### Prerequisites
* **Go**: 1.22 or higher
* **Node.js**: 18+ and `npm`
* **OpenRouter API Key**

### 1. Environment Configuration

Create a `.env` file in the root directory:

```bash
OPENROUTER_API_KEY="sk-or-v1-your-openrouter-key-here"
```

*(Alternatively, export `OPENROUTER_API_KEY` in your shell environment).*

### 2. Run the Service

#### Option A: Using `make` (Recommended Single Command)

Run the backend and frontend concurrently with automatic dependency installation:

```bash
make run
```

This single command will:
1. Download Go backend dependencies.
2. Install Node.js frontend dependencies via `npm`.
3. Start the Fiber backend server on `http://localhost:9000`.
4. Start the Next.js frontend dev server on `http://localhost:3000`.

#### Option B: Without `make` (Manual / Windows / Systems without `make`)

If your system does not have `make` installed, run the backend and frontend in two separate terminal windows:
Was
**Terminal 1 (Backend Server):**
```bash
# Download Go dependencies
go mod tidy

# Start Fiber backend on http://localhost:9000
go run main.go
```

**Terminal 2 (Frontend Dashboard):**
```bash
# Navigate to frontend directory
cd frontend

# Install Node.js dependencies
npm install

# Start Next.js dev server on http://localhost:3000
npm run dev
```

Open **[http://localhost:3000](http://localhost:3000)** in your browser once both servers are running.


---

## API Reference

### 1. `POST /generate`
Enqueues a new quiz question generation session.

* **Headers**: `Idempotency-Key: <string>` (Optional)
* **Request Body**:
```json
{
  "topic_id": "019fcdcd-4cdd-7ef4-8c88-29bfb6ddba5a",
  "requested_count": 10,
  "token_budget": 50000,
  "idempotency_key": "unique-request-key-123"
}
```
* **Response (`202 Accepted`)**:
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

---

### 2. `GET /sessions/:id`
Retrieves the status, question results, and token usage for a specific generation session.

* **Response (`200 OK`)**:
```json
{
  "code": 200,
  "success": true,
  "data": {
    "session": {
      "id": "019fcdf1-22c6-75e0-85ae-709ae1a61c31",
      "topic_id": "019fcdcd-4cdd-7ef4-8c88-29bfb6ddba5a",
      "topic_name": "safety",
      "status": "completed",
      "requested_count": 10,
      "generated_count": 10,
      "token_budget": 50000,
      "tokens_used": 3450,
      "created_at": 1770498435,
      "updated_at": 1770498442
    },
    "questions": [
      {
        "id": "019fcdf1-55a1-7ab3-91cd-40bfb6ddba5a",
        "question": "How often should pressure gauges on fire extinguishers be inspected?",
        "option_1": "Daily",
        "option_2": "Weekly",
        "option_3": "Monthly",
        "option_4": "Annually",
        "correct_answer": 2,
        "explanation": "Safety standards mandate monthly visual inspections of pressure gauges."
      }
    ]
  }
}
```

---

### 3. `GET /sessions`
Lists all generation sessions ordered by creation time (most recent first).

* **Response (`200 OK`)**: Returns a list of session summaries.

---

### 4. `GET /topics`
Lists all available content topics loaded from `content-pack/` alongside document counts.

* **Response (`200 OK`)**:
```json
{
  "code": 200,
  "success": true,
  "data": [
    {
      "id": "019fcdcd-4cdd-7ef4-8c88-29bfb6ddba5a",
      "name": "safety",
      "document_count": 3,
      "total_questions": 12
    }
  ]
}
```

---

### 5. `GET /usage`
Reports aggregated token consumption, prompt vs completion token breakdown, estimated cost ($), and per-session breakdown.

* **Response (`200 OK`)**:
```json
{
  "code": 200,
  "success": true,
  "data": {
    "total_prompt_tokens": 12500,
    "total_completion_tokens": 3200,
    "total_tokens": 15700,
    "total_cost": 0.0,
    "session_count": 4,
    "sessions": []
  }
}
```

---

### 6. `POST /sessions/:id/retry` *(Extension)*
Creates a retry session for a failed or budget-exhausted session with an updated token budget.

---

## Edge Cases & Resilience Behavior

1. **Idempotency**: If `POST /generate` is called with an existing `idempotency_key` or duplicate body key, the service immediately returns the existing session without initiating a new generation job or spending additional tokens.
2. **Per-Topic Concurrency**: When multiple sessions are requested simultaneously, different topics process concurrently via the worker pool. For sessions requesting the same topic, the worker acquires a per-topic mutex (`TopicLockManager`), serializing same-topic processing to ensure candidate questions are evaluated against up-to-date question banks without race conditions.
3. **Malformed LLM Output**: All LLM JSON responses are passed through schema validation (`service/validator`). Invalid JSON structures, incorrect option counts (!= 4), invalid answer indexes (< 0 or > 3), or empty fields are rejected and trigger structured regeneration.
4. **Partial Failure & Token Budget Exhaustion**: Token usage is checked prior to each request and updated immediately after completion. If a budget is exceeded mid-generation, all valid questions generated up to that point are saved to SQLite, and the session status is set to `failed` with a clear explanation.

---

## Testing

The repository includes unit tests built using Go's standard `testing` package.

```bash
# Run unit tests across all packages
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...
```

See [`DECISIONS.md`](file:///home/mishrashardendu22/Coding_Stuff_Fedora/quiz-gen/DECISIONS.md) for rationale on test coverage and engineering scope.

---

## Directory Structure

```
├── main.go                     # Application entry point & service wiring
├── controller/                 # HTTP API handlers (generate, sessions, topics, usage)
├── model/                      # Go structs, DB schemas, and JSON DTOs
├── router/                     # Fiber routes & CORS middleware
├── service/
│   ├── db/                     # SQLite initialization, PRAGMAs (WAL mode), migrations
│   ├── judge/                  # LLM-as-Judge semantic duplicate detection
│   ├── llmretry/               # Resilient LLM client execution & JSON extraction
│   ├── loader/                 # Content loader & markdown chunking
│   ├── openrouter/             # OpenRouter API client & usage calculation
│   ├── prompt/                 # LLM prompt builder
│   ├── storage/                # Hash-based content sync & SQLite transactions
│   ├── validator/              # Question schema validation
│   └── worker/                 # Asynchronous background worker queue & TopicLockManager
├── util/                       # Structured logger & standardized HTTP response helpers
├── content-pack/               # Markdown document content pack (safety, service, privacy)
├── frontend/                   # Next.js 14 dashboard UI
├── DECISIONS.md                # Engineering decisions, trade-offs, and architecture notes
└── README.md                   # System documentation & setup guide
```

---

## Engineering Decisions

For full details on architectural choices, hash-based incremental ingestion, SQLite WAL mode, duplicate prevention tradeoffs, and token accounting, see **[`DECISIONS.md`](file:///home/mishrashardendu22/Coding_Stuff_Fedora/quiz-gen/DECISIONS.md)**.

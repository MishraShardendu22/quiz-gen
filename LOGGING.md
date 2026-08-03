# Logging Implementation

## Design Decisions

### Logger Utility (✅ Introduced)
A simple logging utility package (`pkg/log`) was introduced using Go's standard library `log/slog`.

**Why:**
- Provides structured, JSON-formatted logging
- Centralized configuration and initialization
- Easy to enhance later for different environments (dev, staging, prod)
- Avoids repeating logger initialization across packages
- Provides a clean, simple API

**API:**
- `log.Info(msg, attrs...)` - Info level logging
- `log.Warn(msg, attrs...)` - Warning level logging
- `log.Error(msg, attrs...)` - Error level logging
- `log.Debug(msg, attrs...)` - Debug level logging
- `log.Get()` - Access underlying `*slog.Logger` for advanced usage

### Error Utility (❌ Not Introduced)
No separate error utility was introduced. Current error handling is idiomatic Go and sufficient.

**Why not:**
- `fmt.Errorf()` with context wrapping is the Go standard
- Custom error types aren't needed for this scale project
- Adding an error package would be premature abstraction
- Consistency is easy to maintain without a separate package
- If typed errors become necessary later, they can be introduced then

**Current approach is preserved:**
- Error wrapping with context: `fmt.Errorf("operation: %w", err)`
- Errors bubble up to caller for handling
- HTTP layer converts errors to responses

---

## What Changed

### 1. New Package: `pkg/log/logger.go`
- Single logger instance initialized at startup
- Uses `log/slog` with JSON output for structured logging
- Provides convenience functions: `Info()`, `Error()`, `Warn()`, `Debug()`
- Can be enhanced later with environment-specific handlers

### 2. `main.go`
- Replaced `fmt.Println/Printf` and `log.Fatalf()` with structured logging
- Logs application startup, database initialization, and server startup
- Uses `panic()` on fatal errors (cleaner than `log.Fatalf()`)

### 3. `service/db/db.go`
- Logs database initialization and connection
- Logs migration start and completion
- Provides visibility into database setup phase

### 4. `service/loader/content-loader.go`
- Logs content discovery start and completion
- Logs each topic discovery with document count
- Logs total topics and documents at the end
- Added `countTotalDocuments()` helper function

### 5. `service/storage/storage.go`
- **Most detailed logging:** Documents every operation
- Logs topic creation vs. reuse
- Logs document states: created, updated, skipped, deleted
- Logs chunk count for each document
- Logs orphaned document deletion
- **Sync summary:** Final metrics including:
  - Duration (milliseconds)
  - Topics created/existing
  - Documents: created, updated, unchanged, deleted
  - Total chunks created

### 6. `controller/topics.go`
- Logs HTTP errors when database queries fail
- Logs successful retrieval count
- Improves observability of API operations

---

## Log Output Example

```json
{"time":"2026-08-04T03:30:00.000Z","level":"INFO","msg":"starting application"}
{"time":"2026-08-04T03:30:00.001Z","level":"INFO","msg":"initializing database","path":"./quiz.db"}
{"time":"2026-08-04T03:30:00.050Z","level":"INFO","msg":"database initialized successfully"}
{"time":"2026-08-04T03:30:00.051Z","level":"INFO","msg":"running database migrations"}
{"time":"2026-08-04T03:30:00.100Z","level":"INFO","msg":"database migrations completed"}
{"time":"2026-08-04T03:30:00.101Z","level":"INFO","msg":"starting content discovery","directory":"./content-pack"}
{"time":"2026-08-04T03:30:00.102Z","level":"INFO","msg":"discovering topic","name":"golang"}
{"time":"2026-08-04T03:30:00.150Z","level":"INFO","msg":"topic discovery complete","name":"golang","documents":15}
{"time":"2026-08-04T03:30:00.151Z","level":"INFO","msg":"content discovery complete","topics":3,"total_documents":45}
{"time":"2026-08-04T03:30:00.152Z","level":"INFO","msg":"starting content synchronization","topics":3}
{"time":"2026-08-04T03:30:00.153Z","level":"INFO","msg":"syncing topic","name":"golang","documents":15}
{"time":"2026-08-04T03:30:00.154Z","level":"INFO","msg":"created new topic","name":"golang","id":"uuid-here"}
{"time":"2026-08-04T03:30:00.155Z","level":"INFO","msg":"created new document","topic":"golang","path":"basics.md","chunks":3}
{"time":"2026-08-04T03:30:00.156Z","level":"INFO","msg":"skipping unchanged document","topic":"golang","path":"advanced.md"}
{"time":"2026-08-04T03:30:00.200Z","level":"INFO","msg":"synchronization complete","duration_ms":48,"topics_created":2,"topics_existing":1,"documents_created":40,"documents_updated":3,"documents_unchanged":2,"documents_deleted":0,"chunks_created":95}
{"time":"2026-08-04T03:30:00.201Z","level":"INFO","msg":"starting http server","address":":9000"}
```

---

## Why This Design Is Better

### Before
- Mixed logging mechanisms (log package, fmt, printf)
- No visibility into content sync progress
- No structured metrics
- HTTP errors silently logged or not logged at all
- Hard to parse logs or send them to centralized systems

### After
- **Consistent:** All logging through same interface
- **Structured:** JSON format for easy parsing, filtering, and aggregation
- **Observable:** Every important state transition is logged
- **Minimal:** Logs only meaningful events, not every SQL query or chunk
- **Debuggable:** When something fails, logs show exactly what was being processed
- **Metric-Ready:** Final sync summary in structured format, easy to extract metrics

---

## Where We Did NOT Add Logging (Intentionally)

### What NOT to log:
1. **Every SQL query** - Would create excessive noise. We log state transitions instead.
2. **Every chunk insertion** - Chunk count is logged per document (summary level).
3. **Individual file reads during discovery** - We log topic completion, not each file.
4. **Hash computations** - Not user-visible; doesn't help debugging.
5. **Directory traversals** - Logged at topic level, which is sufficient.

**Philosophy:** Log to understand what the application is doing, not what it's computing.

---

## Future Enhancements

The logging infrastructure makes it easy to add:
- **Environment-specific handlers** (text format for dev, JSON for prod)
- **Log levels control** via environment variables
- **Error sampling** if logs get too large
- **Structured error codes** if needed (e.g., `"error_code":"DOC_001"`)
- **Metrics export** (e.g., export sync stats to Prometheus)

But for now, this implementation is simple, effective, and idiomatic Go.

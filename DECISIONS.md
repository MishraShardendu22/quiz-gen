# Architecture & Engineering Decisions

This document outlines key technical decisions, trade-offs, and intentionally scoped boundaries in the Quiz Generator system.

---

## 1. UUIDv7 over UUIDv4
- **Decision**: Used UUIDv7 for all primary keys (`topics`, `documents`, `chunks`, `sessions`, `questions`, `usage`).
- **Rationale**: UUIDv7 combines a Unix timestamp prefix with random bytes. This time-ordering avoids SQLite B-Tree index fragmentation caused by random UUIDv4 inserts, improving write performance significantly.

## 2. Ingestion Sync Strategy (Hash-Based Incremental Sync)
- **Decision**: Computed SHA-256 hashes of markdown files during filesystem traversal.
- **Rationale**: Avoids re-chunking unchanged files on startup. If a file hash matches SQLite, ingestion is skipped. If modified, old chunks are purged and updated chunks ingested.

did not implement some thing like where only the chunk edited is changed in database, for simplicity all chunks are deleted and re-ingested if the file is modified. This avoids complex diffing logic and ensures consistency between filesystem and database.

why implement chunking at all ? at current implementation its of no use, it would have been usefull if we had implemented some sort of vector embeddings and semantic search or maybe if we had implemented some sort of LLM based question generation from chunks, but currently for simplycity we are just storing the chunks in database and not using them for anything else. there is a seperate hy 500 line code we can use that as well or just not chunk at all. 

cost of llms i am using free models so basically hardcoded free models else I would have to fetch cost of llm using some api there was really no need

## 3. Asynchronous Worker Queue & Per-Topic Locking (`TopicLockManager`)
- **Decision**: In-memory channel queue (`chan uuid.UUID`) processed by a background worker goroutine with a per-topic mutex manager.
- **Rationale**: Ensures concurrent sessions for different topics execute in parallel without lock contention. For the same topic, mutex locks protect existing question lookups and database writes without holding locks during network requests to OpenRouter.

## 4. LLM-as-Judge Duplicate Prevention & 5-Attempt Limit
- **Decision**: Implemented an LLM-as-Judge comparison pass (`service/judge/judge.go`) that checks candidate questions against all previously generated questions for the topic.
- **Rationale**: Rather than using vector embeddings or cosine similarity (which require extra dependencies or local models), LLM judging detects semantic duplicates directly in pro[mpt context. A maximum of 5 regeneration attempts is enforced to prevent infinite loops on topic exhaustion.

## 5. Token Budget Limitation & Partial Failures
- **Decision**: Token budget (`remaining_budget = token_budget - tokens_used`) is evaluated before every OpenRouter completion request.
- **Tradeoff**: Token usage cannot be known before an LLM API call completes. If a completion exceeds the remaining budget, the token usage for that call is recorded, the session is marked `failed` with `"Token budget exhausted after generating X of Y questions."`, and all stored questions remain intact without data loss.

## 6. Immutable Failed Sessions & Retries
- **Decision**: Retrying a failed session (`POST /sessions/:id/retry`) creates a completely new session with an updated budget, keeping the old failed session immutable.
- **Rationale**: Preserves audit trails and cost history. Old failed sessions remain inspectable in the frontend and API.

## 7. SQLite Concurrency Configuration (WAL Mode & Busy Timeout)
- **Decision**: Executed `PRAGMA journal_mode=WAL;`, `PRAGMA busy_timeout=5000;`, `PRAGMA foreign_keys=ON;`, and `PRAGMA synchronous=NORMAL;` on database initialization.
- **Rationale**: Prevents `database is locked` panics under concurrent write operations from worker goroutines and Fiber HTTP handlers.

---

## 8. Intentionally Skipped / Future Improvements
- **Distributed Lock / Redis Queue**: Not implemented; in-memory channels and mutexes meet all requirements within single-process scope.
- **Vector Embeddings / Cosine Similarity**: Not implemented; LLM-as-Judge handles semantic duplicate comparison directly as required.
- **Distributed Worker Scaling**: Horizontal multi-instance scaling would require PostgreSQL and Redis.

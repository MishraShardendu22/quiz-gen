# Architecture & Engineering Decisions

This document outlines the key technical decisions, trade-offs, and intentionally scoped boundaries in the Quiz Generator system.

---

## 1. UUIDv7 over UUIDv4

UUIDv7 is used for all primary keys (`topics`, `documents`, `chunks`, `sessions`, `questions`, `usage`).

UUIDv7 combines a Unix timestamp prefix with random bytes. This time ordering avoids SQLite B-Tree index fragmentation caused by random UUIDv4 inserts, significantly improving write performance.

---

## 2. Ingestion Sync Strategy (Hash-Based Incremental Sync)

SHA-256 hashes of Markdown files are computed during filesystem traversal.

This avoids re-chunking unchanged files on startup. If a file hash matches the one stored in SQLite, ingestion is skipped. If the file has been modified, all old chunks are deleted and the updated chunks are re-ingested.

**Why not update only modified chunks?**

For simplicity, the implementation deletes all existing chunks for a modified file and re-ingests them instead of detecting which individual chunks changed. This avoids implementing complex diffing logic while ensuring the filesystem and database remain consistent.

---

## 3. Document Chunking

**Why implement chunking at all?**

At the current stage, chunking is not actually used by the application. It would become useful if vector embeddings and semantic search were implemented, or if question generation used retrieved chunks instead of the entire document.

Currently, chunks are simply stored in the database and are not used elsewhere. There is also a separate Hy implementation (around 500 lines) that could be used for retrieval later. Alternatively, chunking could have been omitted entirely in the current implementation.

---

## 4. LLM Cost Tracking

The project only uses OpenRouter models with zero-cost pricing, so token generation cost is treated as $0 and does not require runtime cost calculation.

Supporting paid models would require retrieving each model's input and output token pricing, or maintaining an up-to-date pricing table, and calculating costs based on token usage. Since the project exclusively targets free models, implementing dynamic cost tracking was unnecessary.

---

## 5. Asynchronous Worker Queue & Per-Topic Locking (`TopicLockManager`)

The system uses an in-memory channel queue (`chan uuid.UUID`) processed by a configurable worker pool (default `runtime.NumCPU()`) with a per-topic mutex manager (`TopicLockManager`).

This ensures that concurrent sessions for different topics execute in parallel across worker goroutines without lock contention. For sessions targeting the same topic, `TopicLockManager` serializes processing to protect existing question lookups and database writes without holding locks during network requests to OpenRouter.

---

## 6. LLM-as-Judge Duplicate Prevention & 5-Attempt Limit

An LLM-as-Judge comparison pass (`service/judge/judge.go`) checks candidate questions against all previously generated questions for the topic.

Rather than using vector embeddings or cosine similarity, which require additional dependencies or local models, the LLM performs semantic duplicate detection directly. A maximum of five regeneration attempts is enforced to prevent infinite loops when the topic is exhausted.

**Why not use embeddings?**

Suppose the LLM generates a batch containing duplicate questions. We can ask it to regenerate those. However, the next batch may still contain questions that duplicate previously accepted questions. This would require continuously tracking all generated questions and comparing every new batch against them.

Instead, the implementation simply asks the LLM to compare the newly generated questions against all existing questions for the topic and return only unique ones.

A hybrid approach combining embeddings with LLM verification could be implemented later, but this is sufficient for the current project.

---

## 7. Token Budget Limitation & Partial Failures

### Token Budget Enforcement

OpenRouter reports actual token usage only after a request completes. Because of this, the service cannot guarantee that a request will remain within the configured budget without estimating token usage beforehand.

Rather than relying on estimates, the service checks the budget before every LLM request using the actual recorded usage. If sufficient budget remains, one additional request is allowed. After the request completes, the actual usage reported by OpenRouter is recorded. If the configured budget has been exceeded, the session is marked as failed, no further LLM requests are made, and all successfully generated questions are preserved.

This approach may exceed the configured budget by the cost of at most one LLM request, but it keeps accounting accurate and avoids speculative token estimation.

---

## 8. SQLite Concurrency Configuration (WAL Mode & Busy Timeout)

The following PRAGMAs are executed during database initialization:

* `PRAGMA journal_mode=WAL;`
* `PRAGMA busy_timeout=5000;`
* `PRAGMA foreign_keys=ON;`
* `PRAGMA synchronous=NORMAL;`

These settings help prevent `database is locked` errors under concurrent write operations from worker goroutines and Fiber HTTP handlers.

They primarily improve concurrency and performance. WAL mode allows concurrent reads and writes, while the busy timeout lets SQLite wait for a lock instead of immediately returning an error.

---

## 9. Intentionally Skipped / Future Improvements

### Vector Embeddings / Cosine Similarity

Not implemented, as the LLM-as-Judge approach satisfies the duplicate detection requirement for the current project.

A hybrid approach combining embeddings for fast candidate retrieval with LLM verification could be implemented later if the dataset grows significantly.

---

## 10. What if Someone Requests 10,000 Questions?

The application currently rejects requests for more than 20 questions.

If larger requests, for example 500 questions, were to be supported, the system would need to implement batch generation, producing questions incrementally (for example, 20 questions per batch) until the requested number is reached.

This has not been implemented yet because it is unnecessary for the current requirements, but it would be a reasonable improvement if larger-scale generation becomes necessary.

## 11. Why only test for small cleaning utilities?

I only created unit tests for small cleaning utilities like `cleanMarkdown` and `cleanQuestion`. The other functions I simply tested manually with a frontend interface. 

## 12. Cache repeated or identical generation requests to avoid redundant LLM calls.
I didnt implement this because -
    - Even for Identical prompt, the response of llm is not deterministic, plus it has context of previous questions, so it will one way or the other create question that are not duplicate. 
    - Caching is helpful when the response is some what deterministic, but in this case it is not. So caching will not help much. 
    - Although the assignment mentions caching repeated requests, it also requires that questions generated in a new session must not duplicate questions from any previous session. Reusing a cached question set would violate that requirement by returning identical questions across sessions.
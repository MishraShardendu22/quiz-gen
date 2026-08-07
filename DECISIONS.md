# Architecture & Engineering Decisions

This document outlines key technical decisions, trade-offs, and intentionally scoped boundaries in the Quiz Generator system.

---

## 1. UUIDv7 over UUIDv4

* **Decision:** Used UUIDv7 for all primary keys (`topics`, `documents`, `chunks`, `sessions`, `questions`, `usage`).
* **Rationale:** UUIDv7 combines a Unix timestamp prefix with random bytes. This time ordering avoids SQLite B-Tree index fragmentation caused by random UUIDv4 inserts, improving write performance significantly.

---

## 2. Ingestion Sync Strategy (Hash-Based Incremental Sync)

* **Decision:** Computed SHA-256 hashes of Markdown files during filesystem traversal.
* **Rationale:** Avoids re-chunking unchanged files on startup. If a file hash matches the one stored in SQLite, ingestion is skipped. If the file is modified, all old chunks are deleted and the updated chunks are re-ingested.

**Why not update only modified chunks?**

For simplicity, the implementation deletes all existing chunks for a modified file and re-ingests them instead of detecting which individual chunks changed. This avoids implementing complex diffing logic while ensuring the filesystem and database always remain consistent.

---

## 3. Document Chunking

**Why implement chunking at all?**

At the current stage, chunking is not actually used by the application. It would become useful if vector embeddings and semantic search were implemented, or if question generation used retrieved chunks instead of the entire document.

Currently, chunks are simply stored in the database and are not used anywhere else. There is also a separate Hy implementation (around 500 lines) that could be used for retrieval later. Alternatively, chunking could have been omitted entirely in the current implementation.

---

## 4. LLM Cost Tracking

I am only using free OpenRouter models, so model pricing is effectively hardcoded as zero.

Supporting paid models would require fetching model pricing through an API and calculating generation costs dynamically. Since the project only targets free models, there was no practical need to implement this.

---

## 5. Asynchronous Worker Queue & Per-Topic Locking (`TopicLockManager`)

* **Decision:** Uses an in-memory channel queue (`chan uuid.UUID`) processed by a background worker goroutine with a per-topic mutex manager.
* **Rationale:** Ensures concurrent sessions for different topics execute in parallel without lock contention. For the same topic, mutex locks protect existing question lookups and database writes without holding locks during network requests to OpenRouter.

---

## 6. LLM-as-Judge Duplicate Prevention & 5-Attempt Limit

* **Decision:** Implemented an LLM-as-Judge comparison pass (`service/judge/judge.go`) that checks candidate questions against all previously generated questions for the topic.
* **Rationale:** Rather than using vector embeddings or cosine similarity (which require additional dependencies or local models), the LLM directly performs semantic duplicate detection. A maximum of five regeneration attempts is enforced to prevent infinite loops when the topic is exhausted.

**Why not use embeddings?**

Suppose the LLM generates a batch containing duplicate questions. We can ask it to regenerate those. However, the next batch may still contain questions that duplicate previously accepted questions. This would require continuously tracking all generated questions and comparing every new batch against them.

Instead, the implementation simply asks the LLM to compare the newly generated questions against all existing questions for the topic and return only unique ones. A hybrid approach combining embeddings with LLM verification could be implemented later, but this is sufficient for the current project.

---

## 7. Token Budget Limitation & Partial Failures

### Token Budget Enforcement

OpenRouter reports actual token usage only after a request completes. Because of this, the service cannot guarantee that a request will remain within the configured budget without estimating token usage beforehand. Rather than relying on estimates, the service checks the budget before every LLM request using the actual recorded usage. If sufficient budget remains, one additional request is allowed. After the request completes, the actual usage reported by OpenRouter is recorded. If the configured budget has now been exceeded, the session is marked as failed, no further LLM requests are made, and all successfully generated questions are preserved. This approach may exceed the configured budget by the cost of at most one LLM request, but keeps accounting accurate and avoids speculative token estimation.

---

## 8. SQLite Concurrency Configuration (WAL Mode & Busy Timeout)

* **Decision:** Executed `PRAGMA journal_mode=WAL;`, `PRAGMA busy_timeout=5000;`, `PRAGMA foreign_keys=ON;`, and `PRAGMA synchronous=NORMAL;` during database initialization.
* **Rationale:** Prevents `database is locked` errors under concurrent write operations from worker goroutines and Fiber HTTP handlers.

---

## 9. Intentionally Skipped / Future Improvements

* **Vector Embeddings / Cosine Similarity:** Not implemented, as LLM-as-Judge satisfies the duplicate detection requirement for the current project. A hybrid approach combining embeddings for fast candidate retrieval with LLM verification could be implemented later if the dataset grows significantly.

## 10. What if someones asks to make 10K questions ?
frontend wont accept a number more than 20, if i was to allow lets say 500 question as a numebr i would have to implement a batch like system, where it generates incrementally like 20 at a time. Hvaent implemented this yet, but it would be a good idea to do so in the future. (its not needed rn)
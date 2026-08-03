# Take-Home Assignment: AI Question-Generation Service

## Context

We are building a Learning Management System (LMS) for corporate training. One of its features allows trainers to generate quiz questions automatically from the training content they have uploaded. You are building the backend service that powers this feature.

This is a standalone service — it is not connected to our main system. You do not need any infrastructure beyond your machine and the provided API key.

## What you are building

A **Question Generation Service** that:
1. Loads a set of content documents (provided — see "Content Pack" below) organized by topic
2. Accepts a request to generate N quiz questions for a given topic
3. Calls an LLM API to generate the questions
4. Returns generated questions, ensuring no duplicates with previously generated questions
5. Tracks token/cost usage and enforces a budget

## Functional Requirements

### Content Ingestion
- Load the provided content documents from the filesystem at startup or on-demand
- Each document belongs to a topic (the directory structure makes this clear)
- You may index, chunk, or pre-process the content however you see fit

### Question Generation
- `POST /generate` — request N questions for a given topic
- Each question is a multiple-choice question (MCQ) with: question text, 4 options, correct answer index (0-3), and a brief explanation
- Generation involves one or more LLM API calls — these are slow (multi-second). The API must be **asynchronous**: return a session/job ID immediately, expose a status endpoint for polling, and deliver results when ready
- The caller may specify: topic, number of questions (default 10), and an optional idempotency key

### Duplicate Prevention
- Questions generated in a new session must not duplicate questions from any prior session — semantically, not just string-equal
- The approach is entirely your choice: embeddings, LLM-as-judge, heuristics, or any combination. We are grading the reasoning, not the sophistication
- Document your approach and its limitations honestly in DECISIONS.md

### Cost & Usage Tracking
- Track total tokens consumed across all LLM calls
- Enforce a per-session token budget (configurable, default 50,000 tokens)
- `GET /usage` — report total tokens consumed and estimated cost
- Cache repeated or identical generation requests to avoid redundant LLM calls

### Error Handling
The LLM API will, in the real world:
- Return malformed JSON (the model does not always follow instructions)
- Rate-limit you (HTTP 429)
- Time out
- Produce questions that do not match the requested schema (wrong number of options, missing explanation, etc.)

Your service must handle all of these gracefully: validate every LLM response against a schema, retry with backoff where appropriate, and never store or return an invalid question.

## API Surface (minimum)

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/generate` | Start a generation session. Body: `{ topic: string, count?: number, idempotency_key?: string }`. Returns: `{ session_id: string, status: "pending" }` |
| `GET` | `/sessions/:id` | Get session status and results. Returns: `{ session_id, status: "pending" \| "completed" \| "failed", questions?: Question[], error?: string }` |
| `GET` | `/sessions` | List all sessions (most recent first) |
| `GET` | `/usage` | Total tokens consumed, estimated cost, per-session breakdown |
| `GET` | `/topics` | List available topics and their document counts |

You may add endpoints if needed. Document any additions.

## Edge Cases (these are the interesting part)

These are explicitly called out because they are where we evaluate your engineering judgment:

1. **Idempotent generation:** If `POST /generate` is called twice with the same idempotency key, the second call must not start a new generation or spend additional tokens — it must return the existing session.
2. **Partial failure:** A session requests 20 questions but the LLM fails irrecoverably after generating 12. What is the session's final state? What does the client see? Define and document this behavior.
3. **Concurrent sessions on the same topic:** Two requests come in simultaneously for the same topic. They must not race the duplicate-prevention store and produce overlapping questions.
4. **Budget exhaustion mid-generation:** The session hits its token budget after generating 7 of 20 requested questions. What happens?
5. **Malformed LLM output:** The model returns valid JSON but with 3 options instead of 4, or a correct answer index of 5. What does your service do?

## What is provided to you

1. **An OpenRouter API key** — sent to you separately via a secure channel. This gives you access to LLM models via the OpenRouter API (OpenAI-compatible format). Documentation: https://openrouter.ai/docs
2. **A content pack** — 9 training documents across 3 topics:
   - `safety/` — fire safety, hazard reporting, PPE standards (3 documents)
   - `service/` — de-escalation, written communication, difficult customers (3 documents)
   - `privacy/` — data classification, phishing defense, access control (3 documents)
3. This specification document

The content pack directory will be shared alongside this document.

## Constraints

- **Time budget:** approximately 15 hours. We grade scope judgment — an over-built submission scores worse than a well-scoped one. Build what matters, document what you skipped and why.
- **Deadline:** 7 calendar days from receiving the API key.
- **Language:** TypeScript or Go — your choice.
- **Storage:** SQLite or flat files only. No external database server, no Redis, no Docker required. The service must start with a single command (e.g. `bun run dev`, `npm start`, `go run .`).
- **AI tools (Claude, GPT, etc.) are explicitly permitted.** You will be asked to defend every line of your code in a live review. If you do not understand a piece of code, do not include it.
- **Tests:** Write tests where you believe they earn their keep. Your choice of what to test and what not to test is itself graded. Do not write tests just to have tests — justify your coverage decisions in DECISIONS.md.

## Deliverables

1. **A Git repository** (share via GitHub/GitLab) with:
   - All source code
   - A `README.md` with: setup instructions, how to run, API reference (brief), how to run tests
   - A `DECISIONS.md` documenting: key design decisions and trade-offs, what you deliberately chose not to build and why, what you would do differently with more time, known limitations of your duplicate-prevention approach
   - Any seed/migration files needed to set up storage
2. **A working service** that runs with a single command and can be tested against the provided content pack

## What we grade

We evaluate the following, in roughly equal weight:

1. **Resilience correctness** — how well does the service handle malformed output, retries, idempotency, partial failures, and budget limits?
2. **Duplicate-prevention design** — is the reasoning sound? Are the limitations honest? We are not looking for the most sophisticated approach — we are looking for the most honest and well-reasoned one.
3. **API and code taste** — is the async job model clean? Are status codes correct? Is naming consistent? Is there unnecessary abstraction?
4. **Cost awareness** — is the budget enforced? Is caching used where it helps? Is usage reporting accurate?
5. **Documentation and scope judgment** — does DECISIONS.md show clear thinking? Did you build the right things and skip the right things? Is the code readable?
6. **Live defense** — can you explain every decision, modify code on the spot, and reason about edge cases under questioning?

## Questions?

If anything in this spec is ambiguous, make a reasonable assumption, document it in DECISIONS.md, and move on. We are interested in your judgment, not in getting clarifying questions during the assignment. If a clarification is truly blocking, email _ or send a whatsapp message to me at _.
shardendu-take-home.md
Displaying shardendu-take-home.md.
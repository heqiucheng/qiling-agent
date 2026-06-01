# Memory Retrieval And Vector Recall

Status: design ready for implementation  
Date: 2026-06-01

## Goal

Qiling Agent needs a memory retrieval layer that can find useful context beyond the current customer page and beyond compact durable facts.

The retrieval layer must answer:

```text
Given the current customer, task, conversation, and user action,
which past facts, messages, cases, scripts, and SOP snippets should the Agent see now?
```

This layer is not a replacement for short-term memory or long-term memory.

| Layer | Purpose | Current Status |
| --- | --- | --- |
| Short-term memory | Bounded current customer context from recent operational data | Implemented |
| Long-term memory | Correctable durable customer facts | Implemented |
| Vector recall | Fuzzy retrieval across larger history and knowledge sources | This design |

## Non-Goals

- Do not store full transcripts directly in the prompt.
- Do not let vector search override rejected long-term facts.
- Do not call external embedding or LLM APIs without explicit approval.
- Do not require vector infrastructure in default CI.
- Do not hide retrieval decisions. Every recalled item must have source metadata.
- Do not make handlers, repositories, or frontend pages call model providers directly.

## Recommended Architecture

```text
conversation_messages
customer_memory_facts
followup_tasks
agent_runs
audit_events
knowledge_articles
        |
        v
embedding_jobs
        |
        v
embedding provider
        |
        v
memory_vectors
        |
        v
retrieval service
        |
        v
reranker / filters
        |
        v
agent.RunInput.RetrievedMemoryContext
```

## Package Boundaries

| Package | Responsibility |
| --- | --- |
| `internal/service` | Orchestrates recall for a customer/task before calling `agent.Runner` |
| `internal/agent` | Accepts retrieved memory context and places it in the prompt under a separate section |
| `internal/store` | Persists vector metadata, embedding jobs, recall traces, and source links |
| `internal/integration/embedding` | Provider-neutral embedding client interface and mock implementation |
| `internal/integration/vectorstore` | Provider-neutral vector search adapter if an external vector DB is used |
| `internal/job` | Async embedding/indexing workers |

Handlers must not build prompts, call embeddings, or run vector searches directly.

## What Should Be Indexed

| Source | Index? | Reason | Chunk Rule |
| --- | --- | --- | --- |
| `conversation_messages` | Yes | Fuzzy recall of past customer phrasing, objections, and context | One message or small adjacent message group |
| `customer_memory_facts` active facts | Yes | Enables semantic lookup of durable facts | One fact per vector |
| `customer_memory_facts` rejected facts | No | Rejected facts must not re-enter context through vector recall | Not indexed or excluded by status |
| `followup_tasks` copied scripts | Yes | Copied scripts are positive examples | Script + outcome metadata |
| `followup_tasks` skipped / marked_wrong | Conditional | Useful as negative examples, but not prompt context by default | Index for evaluation, exclude from prompt unless requested |
| `agent_runs` | Yes | Prompt/output traces help debug and compare prompt versions | Summary + output fields, not full raw prompt |
| `audit_events` | Yes | Human feedback and governance signals improve ranking | Compact event summary |
| Product/SOP/case knowledge | Yes | Provides reusable sales knowledge and examples | 300-800 tokens per chunk |
| Raw uploads | No direct prompt recall | Raw uploads are source material, not clean memory | Parse first into messages or structured facts |

## Vector Record Model

Proposed table: `memory_vectors`

| Field | Purpose |
| --- | --- |
| `id` | Vector record id |
| `tenant_id` | Future enterprise isolation |
| `customer_id` | Optional customer anchor |
| `source_type` | `conversation_message`, `memory_fact`, `followup_task`, `agent_run`, `audit_event`, `knowledge_article` |
| `source_id` | Source row id |
| `chunk_key` | Stable key for idempotent re-indexing |
| `content` | Text used for embedding and explainability |
| `content_hash` | Detect whether re-embedding is needed |
| `embedding_model` | Embedding model/version |
| `embedding_dim` | Vector dimension |
| `vector_ref` | External vector DB id or local vector key |
| `status` | `active`, `stale`, `deleted`, `blocked` |
| `created_at` / `updated_at` | Operational timestamps |

Proposed table: `embedding_jobs`

| Field | Purpose |
| --- | --- |
| `id` | Job id |
| `source_type` / `source_id` | Source to embed |
| `status` | `pending`, `running`, `succeeded`, `failed`, `dead` |
| `attempts` | Retry count |
| `last_error` | Failure reason |
| `scheduled_at` / `locked_at` / `completed_at` | Worker control |

Proposed table: `recall_traces`

| Field | Purpose |
| --- | --- |
| `id` | Trace id |
| `agent_run_id` | Link recall to AgentRun |
| `query_text` | Compact retrieval query |
| `filters` | Permission/status/source filters |
| `candidate_count` | Number of raw vector hits |
| `selected_count` | Number injected into prompt |
| `selected_items` | Source ids, scores, and reasons |
| `created_at` | Trace timestamp |

## Retrieval Flow

1. Build deterministic base memory:
   - short-term memory;
   - active long-term facts;
   - current upload/task/customer context.
2. Build retrieval query:
   - current customer name;
   - latest customer message snippets;
   - current task type;
   - main concerns;
   - user instruction if regenerating.
3. Apply hard filters before vector search:
   - tenant/customer visibility;
   - source status is active;
   - rejected facts are excluded;
   - deleted/blocked vectors are excluded.
4. Run vector search:
   - retrieve top 30-80 candidates depending on source type.
5. Rerank and deduplicate:
   - prefer recent same-customer context;
   - prefer copied scripts over skipped scripts;
   - prefer human-corrected facts over agent-only facts;
   - dedupe by `source_type + source_id`.
6. Select prompt context:
   - usually top 5-12 compact snippets;
   - keep each snippet source-labeled;
   - cap total retrieved context tokens.
7. Record recall trace:
   - raw candidate count;
   - selected items;
   - scores and filter reasons.
8. Pass context into `agent.RunInput`.

## Prompt Assembly

The Agent prompt should keep retrieval separate from deterministic memory:

```text
Short-term memory:
...

Long-term memory:
...

Retrieved memory:
- [conversation_message/msg_x, score 0.82] Customer mentioned a 5000/month budget.
- [knowledge_article/kb_12, score 0.78] For price-sensitive teams, explain ROI before discount.

Current input:
...
```

This separation is important. If output quality drops, evaluation can identify whether the issue came from recent context, durable facts, retrieved fuzzy context, or the current user request.

## Ranking Rules

Initial score:

```text
final_score =
  semantic_score * 0.55
  + recency_score * 0.15
  + source_quality_score * 0.20
  + same_customer_score * 0.10
```

Suggested source quality scores:

| Source | Score |
| --- | --- |
| human-corrected memory fact | 1.00 |
| copied follow-up script | 0.90 |
| active long-term fact from AgentRun | 0.80 |
| recent same-customer conversation | 0.75 |
| product/SOP knowledge | 0.70 |
| skipped or marked-wrong task | 0.20 by default, excluded from prompt |

These weights are starting values. They must be measured and tuned with recall evaluation.

## Recall Quality Metrics

Recall quality should be measured before adding more complexity.

| Metric | Meaning | Target For MVP |
| --- | --- | --- |
| `recall@k` | Whether expected source appears in top k | Track manually first |
| `selected_relevance_rate` | Human review score for selected snippets | > 80% useful |
| `bad_context_rate` | Snippets that should not have been shown | < 5% |
| `rejected_fact_leak_rate` | Rejected facts appearing in retrieval | 0 |
| `copy_after_recall_rate` | User copies script after retrieved context was used | Trend upward |
| `mark_wrong_after_recall_rate` | User marks output wrong after recall | Trend downward |
| `p95_recall_latency` | Retrieval latency before prompt assembly | < 300 ms local MVP |
| `embedding_job_lag` | Time from source change to vector availability | < 5 min MVP async |

## Evaluation Dataset

Create `recall_eval_cases` later with:

- customer id;
- query text;
- expected source ids;
- sources that must not appear;
- notes from product/QA;
- result snapshots for each retrieval version.

Early evaluation can be manual JSON fixtures. Do not build a large evaluation platform too early.

## Async Indexing

Embedding and vector indexing must be async before production.

Write-trigger examples:

| Event | Job |
| --- | --- |
| conversation message confirmed | index conversation message |
| long-term fact active/corrected | index memory fact |
| long-term fact rejected | mark vector blocked/stale |
| follow-up task copied | index positive script example |
| follow-up task marked wrong | update source quality signal |
| knowledge article changed | re-index article chunks |

The request path can enqueue jobs, but it should not wait for embeddings.

## Provider Strategy

Stage 20 does not connect a real provider.

When implementation begins:

| Layer | Recommended First Choice |
| --- | --- |
| Embedding provider | Provider-neutral interface + mock local embedding in tests |
| Vector store | Start with adapter interface; local MySQL metadata plus external vector DB later |
| Production vector DB | Evaluate Qdrant, Milvus, pgvector, or managed provider based on deployment constraints |
| LLM framework | Do not add LangChain/LlamaIndex unless they reduce real complexity in Go |

Why not immediately use LangChain/LlamaIndex:

- The backend is Go-first, and the current architecture already owns prompt versions, runner boundaries, and structured validation.
- Most framework value is orchestration glue; this project needs strict auditability, permission filters, and source traceability.
- A thin provider-neutral interface is easier to test and replace at MVP stage.
- Frameworks can still be introduced later for specific capabilities, but should not own core business flow.

## Security And Governance

- Retrieval must enforce the same customer visibility rules as customer detail.
- Rejected facts must be excluded even if their old vector still exists.
- Every prompt snippet must include source metadata.
- Do not index secrets, credentials, or raw private files without explicit classification.
- Do not send customer data to external embedding providers without tenant-level approval and retention policy.
- Keep provider keys in environment variables only.
- Store embedding model versions for re-indexing and evaluation comparisons.

## Failure Behavior

| Failure | Behavior |
| --- | --- |
| Embedding provider unavailable | Queue job retry; user workflow continues |
| Vector store unavailable | Agent uses short-term + long-term memory only |
| Reranker error | Fall back to raw filtered top-k or no retrieved memory |
| Recall timeout | Skip retrieval and record recall trace with timeout |
| Source permission mismatch | Drop candidate before prompt assembly |

The Agent must never fail the sales workflow just because vector recall is unavailable.

## Implementation Plan

### Stage 21: Real LLM Boundary

Before vector search, connect the real LLM safely:

- provider client under `internal/integration/llm`;
- environment-based keys;
- structured JSON output parsing;
- timeout and fallback behavior;
- AgentRun trace updates.

### Stage 22: Embedding Boundary

- add `internal/integration/embedding`;
- add mock deterministic embedding for tests;
- add embedding job model;
- add vector metadata tables;
- add tests that do not require external providers.

### Stage 23: Vector Store Adapter

- define vector upsert/search interface;
- add local fake vector store for tests;
- evaluate Qdrant / Milvus / pgvector / managed option;
- keep permission and status filters in service/store layer.

### Stage 24: Recall Injection

- add `RetrievedMemoryContext` to `agent.RunInput`;
- build retrieval query from current task/customer;
- inject selected snippets under `Retrieved memory`;
- record `recall_traces`.

### Stage 25: Recall Evaluation And Load Test

- add manual eval cases;
- record recall@k and bad context rate;
- run read-path and recall-path load tests;
- tune ranking weights.

## Open Technical Decisions

| Decision | Default For Now |
| --- | --- |
| Vector DB | Keep adapter first; choose after deployment constraints are known |
| Embedding model | Mock for tests; real provider decided when LLM keys are introduced |
| Reranker | Start with weighted rule reranker; add model reranker only if needed |
| Cross-customer recall | Disabled for sales users; manager/admin knowledge recall can be designed later |
| Knowledge base ingestion | Defer until customer memory recall is stable |

## Done Criteria For First Implementation

- Rejected long-term facts never appear in retrieved memory.
- Every retrieved snippet has source metadata.
- Retrieval can be disabled without breaking Agent generation.
- Tests pass without external vector DB or real embedding provider.
- MySQL integration tests cover vector metadata and recall trace persistence.
- Load test records p95 recall latency.
- The user can inspect why a retrieved memory was used.

# Short-Term Memory

Status: implemented for local API context assembly

Short-term memory is the bounded, current customer context used before an Agent generates or regenerates a recommendation. It is not a vector database and it is not the final long-term memory system. It is a deterministic context builder that reads recent structured data and turns it into a prompt-ready summary.

## API

```text
GET /api/customers/{customer_id}/short-term-memory
```

The endpoint returns:

- `customer`: the visible customer profile.
- `conversation_highlights`: recent conversation messages, compacted to bounded snippets.
- `recent_tasks`: recent follow-up task state and recommendation evidence.
- `recent_agent_runs`: recent AgentRun trace summaries, including prompt version and validation information.
- `recent_events`: recent audit events related to the customer or their recent tasks.
- `prompt_context`: a compact text block that can be passed into prompt generation.
- `built_at`: the server timestamp when the memory was assembled.

## Design Rules

- Keep reads bounded. Every source uses explicit page size limits.
- Keep permissions identical to customer detail: sales can only read their own customers, managers can read all customers.
- Do not store full raw chat transcripts in memory output. Use compact highlights and evidence snippets.
- Do not call an external LLM while building short-term memory.
- Do not introduce vector storage at this stage.
- Treat short-term memory as current context, not durable facts.

## Data Sources

Current implementation reads:

```text
customers
conversation_messages
followup_tasks
agent_runs
audit_events
```

Audit events are included from direct `customer` events and recent `followup_task` events. Upload-level events can be connected more deeply later once upload, conversation, and customer relationship queries are expanded.

## Why This Comes Before Vector DB

Vector search is useful for fuzzy recall across large history, product knowledge, cases, and sales SOP material. It should not be the first memory primitive because early Agent behavior needs a reliable, inspectable current context first.

Short-term memory gives the project:

- a stable context contract for prompts;
- a simple way to debug what the Agent saw;
- deterministic tests before recall ranking exists;
- permission boundaries before embeddings and retrieval are added;
- a smaller prompt surface that avoids leaking unrelated customer data.

## Difference From Long-Term Memory

Short-term memory is assembled on demand from recent operational data. It answers: "What does the Agent need to know right now for this customer?"

Long-term memory will persist durable facts and preferences. It should answer: "What have we learned over time that remains useful across future interactions?"

Examples of future long-term memory:

- stable customer preferences;
- purchase constraints;
- objections that repeatedly appear;
- successful scripts and feedback patterns;
- relationship history milestones.

## Next Step

The next backend step is to inject this short-term memory into prompt generation, so AgentRun input summaries and LLM prompts are based on the same auditable context returned by the API.

# Long-Term Memory

Status: local durable fact store implemented

Long-term memory stores stable customer facts that can survive beyond the current conversation window. It is not a vector database and it must not become an unbounded chat transcript store.

## API

```text
GET /api/customers/{customer_id}/long-term-memory
POST /api/customers/{customer_id}/long-term-memory/facts/{fact_id}/reject
POST /api/customers/{customer_id}/long-term-memory/facts/{fact_id}/correct
```

The endpoint returns:

- `customer`: the visible customer profile.
- `facts`: active durable facts for the customer.
- `prompt_context`: compact fact context for later prompt assembly.
- `built_at`: server timestamp when the memory was assembled.

## Storage

Facts are stored in `customer_memory_facts`.

Important fields:

- `customer_id`: tenant and permission anchor.
- `category`: broad fact class, such as `profile`, `concern`, `risk`, or `sales`.
- `fact_key`: stable key inside the category.
- `fact_value`: compact fact value.
- `confidence`: local confidence score.
- `source_type` and `source_id`: provenance, currently `agent_run/{id}`.
- `status`: `active`, `superseded`, or `rejected`.

The unique key is:

```text
customer_id + category + fact_key
```

This keeps one active memory slot per stable fact and prevents the same idea from piling up endlessly.

## Current Write Path

After an upload is confirmed, the service persists deterministic facts from the structured Agent recommendation:

- customer stage;
- intent level;
- main concerns;
- recommended action;
- risk flags.

These facts use the AgentRun as their source. That means a future correction can trace a fact back to the model output and prompt context that produced it.

## Guardrails

- Do not store full chat transcripts as long-term facts.
- Do not write facts without a source.
- Do not call external LLMs when persisting local facts.
- Keep facts compact enough to inspect in API responses and AgentRun debugging.
- Keep permissions tied to the customer visibility model.
- Prefer updating stable fact keys over appending duplicates.
- Reject bad facts by changing status to `rejected`; do not physically delete them.
- Preserve the source AgentRun even when a fact is rejected.
- Correct bad facts by superseding the old fact and writing a human-correction active fact.

## Relationship To Short-Term Memory

Short-term memory answers:

```text
What does the Agent need to know right now?
```

Long-term memory answers:

```text
What stable facts have we learned over time?
```

The two should be combined later during prompt assembly, but they should stay separate internally because their lifecycle and error controls are different.

## Relationship To Vector Recall

Vector recall should come later for fuzzy retrieval across product knowledge, case libraries, SOPs, and historical script examples.

Long-term memory should not depend on vector search. It is the durable fact layer that remains inspectable and correctable even if vector ranking changes.

## Next Step

Active long-term memory is injected through `agent.RunInput.LongTermMemoryContext`.

The Agent prompt keeps it separate from short-term memory:

```text
Short-term memory:
...

Long-term memory:
...
```

This separation is intentional. If output quality drops, evaluation can see whether the issue came from recent context, durable facts, or the current user input.

The next step is to design memory correction and rejection flows, so users can mark a durable fact as wrong without deleting the source AgentRun.
## Rejection Flow

Rejected facts are excluded from `GET /api/customers/{customer_id}/long-term-memory` because the read path only returns active facts.

Rejecting a fact writes an audit event:

```text
memory_fact.rejected
```

The event records:

- actor;
- request id;
- memory fact id;
- customer id;
- optional rejection reason.

This keeps memory governance auditable while preventing rejected facts from entering later prompt context.

## Correction Flow

Correction is for cases where the user knows the right fact and wants the Agent to use it later.

Flow:

```text
old fact -> superseded
new fact -> active
new source_type -> human_correction
new source_id -> old fact id
audit event -> memory_fact.corrected
```

If the correction keeps the same `category + key`, the same row is updated in place because the table enforces one slot per stable fact key. If the correction changes the key, the old row becomes `superseded` and the new row becomes `active`.

This keeps prompt context clean while preserving the correction trail.

## Vector Recall Boundary

The vector recall design is documented in `docs/architecture/memory-retrieval.md`.

Long-term memory remains the durable, inspectable, and correctable fact layer. Vector recall may index active facts for fuzzy search later, but rejected facts must never re-enter prompt context through vector search.

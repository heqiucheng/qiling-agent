# LLM Integration Boundary

Status: OpenAI-compatible provider boundary implemented  
Date: 2026-06-01

## Goal

Qiling Agent must not scatter model calls, prompt strings, or output parsing across handlers, services, or repositories. The LLM boundary keeps model integration replaceable and testable.

Default local behavior still uses mock mode. Real Codex/OpenAI-compatible calls are enabled only when `QILING_LLM_PROVIDER=openai_compatible` and the configured API key environment variable is present.

## Package Boundaries

```text
internal/agent
```

Owns:

- Prompt versions.
- Prompt templates.
- Output schema.
- Runner interface.
- Structured output validation.
- Conversion from business input to LLM request shape.

```text
internal/integration/llm
```

Owns:

- Provider-neutral LLM client interface.
- Request and response DTOs.
- Mock client for deterministic local tests.
- Future provider clients, such as OpenAI or other vendors.

```text
internal/service
```

Owns:

- Business orchestration.
- Calling `agent.Runner`.
- Passing structured Agent output to the store.

```text
internal/store
```

Owns:

- Persisting AgentRun records.
- Persisting customers, tasks, uploads, and audit events in transactions.
- No prompt assembly.
- No model calls.

## Prompt Versions

Current versions:

```text
followup_v1
followup_regenerate_v1
```

Prompt versions are stable identifiers. Changing prompt behavior should create a new version when it can affect output quality, sales compliance, user trust, or evaluation comparability.

## Prompt Context

Agent calls receive business context through `agent.RunInput`.

Current prompt inputs:

```text
CustomerName
OwnerID
RawContent
ShortTermMemoryContext
LongTermMemoryContext
Instruction
ExistingTask
Now
```

Memory context is produced by the service layer and injected into the user prompt as separate sections:

```text
Short-term memory
Long-term memory
```

The Agent package owns the final prompt assembly, so handlers and repositories still do not build prompts.

AgentRun `input_summary` includes a compact memory-aware summary. This gives later debugging and evaluation a trace of the context used for generation without storing unbounded full transcripts.

## Output Schema

The follow-up recommendation output must include:

```text
customer_stage
intent_level
main_concerns
recommended_action
script
reasoning
risk_flags
next_followup_time
```

Validation currently checks required fields before results are recorded as AgentRun output.

## Parsing and Fallback

LLM output must be parsed as JSON before it can become a business recommendation.

Failure cases:

- Empty output.
- Invalid JSON.
- Missing required fields.
- Provider/client error.

Current behavior:

- The runner falls back to deterministic local output.
- The user workflow continues.
- Validation or provider errors are preserved in `validation_errors`.
- The generated AgentRun still records model, prompt version, input summary, output, risk flags, and validation errors.

This is intentional. A model failure should not break the sales workflow, but it must remain visible for debugging, evaluation, and prompt iteration.

## Configuration

Local defaults:

```text
QILING_LLM_PROVIDER=mock
QILING_LLM_MODEL=mock-local-v1
```

Codex proxy mode:

```text
QILING_LLM_PROVIDER=openai_compatible
QILING_LLM_MODEL=gpt-5.4
QILING_LLM_BASE_URL=https://api.claudecode.net.cn/api/codex/backend-api/codex
QILING_LLM_API_KEY_ENV=AICODEMIRROR_API_KEY
AICODEMIRROR_API_KEY=<local secret only>
```

Rules:

- Do not commit provider API keys.
- Do not require real LLM credentials in CI.
- Do not call paid/external LLM APIs without explicit user approval.
- Real provider clients must have local mock tests before business code uses them.
- Provider API keys must be read from environment variables and must never be committed.

## Implemented Provider Client

`internal/integration/llm.OpenAICompatibleClient` calls an OpenAI-compatible Chat Completions endpoint:

```text
POST {QILING_LLM_BASE_URL}/v1/chat/completions
Authorization: Bearer {AICODEMIRROR_API_KEY}
```

The request uses:

- `model` from `QILING_LLM_MODEL`;
- system/user `messages` from `agent.BuildGenerateRequest`;
- `response_format: {"type":"json_object"}`;
- metadata for task type and memory-context presence.

The response content is still parsed by `agent.ParseRecommendation`, then validated before it can become a business recommendation. Provider failures, invalid JSON, or missing required fields fall back to deterministic local output and are recorded in AgentRun validation errors.

## Next Implementation Step

Next steps:

1. Run a real-provider smoke test with a local `AICODEMIRROR_API_KEY`.
2. Confirm which Codex models the proxy account can call.
3. Tune prompt wording if the provider returns non-JSON despite `response_format`.
4. Add model routing after the single-model path is stable.
5. Add usage/cost logging if the proxy exposes token usage.

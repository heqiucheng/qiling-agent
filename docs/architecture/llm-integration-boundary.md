# LLM Integration Boundary

Status: pre-integration foundation  
Date: 2026-06-01

## Goal

Qiling Agent must not scatter model calls, prompt strings, or output parsing across handlers, services, or repositories. The LLM boundary keeps model integration replaceable and testable.

Current stage uses only local mock behavior. No real external model API is called.

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

## Configuration

Local defaults:

```text
QILING_LLM_PROVIDER=mock
QILING_LLM_MODEL=mock-local-v1
```

Rules:

- Do not commit provider API keys.
- Do not require real LLM credentials in CI.
- Do not call paid/external LLM APIs without explicit user approval.
- Real provider clients must have local mock tests before business code uses them.

## Next Implementation Step

When moving from mock to real LLM:

1. Add a provider client under `internal/integration/llm`.
2. Keep the `llm.Client` interface stable.
3. Parse model output into `domain.AgentRecommendation`.
4. Validate schema before creating AgentRun output.
5. Record provider/model/prompt version in AgentRun.
6. Add failure handling and fallback behavior before enabling user-facing traffic.

# Audit Events and Agent Event Stream

Status: MVP foundation  
Date: 2026-06-01

## Purpose

Audit events are the first durable event stream for Qiling Agent. They record the important business actions that happen around uploaded conversations, generated follow-up tasks, user feedback, and future Agent execution.

This is not just an operation log. It is the source trail for later customer profile review, sales behavior analysis, short-term memory, long-term memory, vector indexing jobs, and recall quality evaluation.

## Current Scope

The MVP writes audit events for:

- Conversation upload created.
- Upload result confirmed.
- Follow-up script copied.
- Follow-up task skipped.
- Follow-up task marked wrong.
- Follow-up task regenerated.

The MVP also exposes a paginated read API:

```text
GET /api/audit-events
```

Supported filters:

```text
action
actor_id
entity_type
entity_id
page
page_size
```

Sales users only read their own events. Managers can read all events or filter by `actor_id`.

## Data Model

Table:

```text
audit_events
```

Important fields:

- `action`: stable event name, such as `followup_task.copied`.
- `actor_user_id` and `actor_role`: who triggered the action.
- `request_id`: links the event to HTTP access logs.
- `entity_type` and `entity_id`: primary business object.
- `related_type` and `related_id`: optional secondary object, such as customer or AgentRun.
- `metadata`: compact structured context.
- `created_at`: event time.

The table is indexed by action, actor, entity, and related entity so review and trace queries do not require full table scans.

## Design Rules

Audit metadata must stay compact and structured. Do not store full chat transcripts, secrets, credentials, or large AI outputs inside `metadata`.

Business writes should record audit events only after the business action succeeds. A failed validation or forbidden request is still visible in access logs, but it should not be treated as a completed business event.

Agent features that create or transform customer knowledge must be traceable back to source events. A later customer profile statement should be able to explain which upload, task action, user feedback, or AgentRun contributed to it.

## Why This Matters for Agent Features

Short-term memory needs the latest task and conversation actions.

Long-term memory needs durable, clean, deduplicated facts about customers, sales feedback, and accepted or rejected scripts.

Vector search needs source events so every retrieved memory can show evidence instead of becoming black-box text.

Recall evaluation needs event outcomes. For example, if a regenerated script is copied after user feedback, that is a signal that the new prompt or retrieval result may be better.

Review reports need an evidence chain. A manager should see not only "this customer is high risk", but also which actions and feedback created that conclusion.

## Verification

Local deterministic checks:

```powershell
.\scripts\check.ps1
```

Opt-in MySQL integration check:

```powershell
cd backend
$env:QILING_INTEGRATION_DATABASE_URL='root:<password>@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local'
go test ./internal/store -run MySQLRepository -count=1 -v
```

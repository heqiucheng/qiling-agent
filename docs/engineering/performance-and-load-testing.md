# Performance And Load Testing

This document defines the executable performance workflow for Qiling Agent.

## Goals

The backend must be designed and verified as an enterprise service, not only as a demo API.

Baseline targets for MVP:

| Path type | Target |
| --- | --- |
| Normal list APIs | P95 < 300 ms |
| Detail APIs | P95 < 500 ms |
| Upload acceptance | P95 < 800 ms before async parsing is introduced |
| Error rate | < 1% under local baseline load |

AI generation, vector indexing, WeCom sync, statistics aggregation, and memory rebuild jobs must move to async workers before production use.

## Required Metrics

Every load test report must record:

- `P50`, `P95`, `P99`, and max latency.
- Total requests, success count, error count, and request rate.
- HTTP status code distribution.
- Test duration, concurrency, scenario, and target base URL.
- Manual notes for CPU, memory, slow SQL, and queue backlog when those observability hooks exist.

The backend access log must also include `request_id`, HTTP method, path, status, `duration_ms`, slow-request marker, `user_id`, and role, so client-side load-test metrics can be matched with server-side request traces.

## Executable Tool

Start the backend first:

```powershell
cd backend
$env:QILING_STORE_DRIVER="mysql"
$env:QILING_DATABASE_URL="root:your_password@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local"
go run ./cmd/server
```

Run the default read-path load test from another terminal:

```powershell
.\scripts\loadtest.ps1
```

Override parameters with environment variables:

```powershell
$env:QILING_LOADTEST_BASE_URL="http://127.0.0.1:8080"
$env:QILING_LOADTEST_DURATION="60s"
$env:QILING_LOADTEST_CONCURRENCY="32"
$env:QILING_LOADTEST_SCENARIO="read"
.\scripts\loadtest.ps1
```

Available scenarios:

| Scenario | Behavior | Data impact |
| --- | --- | --- |
| `read` | Exercises health, dashboard, customers, detail, conversations, tasks, and review summary. | No writes |
| `upload` | Repeatedly posts uploaded conversation records. | Writes test upload rows |

Use `upload` only against local or disposable environments until a reset command is available.

After running write-heavy tests locally, reset demo data with:

```powershell
cd backend
go run ./cmd/dbreset -confirm qiling_agent
```

The reset command drops Qiling development tables and reruns migrations. Never run it against production or shared databases.

## When To Run

Run read-path load tests after:

- Adding or changing repository queries.
- Adding list, detail, statistics, or review APIs.
- Changing pagination or filter logic.
- Introducing cache, async jobs, vector retrieval, or memory recall.

Run upload-path load tests after:

- Changing upload parsing.
- Changing customer/profile generation.
- Changing task generation.
- Adding async queue behavior.

## Quality Gate

`scripts/check.ps1` remains deterministic and does not run load tests by default. Load tests are required before merging performance-sensitive backend, AI, memory, vector, recall, or queue changes.

## Local Baselines

These numbers are local development baselines, not production capacity guarantees. They are useful for regression comparison after backend query, logging, memory, vector, recall, or queue changes.

| Date | Store | Scenario | Duration | Concurrency | Requests | Errors | RPS | P50 | P95 | P99 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-06-01 | MySQL local demo data | read | 10s | 8 | 34,414 | 0 | 3,441.03 | 1.754ms | 5.495ms | 8.54ms |

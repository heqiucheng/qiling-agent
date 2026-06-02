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
| `report` | Prepares one customer-intent report, then exercises report list, detail, report generation, Markdown export, and XLSX export. | Writes saved report rows |

Use `upload` and `report` only against local or disposable environments until a reset command is available.

Run the report/export baseline:

```powershell
$env:QILING_LOADTEST_SCENARIO="report"
$env:QILING_LOADTEST_DURATION="10s"
$env:QILING_LOADTEST_CONCURRENCY="8"
.\scripts\loadtest.ps1
```

The load-test output includes both overall latency and `by_endpoint` latency. Use `by_endpoint` to spot slow exporters, especially `format=xlsx`, instead of relying only on total P95.

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
| 2026-06-02 | Mock store local | report | 5s | 4 | 12,548 | 0 | 2,508.01 | 0.523ms | 5.51ms | 7.511ms |
| 2026-06-02 | MySQL local demo data | report | 5s | 4 | 4,432 | 0 | 885.79 | 2.08ms | 13.255ms | 16.598ms |

Report scenario endpoint baseline from 2026-06-02 mock-store run:

| Endpoint | Requests | Errors | P50 | P95 | P99 | Max |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/reports/{id}` | 2,510 | 0 | 0ms | 1.039ms | 2.183ms | 4.993ms |
| `GET /api/reports/{id}/export?format=markdown` | 2,510 | 0 | 0.518ms | 1.674ms | 3.224ms | 8.687ms |
| `GET /api/reports/{id}/export?format=xlsx` | 2,510 | 0 | 4.454ms | 7.499ms | 9.348ms | 14.117ms |
| `GET /api/reports?page=1&page_size=20` | 2,509 | 0 | 0.518ms | 1.555ms | 3.066ms | 4.448ms |
| `POST /api/reports/customer-intent` | 2,509 | 0 | 0.516ms | 1.173ms | 2.61ms | 5.681ms |

Report scenario endpoint baseline from 2026-06-02 MySQL local run:

| Endpoint | Requests | Errors | P50 | P95 | P99 | Max |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/reports/{id}` | 887 | 0 | 1.039ms | 2.714ms | 4.291ms | 14.595ms |
| `GET /api/reports/{id}/export?format=markdown` | 887 | 0 | 1.056ms | 3.19ms | 5.407ms | 10.814ms |
| `GET /api/reports/{id}/export?format=xlsx` | 886 | 0 | 4.845ms | 8.573ms | 10.625ms | 18.77ms |
| `GET /api/reports?page=1&page_size=20` | 886 | 0 | 1.534ms | 3.806ms | 5.783ms | 10.067ms |
| `POST /api/reports/customer-intent` | 886 | 0 | 11.409ms | 16.589ms | 21.144ms | 34.078ms |

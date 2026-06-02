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
| `report` | Prepares one customer-intent report, then exercises report list, detail, report generation, Markdown export, XLSX export, and PDF export. | Writes saved report rows |
| `report-read` | Prepares one customer-intent report, then exercises report list and detail only. | Writes one setup report |
| `report-export` | Prepares one customer-intent report, then exercises Markdown, XLSX, and PDF exports only. | Writes one setup report |
| `report-generate` | Repeatedly generates customer-intent reports. | Writes saved report rows |

Use `upload` and `report` only against local or disposable environments until a reset command is available.

Run the report/export baseline:

```powershell
$env:QILING_LOADTEST_SCENARIO="report"
$env:QILING_LOADTEST_DURATION="10s"
$env:QILING_LOADTEST_CONCURRENCY="8"
.\scripts\loadtest.ps1
```

The load-test output includes both overall latency and `by_endpoint` latency. Use `by_endpoint` to spot slow exporters, especially `format=xlsx` and `format=pdf`, instead of relying only on total P95.

Use the split report scenarios when diagnosing regressions:

```powershell
$env:QILING_LOADTEST_SCENARIO="report-export"
.\scripts\loadtest.ps1
```

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
| 2026-06-02 | MySQL local demo data with PDF export | report | 5s | 4 | 989 | 0 | 195.23 | 4.13ms | 92.999ms | 135.656ms |
| 2026-06-02 | MySQL local demo data with export cache | report | 5s | 4 | 5,249 | 0 | 1,045.33 | 1.111ms | 13.578ms | 19.896ms |

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

Report scenario endpoint baseline from 2026-06-02 MySQL local run after adding PDF export:

| Endpoint | Requests | Errors | P50 | P95 | P99 | Max |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/reports/{id}` | 165 | 0 | 1.135ms | 3.857ms | 4.839ms | 7.579ms |
| `GET /api/reports/{id}/export?format=markdown` | 165 | 0 | 1.553ms | 4.622ms | 7.842ms | 11.881ms |
| `GET /api/reports/{id}/export?format=pdf` | 165 | 0 | 85.796ms | 136.733ms | 152.198ms | 159.095ms |
| `GET /api/reports/{id}/export?format=xlsx` | 165 | 0 | 5.24ms | 9.712ms | 13.007ms | 17.619ms |
| `GET /api/reports?page=1&page_size=20` | 164 | 0 | 1.647ms | 5.074ms | 8.981ms | 14.73ms |
| `POST /api/reports/customer-intent` | 165 | 0 | 13.427ms | 24.193ms | 30.784ms | 110.471ms |

PDF export is now the heaviest report endpoint in local baseline testing. Keep it synchronous for MVP, but revisit async export jobs or file caching before production workloads.

Report scenario endpoint baseline from 2026-06-02 MySQL local run after adding in-process export cache:

| Endpoint | Requests | Errors | P50 | P95 | P99 | Max |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/reports/{id}` | 875 | 0 | 1.047ms | 3.012ms | 4.923ms | 16.841ms |
| `GET /api/reports/{id}/export?format=markdown` | 875 | 0 | 1.09ms | 3.974ms | 7.479ms | 13.739ms |
| `GET /api/reports/{id}/export?format=pdf` | 875 | 0 | 1.049ms | 3.04ms | 7.544ms | 103.838ms |
| `GET /api/reports/{id}/export?format=xlsx` | 875 | 0 | 1.045ms | 2.767ms | 5.301ms | 10.665ms |
| `GET /api/reports?page=1&page_size=20` | 874 | 0 | 1.557ms | 3.726ms | 6.183ms | 21.263ms |
| `POST /api/reports/customer-intent` | 875 | 0 | 11.784ms | 20.225ms | 31.508ms | 52.512ms |

The export cache brought PDF P95 from 136.733ms to 3.04ms for repeated downloads of the same report. The max value still includes first-render cost, so production should still consider shared artifact storage or async export jobs for cold exports.

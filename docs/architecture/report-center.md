# Report Center Architecture

Status: MVP design  
Date: 2026-06-01

Updated: 2026-06-02 - Report history persistence, export cache, and export task records are implemented.

## Goal

Report Center turns operational data into structured, shareable reports. It should reuse existing customer visibility, AgentRun, memory, and task data instead of creating a parallel analytics system.

## Package Boundary

MVP adds a focused report module:

```text
backend/internal/report/
  types.go              // report domain types
  service.go            // report orchestration
  markdown.go           // Markdown rendering
```

HTTP handlers stay in:

```text
backend/internal/http/handler/reports.go
```

The main `service.QilingService` can expose report methods at MVP stage because the repository interface already owns customer, task, and AgentRun access. If report logic grows, it can later move behind a dedicated `report.Service`.

## Data Sources

| Source | Use |
|---|---|
| `customers` | Customer name, stage, intent, owner, concerns, tags, risk flags. |
| `followup_tasks` | Latest recommendations, scripts, task status, feedback. |
| `agent_runs` | Model reasoning, validation errors, prompt version, risk flags. |
| `conversation_messages` | Evidence snippets and recent customer phrasing. |
| `customer_memory_facts` | Durable facts after the memory layer is stable. |
| `saved_reports` | Persisted structured report JSON, Markdown, owner scope, summary counts, and generated time. |
| `report_export_tasks` | Persistent export task records, owner scope, output metadata, status, and error details. |

MVP uses customers and follow-up tasks first, with room to add AgentRun and conversation evidence in the next iteration.

## API

### POST `/api/reports/customer-intent`

Request:

```json
{
  "range": "last_7_days",
  "format": "structured"
}
```

Response:

```json
{
  "data": {
    "id": "rpt_customer_intent_20260601",
    "type": "customer_intent",
    "title": "最近 7 天客户意愿分析报告",
    "summary": "...",
    "metrics": [],
    "sections": [],
    "action_items": [],
    "markdown": "..."
  },
  "error": null,
  "meta": {}
}
```

### GET `/api/reports/customer-intent/preview`

Deferred. Preview can later return the current default report without creating a persistent record.

### Export APIs

Markdown, XLSX, DOCX, and PDF exports are implemented:

```text
GET /api/reports/{report_id}/export?format=markdown
GET /api/reports/{report_id}/export?format=docx
GET /api/reports/{report_id}/export?format=pdf
GET /api/reports/{report_id}/export?format=xlsx
```

`format=markdown` returns a file response:

| Header | Value |
|---|---|
| `Content-Type` | `text/markdown; charset=utf-8` |
| `Content-Disposition` | `attachment; filename="{report_id}.md"` |

The report response still includes Markdown inline so users can copy quickly without downloading.

Export format decisions live in `service.ExportReport`. HTTP handlers only translate the export result into response headers and body. This keeps DOCX, PDF, and XLSX support out of the routing layer when those formats are added.

Exports are cached in-process by `report_id + generated_at + format`. The cache stores defensive byte copies so callers cannot mutate cached files. This is an MVP optimization for repeated downloads of the same report, especially PDF. It is not a cross-instance cache; production deployments should move hot export artifacts to object storage or a shared cache if multiple backend instances are used.

### Export Task APIs

Export tasks add a durable Agent action record around report exports:

```text
POST /api/reports/{report_id}/export-tasks
GET /api/report-export-tasks
GET /api/report-export-tasks/{task_id}
```

Request:

```json
{
  "format": "pdf"
}
```

MVP behavior is synchronous: creating a task calls the existing `ExportReport` path immediately, then saves the task as `completed` or `failed`. This keeps the API shape stable while avoiding premature queue infrastructure. The next backend step is to move execution behind a worker:

| Stage | Behavior |
|---|---|
| MVP | HTTP request creates a task, renders export synchronously, stores status and file metadata. |
| Worker stage | HTTP request stores `queued`; worker renders file; task moves to `completed` or `failed`. |
| Artifact stage | Export bytes move to object storage; task stores storage key, checksum, and download URL/expiry. |

Task visibility follows the same owner rule as reports. The frontend must list export tasks from the backend instead of deriving task state locally.

### GET `/api/reports`

Lists persisted reports visible to the current actor. The backend filters by `owner_id` and `owner_role`; the frontend must not request broad report history and filter it locally.

Response shape:

```json
{
  "data": {
    "items": [
      {
        "id": "rpt_customer_intent_20260602",
        "type": "customer_intent",
        "title": "最近 7 天客户意愿分析报告",
        "range_label": "最近 7 天",
        "summary": "...",
        "owner_id": "usr_001",
        "owner_role": "sales",
        "metric_count": 4,
        "section_count": 3,
        "action_item_count": 5,
        "generated_at": "2026-06-02T09:00:00Z"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1
  },
  "error": null,
  "meta": {}
}
```

### GET `/api/reports/{report_id}`

Loads the full persisted report, including structured sections and Markdown. The service checks ownership before returning detail:

- sales users can only read reports generated under their own `user_id` and `role`;
- manager users can read reports generated under their manager actor scope;
- unauthorized detail reads return `FORBIDDEN`.

## Report Types

Report response should be structured first and rendered second:

```text
Report
  ID
  Type
  Title
  RangeLabel
  Summary
  OwnerID
  OwnerRole
  Metrics[]
  Sections[]
  ActionItems[]
  Markdown
  GeneratedAt
```

This avoids coupling frontend layout to a single Markdown blob and keeps future exporters easier.

## Generation Strategy

MVP uses deterministic generation from existing structured data:

1. Filter visible customers by actor.
2. Group customers by intent, stage, and risk.
3. Attach latest follow-up task recommendation when available.
4. Build report sections.
5. Render Markdown from the same structured report.
6. Save the structured report and Markdown to `saved_reports`.

LLM report polishing should be added only after deterministic evidence is reliable. The model should rewrite wording, not invent facts.

## Persistence Rules

`saved_reports` stores both a query-friendly summary and the full `report_json` payload:

| Column group | Why it exists |
|---|---|
| `owner_id`, `owner_role` | Enforces report history visibility without relying on frontend filtering. |
| `metrics_count`, `sections_count`, `action_items_count` | Keeps report list fast and compact. |
| `report_json` | Preserves the full structured report for future UI and exporters. |
| `markdown` | Allows quick copy/export of the exact generated content. |
| `generated_at`, `created_at` | Separates business report time from database insert time. |

The important Agent design point: report generation is not only "answer once". It becomes a durable memory artifact that can be reviewed, copied, compared, and exported later. Think of it as saving the meeting minutes after a discussion instead of asking the Agent to remember everything from scratch each time.

`report_export_tasks` stores the operational trace of report export work:

| Column group | Why it exists |
|---|---|
| `report_id`, `export_format` | Connects the task to the source report and renderer. |
| `status`, `error_message` | Makes failed exports reviewable instead of disappearing inside a request error. |
| `owner_id`, `owner_role` | Keeps task history scoped to the same actor that owns the report. |
| `filename`, `content_type`, `size_bytes` | Records output metadata without forcing the frontend to inspect file bytes. |
| `created_at`, `completed_at` | Supports queue latency, retry, and SLA metrics later. |

## Permission Rules

Report generation must call the same visibility logic used by customer and task lists:

- manager: all customers;
- sales: only own customers.

No report service may query broad data and filter later in the frontend.

## Evidence Rules

Each customer-level recommendation should include at least one evidence string from:

- customer profile summary;
- concern list;
- risk flag;
- latest task recommendation;
- latest AgentRun reasoning;
- recent conversation message.

If evidence is weak, the report should say "信息不足，需要补充确认" instead of overstating intent.

## Frontend

MVP adds:

```text
frontend/src/features/reports/
  ReportCenterPage.tsx
```

Navigation adds:

```text
/app/reports
```

The page should show:

- report type selector;
- generate button;
- persisted report history;
- metrics strip;
- report sections;
- action list;
- copy Markdown button.

No inline styles. Use existing shared components and CSS tokens.

## Testing

Backend:

- report generation respects actor visibility;
- report contains metrics, sections, action items, and Markdown;
- low-information or pending-confirmation customers are not forced into high-intent language.

Frontend:

- page renders generated report;
- copy button has a safe fallback state;
- API mapper covers report response shape.

## Future Exporters

| Exporter | Implementation note |
|---|---|
| Markdown | Implemented as `GET /api/reports/{report_id}/export?format=markdown`. |
| XLSX | Implemented as `GET /api/reports/{report_id}/export?format=xlsx`, with sheets for summary metrics, action items, and customer details. |
| DOCX | Implemented as `GET /api/reports/{report_id}/export?format=docx`, with title, summary, metrics, customer analysis, and action items. |
| PDF | Implemented as `GET /api/reports/{report_id}/export?format=pdf`, using server-side PDF rendering. Configure `QILING_PDF_FONT_PATH` in Linux environments if Chinese text needs a specific font. |

The XLSX workbook is intentionally table-first instead of visually heavy. It is meant for operations review: leaders scan metrics, sales users execute action items, and analysts inspect customer evidence.

## Failure Handling

| Failure | Behavior |
|---|---|
| No visible customers | Return an empty report with explanation and no action items. |
| LLM unavailable | Use deterministic report generation. |
| Export fails | Keep in-app report available and return a clear export error. |
| Permission denied | Return the same authorization behavior as customer list. |

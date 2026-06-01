# Report Center Architecture

Status: MVP design  
Date: 2026-06-01

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

Deferred until report schema is stable:

```text
GET /api/reports/{report_id}/export?format=markdown
GET /api/reports/{report_id}/export?format=docx
GET /api/reports/{report_id}/export?format=pdf
GET /api/reports/{report_id}/export?format=xlsx
```

MVP returns Markdown inline in the report response.

## Report Types

Report response should be structured first and rendered second:

```text
Report
  ID
  Type
  Title
  RangeLabel
  Summary
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

LLM report polishing should be added only after deterministic evidence is reliable. The model should rewrite wording, not invent facts.

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
| Markdown | Current MVP renderer. |
| DOCX | Generate from structured sections, then render-check document layout. |
| PDF | Use HTML print or server-side renderer after layout is stable. |
| XLSX | Export customer rows, metrics, and action items as separate sheets. |

## Failure Handling

| Failure | Behavior |
|---|---|
| No visible customers | Return an empty report with explanation and no action items. |
| LLM unavailable | Use deterministic report generation. |
| Export fails | Keep in-app report available and return a clear export error. |
| Permission denied | Return the same authorization behavior as customer list. |

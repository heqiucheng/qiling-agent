# Report Center PRD

Status: MVP design  
Date: 2026-06-01

## Goal

Qiling Agent should turn customer conversations, AgentRun outputs, follow-up tasks, and memory facts into shareable reports. The first version focuses on reports users can read in the product, copy as Markdown, and use for sales review or manager updates.

This is not only a sales report feature. Reports must support business conversations, work coordination, customer service, life-event context, schedule confirmation, relationship maintenance, and low-information chats.

## MVP Scope

### Report Types

| Report | Primary user | Purpose |
|---|---|---|
| Customer intent report | Sales, manager | Share recent customer willingness, opportunity level, risks, and next actions. |
| Daily follow-up review | Sales | Review today's follow-up quality and tomorrow's action list. |
| Manager weekly report | Manager | Summarize team-level customer opportunities, risks, and suggested priorities. |

MVP implements the customer intent report first. The other two reports share the same architecture and can be added after the first report is stable.

### First Report: Customer Intent Report

The report must include:

- report title and time range;
- summary metrics;
- high-intent customers;
- pending-confirmation customers;
- low-information customers that need clarification;
- risk customers;
- recommended next actions;
- copy-ready follow-up scripts;
- evidence notes from customer profile, AgentRun, follow-up task, or recent conversation context.

### Output Formats

| Format | MVP status | Notes |
|---|---|---|
| In-app report view | Included | Structured report page. |
| Markdown | Included | Copyable and later exportable. |
| DOCX | Deferred | Requires document rendering and visual QA. |
| PDF | Deferred | Requires print layout and export pipeline. |
| XLSX | Deferred | Best for customer lists and metrics after report schema is stable. |

## User Workflow

1. User opens Report Center.
2. User selects "Customer Intent Report".
3. User chooses a time range or uses the default recent 7 days.
4. System gathers visible customers, tasks, AgentRuns, and recent conversation evidence.
5. Agent generates a structured report.
6. User reads the report, copies Markdown, and uses the action list.

MVP can generate synchronously because the dataset is small. Later versions should use background jobs for large tenants.

## Product Rules

| Rule | Requirement |
|---|---|
| Evidence first | Every important conclusion needs evidence text or source context. |
| No hallucinated customers | Report content can only use customers visible to the current actor. |
| Scene-aware judgment | Do not force life or work-coordination chats into sales intent. |
| Low-information caution | Short chats should be marked as pending or needs clarification unless there is strong evidence. |
| Actionable output | Every report must end with a concrete next-action list. |
| Copy-friendly wording | Markdown should be readable in WeChat, Feishu, Notion, and email. |

## Permissions

- Sales users can only generate reports from their own visible customers and tasks.
- Managers can generate reports across all visible team customers.
- Report generation should reuse the same visibility rules as customer list and follow-up task list.

## Success Criteria

| Metric | Target |
|---|---|
| First report generation | Returns in less than 2 seconds on MVP demo data without real LLM. |
| Report usefulness | Includes at least one summary, one customer section, and one action item when data exists. |
| Evidence quality | Key recommendations include evidence snippets. |
| Safety | No report is generated from customers outside actor visibility. |
| Copy quality | Markdown is directly usable without editing headings or broken layout. |

## Out of Scope For MVP

- Scheduled automatic reports.
- Email or WeChat push.
- DOCX/PDF/XLSX export.
- Custom report builder.
- Chart rendering.
- Cross-tenant benchmark analysis.

## Follow-up Roadmap

1. Add daily follow-up review report.
2. Add manager weekly report.
3. Persist generated reports in MySQL.
4. Add report feedback: useful, inaccurate, missing evidence.
5. Add DOCX/PDF/XLSX exporters.
6. Add async report jobs for large datasets.
7. Add custom report templates.

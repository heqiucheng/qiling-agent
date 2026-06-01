# CI Quality Gates

Qiling Agent uses two levels of quality gates.

## Default CI

GitHub Actions runs deterministic checks on every push to `main` and every pull request:

```text
frontend npm ci
frontend npm run lint
frontend npm run typecheck
frontend npm run test
frontend npm run build
backend go test ./...
```

The default CI must not require MySQL, external paid APIs, WeCom credentials, LLM credentials, vector databases, or production data.

## Local Full Check

Use the local wrapper before committing meaningful work:

```powershell
.\scripts\check.ps1
```

This mirrors the deterministic CI gate.

## MySQL Integration Tests

MySQL integration tests are opt-in because they reset the configured database:

```powershell
cd backend
$env:QILING_INTEGRATION_DATABASE_URL="root:your_password@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local"
go test ./internal/store -run MySQLRepository -count=1 -v
```

Rules:

- Run only against local or disposable databases.
- Never point `QILING_INTEGRATION_DATABASE_URL` at production or shared customer databases.
- Use these tests after changing MySQL repository behavior, migrations, upload confirmation, task status transitions, or pagination.

## Performance Tests

Performance tests remain separate from default CI because they depend on machine load, dataset size, and database state.

```powershell
.\scripts\loadtest.ps1
```

Run performance tests before merging performance-sensitive backend, AI, memory, vector, recall, queue, or statistics changes.

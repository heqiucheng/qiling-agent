# Qiling Agent Backend

Go backend for Qiling Agent.

## Commands

```powershell
go test ./...
go run ./cmd/server
```

Default address:

```text
:8080
```

Override with:

```text
QILING_HTTP_ADDR=:9090
QILING_ENV=development
```

## Current API

```text
GET /api/health
```
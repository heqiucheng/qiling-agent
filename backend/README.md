# Qiling Agent Backend

Go backend for Qiling Agent.

## Commands

```powershell
go test ./...
go run ./cmd/server
go run ./cmd/dbping
go run ./cmd/dbmigrate
```

Default address:

```text
:8080
```

Override with:

```text
QILING_HTTP_ADDR=:9090
QILING_ENV=development
QILING_DATABASE_URL=root:change_me@tcp(127.0.0.1:3306)/qiling_agent?parseTime=true&charset=utf8mb4&loc=Local
```

## Current API

```text
GET /api/health
```

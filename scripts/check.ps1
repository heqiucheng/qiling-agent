$ErrorActionPreference = "Stop"

Push-Location "$PSScriptRoot\..\frontend"
npm run lint
npm run typecheck
npm run test
npm run build
Pop-Location

Push-Location "$PSScriptRoot\..\backend"
go test ./...
Pop-Location

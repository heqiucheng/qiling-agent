$ErrorActionPreference = "Stop"

$BaseUrl = if ($env:QILING_LOADTEST_BASE_URL) { $env:QILING_LOADTEST_BASE_URL } else { "http://127.0.0.1:8080" }
$Duration = if ($env:QILING_LOADTEST_DURATION) { $env:QILING_LOADTEST_DURATION } else { "30s" }
$Concurrency = if ($env:QILING_LOADTEST_CONCURRENCY) { $env:QILING_LOADTEST_CONCURRENCY } else { "16" }
$Scenario = if ($env:QILING_LOADTEST_SCENARIO) { $env:QILING_LOADTEST_SCENARIO } else { "read" }

Push-Location "$PSScriptRoot\..\backend"
go run ./cmd/loadtest `
  -base-url "$BaseUrl" `
  -duration "$Duration" `
  -concurrency "$Concurrency" `
  -scenario "$Scenario"
Pop-Location

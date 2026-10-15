# ==============================================================================
# Search & Infinite Scroll Benchmark Script
# ==============================================================================
param(
    [string]$BaseUrl = "http://localhost:8080/api/v1",
    [int]$Iterations = 30
)

Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "   Search & Infinite Scroll Benchmark        " -ForegroundColor Cyan
Write-Host "   Target: $BaseUrl                          " -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

function Measure-Endpoint {
    param(
        [string]$Name,
        [string]$Url
    )
    $times = @()
    for ($i = 0; $i -lt $Iterations; $i++) {
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        try {
            $resp = Invoke-RestMethod -Uri $Url -Method Get -TimeoutSec 5
            $sw.Stop()
            $times += $sw.ElapsedMilliseconds
        } catch {
            $sw.Stop()
            Write-Host "Error on ${Url}: $_" -ForegroundColor Red
        }
    }

    if ($times.Count -gt 0) {
        $sorted = $times | Sort-Object
        $p50 = $sorted[[Math]::Floor($sorted.Count * 0.50)]
        $p90 = $sorted[[Math]::Floor($sorted.Count * 0.90)]
        $p99 = $sorted[[Math]::Floor($sorted.Count * 0.99)]
        $avg = ($times | Measure-Object -Average).Average

        Write-Host ("[{0}] Iterations: {1} | Avg: {2:N1}ms | p50: {3}ms | p90: {4}ms | p99: {5}ms" -f $Name, $times.Count, $avg, $p50, $p90, $p99) -ForegroundColor Green
    }
}

# 1. Autocomplete Suggestions Benchmark
Measure-Endpoint -Name "Autocomplete (/search/suggestions?q=tech)" -Url "$BaseUrl/search/suggestions?q=tech"

# 2. Search Posts Top Tab Benchmark
Measure-Endpoint -Name "Search Posts Top Tab (/search/posts?q=golang&tab=top)" -Url "$BaseUrl/search/posts?q=golang&tab=top"

# 3. Search Posts Latest Tab Keyset Pagination Benchmark
Measure-Endpoint -Name "Search Posts Latest Tab (/search/posts?q=golang&tab=latest)" -Url "$BaseUrl/search/posts?q=golang&tab=latest"

# 4. Search Users Benchmark
Measure-Endpoint -Name "Search Users (/search/users?q=tech)" -Url "$BaseUrl/search/users?q=tech"

Write-Host "`nBenchmark completed successfully!" -ForegroundColor Cyan

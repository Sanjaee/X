param(
    [string]$TargetUrl = "http://localhost",
    [int]$TotalRequests = 300,
    [int]$Concurrency = 15,
    [switch]$CheckMetrics
)

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  X-PLATFORM TRAFFIC SPIKE & RESILIENCY LOAD TEST" -ForegroundColor Cyan
Write-Host "  Target: $TargetUrl" -ForegroundColor Cyan
Write-Host "  Concurrency: $Concurrency threads | Total: $TotalRequests requests" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Fetch Baseline Health & Metrics
Write-Host "`n[1/3] Querying initial system metrics..." -ForegroundColor Yellow
try {
    $preMetrics = Invoke-RestMethod -Uri "$TargetUrl/health/metrics" -TimeoutSec 3 -ErrorAction Stop
    Write-Host "  Initial Goroutines: $($preMetrics.goroutines)" -ForegroundColor Gray
    Write-Host "  Initial Heap Alloc: $($preMetrics.alloc_mb)" -ForegroundColor Gray
    Write-Host "  Active Inflight Requests: $($preMetrics.active_requests)" -ForegroundColor Gray
} catch {
    Write-Host "  Could not reach /health/metrics (Server might be starting up): $_" -ForegroundColor DarkYellow
}

# 2. Run Burst Load Test
Write-Host "`n[2/3] Unleashing synthetic traffic burst..." -ForegroundColor Yellow
$startTime = [System.Diagnostics.Stopwatch]::StartNew()

$requestsPerThread = [math]::Ceiling($TotalRequests / $Concurrency)

$scriptBlock = {
    param($url, $count)
    $ok = 0
    $rateLimited = 0
    $backpressured = 0
    $errors = 0
    $endpoints = @(
        "/api/v1/feed?limit=20",
        "/api/v1/search?q=test&limit=10",
        "/api/v1/explore/trending"
    )

    1..$count | ForEach-Object {
        $ep = $endpoints[$_ % $endpoints.Length]
        $reqUrl = "$url$ep"
        try {
            $resp = Invoke-WebRequest -Uri $reqUrl -UseBasicParsing -TimeoutSec 4 -ErrorAction Stop
            if ($resp.StatusCode -eq 200) {
                $ok++
            }
        } catch {
            if ($_.Exception.Response -ne $null) {
                $code = $_.Exception.Response.StatusCode.value__
                if ($code -eq 429) {
                    $rateLimited++
                } elseif ($code -eq 503) {
                    $backpressured++
                } else {
                    $errors++
                }
            } else {
                $errors++
            }
        }
    }

    [PSCustomObject]@{
        OK            = $ok
        RateLimited   = $rateLimited
        Backpressured = $backpressured
        Errors        = $errors
    }
}

$jobs = 1..$Concurrency | ForEach-Object {
    Start-Job -ScriptBlock $scriptBlock -ArgumentList $TargetUrl, $requestsPerThread
}

$results = $jobs | Wait-Job | Receive-Job
$jobs | Remove-Job

$startTime.Stop()
$elapsedSec = $startTime.Elapsed.TotalSeconds

# 3. Aggregate Results
$totalOK = ($results | Measure-Object -Property OK -Sum).Sum
$totalRL = ($results | Measure-Object -Property RateLimited -Sum).Sum
$totalBP = ($results | Measure-Object -Property Backpressured -Sum).Sum
$totalErr = ($results | Measure-Object -Property Errors -Sum).Sum
$totalProcessed = $totalOK + $totalRL + $totalBP + $totalErr
$rps = [math]::Round($totalProcessed / $elapsedSec, 2)

Write-Host "`n[3/3] Load Test Completed!" -ForegroundColor Green
Write-Host "----------------------------------------------------------"
Write-Host "  Elapsed Time       : $([math]::Round($elapsedSec, 2)) seconds"
Write-Host "  Total Requests     : $totalProcessed"
Write-Host "  Throughput         : $rps req/sec"
Write-Host "  HTTP 200 (Success) : $totalOK" -ForegroundColor Green
Write-Host "  HTTP 429 (Throttled): $totalRL" -ForegroundColor Yellow
Write-Host "  HTTP 503 (Backpressure): $totalBP" -ForegroundColor Cyan
Write-Host "  Other Errors       : $totalErr" -ForegroundColor Red
Write-Host "----------------------------------------------------------"

# Post-test metrics inspection
try {
    $postMetrics = Invoke-RestMethod -Uri "$TargetUrl/health/metrics" -TimeoutSec 3 -ErrorAction Stop
    Write-Host "`nPost-Burst Diagnostics:" -ForegroundColor Cyan
    Write-Host "  Active Goroutines  : $($postMetrics.goroutines)"
    Write-Host "  Heap Allocation    : $($postMetrics.alloc_mb)"
    Write-Host "  Total GC Runs      : $($postMetrics.num_gc)"
    Write-Host "  GC Pause Total     : $($postMetrics.gc_pause_total_ms)"
} catch {
    Write-Host "  Could not fetch post-metrics." -ForegroundColor Gray
}

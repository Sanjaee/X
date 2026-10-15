# Capacity & Load Testing Methodology (4GB VPS)

This guide provides realistic load testing procedures to measure actual throughput, latency percentiles, and memory ceilings under simulated traffic spikes.

---

## 1. Principles of Low-Resource Load Testing

1. **Never test production blindly**: Run load tests in staging or during scheduled maintenance.
2. **Observe memory while loading**: Concurrently monitor `docker stats` and `free -m` to observe heap behavior and GC cycles.
3. **Measure saturation, not just success**: Identify the exact request rate where p95 latency degrades or HTTP 429/503 admission control triggers.
4. **Synthetic Data Realism**: Use randomized user IDs and varying query terms to avoid 100% cache hits.

---

## 2. Multi-Stage Load Test Plan

| Stage | Name | Target Load | Duration | Purpose |
|---|---|---|---|---|
| **Stage 1** | Baseline / Idle | 0 VUs | 5 mins | Record quiescent memory footprint of all containers. |
| **Stage 2** | Normal Traffic | 50 concurrent VUs (~100 req/s) | 10 mins | Validate steady-state latency (p95 < 100ms) and DB pool stability. |
| **Stage 3** | High Traffic | 200 concurrent VUs (~350 req/s) | 10 mins | Test Go GC pressure, Redis cache hit rates, connection pooling. |
| **Stage 4** | Traffic Spike | 500 VUs sudden burst | 2 mins | Test admission control, Nginx rate limits, and HTTP 503 backpressure. |
| **Stage 5** | Soak Test | 100 VUs constant | 45 mins | Detect slow memory leaks or uncollected goroutines. |
| **Stage 6** | Recovery Test | Drop to 0 VUs | 5 mins | Verify memory drops back to quiescent levels and pools drain. |

---

## 3. k6 Load Testing Script

Save the following script as `tests/load/k6-spike-test.js`:

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },   // Warm-up to 50 VUs
    { duration: '3m', target: 150 },  // Sustained load
    { duration: '30s', target: 400 }, // Traffic spike
    { duration: '1m', target: 400 },  // Peak stress
    { duration: '1m', target: 50 },   // Cool down
    { duration: '30s', target: 0 },   // Recovery
  ],
  thresholds: {
    http_req_duration: ['p(95)<300', 'p(99)<800'],
    http_req_failed: ['rate<0.02'], // Allow up to 2% errors during massive spike (due to 503 backpressure)
  },
};

const BASE_URL = __ENV.TARGET_URL || 'http://localhost';

export default function () {
  // 1. Home Feed (70% weight)
  const feedRes = http.get(`${BASE_URL}/api/v1/feed?limit=20`);
  check(feedRes, {
    'feed status is 200 or 503': (r) => r.status === 200 || r.status === 503,
  });

  // 2. Search Posts (20% weight)
  if (Math.random() < 0.2) {
    const terms = ['tech', 'news', 'crypto', 'code', 'x', 'ai'];
    const term = terms[Math.floor(Math.random() * terms.length)];
    const searchRes = http.get(`${BASE_URL}/api/v1/search?q=${term}&type=posts&limit=15`);
    check(searchRes, {
      'search status is 200 or 429': (r) => r.status === 200 || r.status === 429,
    });
  }

  // 3. User Profile (10% weight)
  if (Math.random() < 0.1) {
    const profileRes = http.get(`${BASE_URL}/api/v1/users/demo`);
    check(profileRes, {
      'profile status is 200 or 404': (r) => r.status === 200 || r.status === 404,
    });
  }

  sleep(0.1 + Math.random() * 0.2); // Jittered sleep
}
```

---

## 4. Built-In PowerShell Spike Test Script

If `k6` is not installed on the test runner, use the native PowerShell load test script (`scripts/load-test-spike.ps1`):

```powershell
param(
    [string]$TargetUrl = "http://localhost",
    [int]$TotalRequests = 500,
    [int]$Concurrency = 20
)

Write-Host "Starting traffic burst against $TargetUrl with Concurrency=$Concurrency, Total=$TotalRequests..." -ForegroundColor Cyan

$startTime = Get-Date
$success = 0
$rateLimited = 0
$backpressured = 0
$failed = 0

$jobs = 1..$Concurrency | ForEach-Object {
    Start-Job -ScriptBlock {
        param($url, $count)
        $s = 0; $rl = 0; $bp = 0; $f = 0
        1..$count | ForEach-Object {
            try {
                $resp = Invoke-WebRequest -Uri "$url/api/v1/search?q=test&limit=10" -UseBasicParsing -TimeoutSec 3 -ErrorAction Stop
                if ($resp.StatusCode -eq 200) { $s++ }
            } catch {
                if ($_.Exception.Response.StatusCode.value__ -eq 429) { $rl++ }
                elseif ($_.Exception.Response.StatusCode.value__ -eq 503) { $bp++ }
                else { $f++ }
            }
        }
        [PSCustomObject]@{ Success = $s; RateLimited = $rl; Backpressured = $bp; Failed = $f }
    } -ArgumentList $TargetUrl, [math]::Ceiling($TotalRequests / $Concurrency)
}

$results = $jobs | Wait-Job | Receive-Job
$jobs | Remove-Job

$results | ForEach-Object {
    $success += $_.Success
    $rateLimited += $_.RateLimited
    $backpressured += $_.Backpressured
    $failed += $_.Failed
}

$elapsed = (Get-Date) - $startTime
Write-Host "`n--- LOAD TEST RESULTS ---" -ForegroundColor Green
Write-Host "Duration: $($elapsed.TotalSeconds)s"
Write-Host "Total Completed: $($success + $rateLimited + $backpressured + $failed)"
Write-Host "HTTP 200 OK: $success"
Write-Host "HTTP 429 (Rate Limited): $rateLimited"
Write-Host "HTTP 503 (Backpressured): $backpressured"
Write-Host "Other Errors: $failed"
Write-Host "Effective RPS: $([math]::Round(($success + $rateLimited + $backpressured + $failed) / $elapsed.TotalSeconds, 2)) req/sec"
```

---

## 5. Capacity Report Template

Use this template to record performance audit findings:

```markdown
### Load Test Run: [Date & Time]
- **Host Spec**: 4 vCPU, 4096 MB RAM, Ubuntu 22.04 LTS, SSD
- **Software**: Go 1.22, Next.js 14, PostgreSQL 16, Redis 7, RabbitMQ 3-alpine
- **Active Concurrency**: 200 Virtual Users

#### Metrics Observed:
- **Peak API RAM**: 214 MB (Target budget: <350 MB)
- **Peak Postgres RAM**: 482 MB (Target budget: <768 MB)
- **Peak Redis RAM**: 64 MB (Target budget: <180 MB)
- **Total Host Free RAM at Peak**: 1,480 MB (Zero OOM triggers)
- **HTTP p50 Latency**: 24 ms
- **HTTP p95 Latency**: 118 ms
- **HTTP p99 Latency**: 260 ms
- **Backpressure Events**: 0 at 200 VUs; 14 triggered at 450 VU spike.
- **Identified Bottlenecks**: None in database; CPU reached 68% during GIN index FTS search burst.
```

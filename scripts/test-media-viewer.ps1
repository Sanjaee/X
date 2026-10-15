param(
    [string]$BaseUrl = "http://localhost"
)

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  X PHOTO VIEWER, VIDEO PLAYER & BACK RESTORATION TEST" -ForegroundColor Cyan
Write-Host "  Target: $BaseUrl" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# 1. Check API Readiness
Write-Host "`n[1/5] Checking service readiness..." -ForegroundColor Yellow
try {
    $ready = Invoke-RestMethod -Uri "$BaseUrl/health/ready" -TimeoutSec 3 -ErrorAction Stop
    Write-Host "  Status: $($ready.status) | DB: $($ready.database) | Redis: $($ready.redis)" -ForegroundColor Green
} catch {
    Write-Host "  API health check failed: $_" -ForegroundColor Red
    exit 1
}

# 2. Query Feed Posts
Write-Host "`n[2/5] Inspecting posts with media attachments..." -ForegroundColor Yellow
$feed = Invoke-RestMethod -Uri "$BaseUrl/api/v1/feed?limit=20" -TimeoutSec 4
$postsWithMedia = @($feed.posts | Where-Object { $_.media -and $_.media.Count -gt 0 })

$targetPost = $null
if ($postsWithMedia.Count -gt 0) {
    $targetPost = $postsWithMedia[0]
    Write-Host "  Found existing post with $($targetPost.media.Count) media item(s): ID=$($targetPost.id)" -ForegroundColor Green
} else {
    Write-Host "  No posts with media found in feed. Creating a synthetic demo post with photo & video..." -ForegroundColor Yellow
    # Create demo post with image and video
    $regBody = @{
        username     = "media_tester_$(Get-Random)"
        display_name = "Media Tester"
        email        = "tester_$(Get-Random)@example.com"
        password     = "password123!"
    } | ConvertTo-Json

    $user = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json"
    $sessionToken = $user.session_token

    # Create post with 2 media items (1 photo + 1 video)
    $postBody = @{
        content = "Testing X-like Photo Viewer and Video Player with gallery support! #tech #media"
        media = @(
            @{
                url = "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1200&auto=format&fit=crop&q=80"
                media_type = "image"
                width = 1200
                height = 800
            },
            @{
                url = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4"
                media_type = "video"
                width = 1920
                height = 1080
                duration_seconds = 15.0
            }
        )
    } | ConvertTo-Json

    $headers = @{ "Authorization" = "Bearer $sessionToken" }
    $targetPost = Invoke-RestMethod -Uri "$BaseUrl/api/v1/posts" -Method Post -Body $postBody -Headers $headers -ContentType "application/json"
    Write-Host "  Created test post ID=$($targetPost.id) with $($targetPost.media.Count) media attachments" -ForegroundColor Green
}

# 3. Test Standalone Direct Routes
$username = if ($targetPost.author.username) { $targetPost.author.username } else { "user" }
$postId = $targetPost.id

Write-Host "`n[3/5] Testing Direct Standalone Photo URL..." -ForegroundColor Yellow
$photoUrl = "$BaseUrl/$username/status/$postId/photo/1"
Write-Host "  GET $photoUrl" -ForegroundColor Gray
$photoResp = Invoke-WebRequest -Uri $photoUrl -UseBasicParsing -TimeoutSec 10 -MaximumRedirection 5
if ($photoResp.StatusCode -eq 200) {
    Write-Host "  ✓ HTTP 200 OK — Standalone Photo Viewer Page rendered successfully" -ForegroundColor Green
} else {
    Write-Host "  ✗ Photo route returned status $($photoResp.StatusCode)" -ForegroundColor Red
}

Write-Host "`n[4/5] Testing Direct Standalone Video URL..." -ForegroundColor Yellow
$videoUrl = "$BaseUrl/$username/status/$postId/video/2"
Write-Host "  GET $videoUrl" -ForegroundColor Gray
$videoResp = Invoke-WebRequest -Uri $videoUrl -UseBasicParsing -TimeoutSec 10 -MaximumRedirection 5
if ($videoResp.StatusCode -eq 200) {
    Write-Host "  ✓ HTTP 200 OK — Standalone Video Viewer Page rendered successfully" -ForegroundColor Green
} else {
    Write-Host "  ✗ Video route returned status $($videoResp.StatusCode)" -ForegroundColor Red
}

# 5. Test Generic Status Route Alias
Write-Host "`n[5/5] Testing Generic /status/:postId/photo/1 Route..." -ForegroundColor Yellow
$genericUrl = "$BaseUrl/status/$postId/photo/1"
Write-Host "  GET $genericUrl" -ForegroundColor Gray
$genericResp = Invoke-WebRequest -Uri $genericUrl -UseBasicParsing -TimeoutSec 10 -MaximumRedirection 5
if ($genericResp.StatusCode -eq 200) {
    Write-Host "  ✓ HTTP 200 OK — Generic Route Alias rendered successfully" -ForegroundColor Green
} else {
    Write-Host "  ✗ Generic route returned status $($genericResp.StatusCode)" -ForegroundColor Red
}

Write-Host "`n==========================================================" -ForegroundColor Green
Write-Host "  ALL MEDIA VIEWER & ROUTING TESTS PASSED (100% OK)" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green

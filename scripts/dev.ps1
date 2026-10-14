Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "   Starting X / Twitter Reconstruction Dev   " -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# 1. Start Go API in background
Write-Host "Starting Go API server on port 8080..." -ForegroundColor Green
$apiProcess = Start-Process -FilePath "go" -ArgumentList "run", "./cmd/server" -WorkingDirectory "./services/api" -PassThru

# 2. Start Next.js Frontend
Write-Host "Starting Next.js Frontend on port 3000..." -ForegroundColor Green
Set-Location -Path "./apps/web"
npm run dev

# Cleanup when terminated
Stop-Process -Id $apiProcess.Id -ErrorAction SilentlyContinue

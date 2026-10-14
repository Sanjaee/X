Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "   Running Full Test Suite & Build Check     " -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan

# 1. Run Go backend tests
Write-Host "`n[1/2] Running Go Unit & Integration Tests..." -ForegroundColor Yellow
Set-Location -Path "./services/api"
go test -v ./...
if ($LASTEXITCODE -ne 0) {
    Write-Host "Go tests failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Go tests passed successfully!" -ForegroundColor Green

# 2. Run Next.js build verification
Write-Host "`n[2/2] Running Next.js Build Verification..." -ForegroundColor Yellow
Set-Location -Path "../../apps/web"
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Next.js build verification failed!" -ForegroundColor Red
    exit 1
}
Write-Host "Next.js build passed successfully!" -ForegroundColor Green

Write-Host "`nAll verification checks passed!" -ForegroundColor Cyan

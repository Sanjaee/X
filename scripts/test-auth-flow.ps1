param(
    [string]$BaseUrl = "https://lostmedia.pro"
)

$rand = Get-Random
$username = "authtest_$rand"
$email = "authtest_$rand@example.com"
$password = "SecretPassword123!"

Write-Host "1. Testing Registration at $BaseUrl/api/v1/auth/register..." -ForegroundColor Cyan
$regBody = @{
    username     = $username
    display_name = "Auth Tester $rand"
    email        = $email
    password     = $password
} | ConvertTo-Json

$reg = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json"
Write-Host "  ✓ Registration OK: username=$($reg.user.username), ID=$($reg.user.id)" -ForegroundColor Green

Write-Host "`n2. Testing Login at $BaseUrl/api/v1/auth/login..." -ForegroundColor Cyan
$loginBody = @{
    identifier = $username
    password   = $password
} | ConvertTo-Json

$login = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
Write-Host "  ✓ Login OK: token=$($login.session_token.Substring(0, 16))..." -ForegroundColor Green

Write-Host "`n3. Testing Login alias at $BaseUrl/api/auth/login..." -ForegroundColor Cyan
$loginAlias = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
Write-Host "  ✓ Login Alias OK: username=$($loginAlias.user.username)" -ForegroundColor Green

Write-Host "`n4. Testing Authenticated /api/v1/me..." -ForegroundColor Cyan
$headers = @{ "Authorization" = "Bearer $($login.session_token)" }
$me = Invoke-RestMethod -Uri "$BaseUrl/api/v1/me" -Headers $headers
Write-Host "  ✓ /api/v1/me OK: username=$($me.username), email=$($me.email)" -ForegroundColor Green

Write-Host "`n5. Testing Authenticated /api/me (alias)..." -ForegroundColor Cyan
$meAlias = Invoke-RestMethod -Uri "$BaseUrl/api/me" -Headers $headers
Write-Host "  ✓ /api/me OK: username=$($meAlias.username), email=$($meAlias.email)" -ForegroundColor Green

Write-Host "==========================================================" -ForegroundColor Green
Write-Host "  ALL AUTH AND LOGIN TESTS PASSED" -ForegroundColor Green
Write-Host "==========================================================" -ForegroundColor Green

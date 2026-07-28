# dev.ps1 - Development Hot-Reload Launcher for Cheesy Arena
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "  🚀 Cheesy Arena Development Hot-Reload Server " -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan

$airCmd = Get-Command air -ErrorAction SilentlyContinue
if ($null -ne $airCmd) {
    air
} else {
    $gopath = go env GOPATH
    $airExe = Join-Path $gopath "bin\air.exe"
    if (Test-Path $airExe) {
        Write-Host "Running Air from $airExe..." -ForegroundColor Green
        & $airExe
    } else {
        Write-Host "⚠️ Air binary not found. Running standard compilation: go build -o cheesy-arena.exe ." -ForegroundColor Yellow
        go build -o cheesy-arena.exe .
    }
}

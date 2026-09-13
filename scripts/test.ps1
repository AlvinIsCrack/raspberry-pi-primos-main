[CmdletBinding()]
param (
    [Alias("v")]
    [switch]$VisualDelay,

    [string]$HttpPort = "8080"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot

# Lanzamiento del servidor en background con go run
Write-Host "[1/2] Iniciando servidor temporal..." -ForegroundColor Cyan

$srcDir = Join-Path $RepoRoot "src"
$env:HTTP_ADDR = ":$HttpPort"
$serverProcess = Start-Process -FilePath "go" -ArgumentList "run", "main.go" -WorkingDirectory $srcDir -PassThru -NoNewWindow

try {
    $maxRetries = 30
    $isReady = $false
    for ($i = 0; $i -lt $maxRetries; $i++) {
        Start-Sleep -Milliseconds 200
        try {
            $resp = Invoke-RestMethod -Uri "http://127.0.0.1:$HttpPort/api/system/status" -TimeoutSec 1 -ErrorAction Stop
            if ($null -ne $resp) {
                $isReady = $true
                break
            }
        }
        catch { }

        if ($serverProcess.HasExited) {
            throw "El servidor de pruebas se cerró inesperadamente con código: $($serverProcess.ExitCode)"
        }
    }

    if (-not $isReady) {
        throw "Tiempo de espera agotado: El servidor no respondió en http://127.0.0.1:$HttpPort"
    }

    Write-Host "Servidor listo y respondiendo." -ForegroundColor Green

    # Pruebas de integración con Pytest
    Write-Host "[2/2] Ejecutando suite de integración en Python..." -ForegroundColor Cyan
    $pytestArgs = @("-v", (Join-Path $RepoRoot "tests/integration"))
    if ($VisualDelay) {
        $pytestArgs += "--visual-delay"
    }

    $pythonExec = if (Get-Command "pytest" -ErrorAction SilentlyContinue) { "pytest" }
    elseif (Get-Command "python" -ErrorAction SilentlyContinue) { "python" }
    elseif (Get-Command "py" -ErrorAction SilentlyContinue) { "py" }
    else { throw "Neither pytest, python, nor py were found in PATH." }

    if ($pythonExec -eq "pytest") {
        & pytest @pytestArgs
    }
    else {
        & $pythonExec -m pytest @pytestArgs
    }

    if ($LASTEXITCODE -ne 0) {
        throw "La suite de pruebas de integración reportó fallos."
    }

    Write-Host "Todas las pruebas pasaron exitosamente." -ForegroundColor Green
}
finally {
    Write-Host "Deteniendo servidor de pruebas..." -ForegroundColor DarkGray
    if ($serverProcess -and -not $serverProcess.HasExited) {
        if ($IsWindows -or $env:OS -match "Windows") {
            taskkill /PID $serverProcess.Id /T /F 2>$null | Out-Null
        }
        else {
            pkill -P $serverProcess.Id 2>$null | Out-Null
            Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
            fuser -k "$($HttpPort)/tcp" 2>$null | Out-Null
        }
    }
}
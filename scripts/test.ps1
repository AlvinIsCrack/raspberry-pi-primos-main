[CmdletBinding()]
param (
    [Alias("v")]
    [switch]$VisualDelay,

    [Alias("g")]
    [switch]$SkipGoTests,

    [Alias("t")]
    [switch]$SkipPyTests,

    [string]$HttpPort = "8080"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot

# [1/2] Pruebas unitarias de Go
if (-not $SkipGoTests) {
    Write-Host "Ejecutando pruebas de Go..." -ForegroundColor Cyan
    Push-Location (Join-Path $RepoRoot "src")
    try {
        $goArch = (go env GOARCH).Trim()
        $flags = @("-v")
        if ($goArch -in @("amd64", "arm64")) { $flags += "-race" }
        go test @flags ./...
        if ($LASTEXITCODE -ne 0) {
            throw "Las pruebas de Go fallaron."
        }
    }
    finally {
        Pop-Location
    }
}
else {
    Write-Host "Omitiendo pruebas de Go (-SkipGoTests activo)." -ForegroundColor Yellow
}

# [2/2] Pruebas de la suite /tests
if (-not $SkipPyTests) {
    Write-Host "Iniciando servidor temporal..." -ForegroundColor Cyan

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

        Write-Host "Ejecutando suite de pruebas (/tests)..." -ForegroundColor Cyan
        $pytestArgs = @("-v", (Join-Path $RepoRoot "tests"))
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
            throw "La suite de pruebas reportó fallos."
        }

        Write-Host "Todas las pruebas de /tests pasaron exitosamente." -ForegroundColor Green
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
}
else {
    Write-Host "Omitiendo suite de /tests (-SkipPyTests activo)." -ForegroundColor Yellow
}
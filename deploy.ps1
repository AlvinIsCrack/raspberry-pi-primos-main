[CmdletBinding()]
param (
    [Alias("b")]
    [switch]$Build
)

$ErrorActionPreference = "Stop"

$IP = "192.168.1.11"
$USER = "dietpi"

if ($Build) {
    Write-Host "Construyendo assets web con pnpm..." -ForegroundColor Cyan
    Push-Location src/web
    try {
        pnpm run build
        if ($LASTEXITCODE -ne 0) {
            throw "El comando 'pnpm run build' finalizo con errores."
        }
    }
    finally {
        Pop-Location
    }
}

Write-Host "Compilando backend Go para ARMv6..." -ForegroundColor Cyan

$prevGOOS = $env:GOOS
$prevGOARCH = $env:GOARCH
$prevGOARM = $env:GOARM

Push-Location src
try {
    $env:GOOS = "linux"
    $env:GOARCH = "arm"
    $env:GOARM = "6"

    # Asegura que exista el directorio de salida si no está creado
    if (-not (Test-Path "../build")) {
        New-Item -ItemType Directory -Path "../build" | Out-Null
    }

    go build -ldflags="-s -w" -o ../build/dashboard .
    $buildSuccess = ($LASTEXITCODE -eq 0)
}
finally {
    $env:GOOS = $prevGOOS
    $env:GOARCH = $prevGOARCH
    $env:GOARM = $prevGOARM
    Pop-Location
}

if ($buildSuccess) {
    Write-Host "Transfiriendo binario y reiniciando servicios..." -ForegroundColor Cyan

    $remoteScript = "set -e; " +
                    "sudo systemctl stop dashboard.service kiosk.service 2>/dev/null || true; " +
                    "sudo fuser -k 8080/tcp 1883/tcp >/dev/null 2>&1 || true; " +
                    "cat > /home/$USER/dashboard.new; " +
                    "chmod +x /home/$USER/dashboard.new; " +
                    "mv -f /home/$USER/dashboard.new /home/$USER/dashboard; " +
                    "sudo systemctl start dashboard.service kiosk.service"

    $binaryPath = (Resolve-Path "build/dashboard").Path

    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.RedirectStandardError = $true
    $startInfo.FileName = "ssh"
    $startInfo.Arguments = "$($USER)@$($IP) `"$remoteScript`""
    $startInfo.UseShellExecute = $false
    $startInfo.RedirectStandardInput = $true

    $process = [System.Diagnostics.Process]::Start($startInfo)
    $fileStream = [System.IO.File]::OpenRead($binaryPath)

    try {
        $fileStream.CopyTo($process.StandardInput.BaseStream)
    }
    finally {
        $fileStream.Dispose()
        $process.StandardInput.BaseStream.Dispose()
    }

    $stderr = $process.StandardError.ReadToEnd()
    $process.WaitForExit()

    if ($process.ExitCode -eq 0) {
        Write-Host "Despliegue completado exitosamente." -ForegroundColor Green
    }
    else {
        Write-Host "Fallo durante la transferencia o reinicio remoto: $stderr" -ForegroundColor Red
        exit 1
    }
}
else {
    Write-Host "Error en la compilacion de Go." -ForegroundColor Red
    exit 1
}
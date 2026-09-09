[CmdletBinding()]
param (
    [Alias("b")]
    [switch]$Build,

    [Alias("s")]
    [switch]$Setup
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

if ($Setup) {
    Write-Host "Configurando unidades systemd remotas..." -ForegroundColor Cyan
    scp template.dashboard.service template.kiosk.service "$($USER)@$($IP):/tmp/"
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo al copiar los archivos de servicio mediante SCP."
    }

    $setupScript = "set -e; " +
    "trap 'rm -f /tmp/template.dashboard.service /tmp/template.kiosk.service' EXIT; " +
    "sudo install -m 644 -o root -g root /tmp/template.dashboard.service /etc/systemd/system/dashboard.service; " +
    "sudo install -m 644 -o root -g root /tmp/template.kiosk.service /etc/systemd/system/kiosk.service; " +
    "sudo systemctl daemon-reload; " +
    "sudo systemctl enable --now dashboard.service kiosk.service"
    
    ssh "$($USER)@$($IP)" $setupScript
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo en la configuración remota de systemd."
    }
    Write-Host "Configuración de servicios completada." -ForegroundColor Green
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
    Write-Host "Transfiriendo binario..." -ForegroundColor Cyan
    $binaryPath = (Resolve-Path "build/dashboard").Path

    scp $binaryPath "$($USER)@$($IP):/home/$($USER)/dashboard.new"
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Fallo al transferir el binario mediante scp." -ForegroundColor Red
        exit 1
    }

    $remoteScript = "set -e; " +
    "sudo systemctl stop dashboard.service kiosk.service 2>/dev/null || true; " +
    "sudo fuser -k 8080/tcp 1883/tcp >/dev/null 2>&1 || true; " +
    "chmod +x /home/$USER/dashboard.new; " +
    "mv -f /home/$USER/dashboard.new /home/$USER/dashboard; " +
    "sudo systemctl start dashboard.service kiosk.service"

    Write-Host "Reiniciando servicios remotos..." -ForegroundColor Cyan
    ssh "$($USER)@$($IP)" $remoteScript

    if ($LASTEXITCODE -eq 0) {
        Write-Host "Despliegue completado exitosamente." -ForegroundColor Green
    }
    else {
        Write-Host "Fallo durante la configuración o reinicio remoto." -ForegroundColor Red
        exit 1
    }
}
else {
    Write-Host "Error en la compilacion de Go." -ForegroundColor Red
    exit 1
}
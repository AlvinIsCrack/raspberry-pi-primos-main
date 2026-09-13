[CmdletBinding()]
param (
    [Alias("b")]
    [switch]$Build,

    [Alias("s")]
    [switch]$Setup,

    [string]$TargetIP = "192.168.1.11",
    [string]$TargetUser = "dietpi",
    [string]$HttpPort = "8080",
    [string]$UdpPort = "1884"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent $PSScriptRoot

if ($Build) {
    Write-Host "Construyendo assets web con pnpm..." -ForegroundColor Cyan
    Push-Location (Join-Path $RepoRoot "frontend")
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
    foreach ($file in @($tmplDash, $tmplKiosk)) {
        if (-not (Test-Path $file)) {
            throw "Deployment template missing: $file"
        }
    }

    Write-Host "Configurando unidades systemd remotas..." -ForegroundColor Cyan
    $tmplDash = Join-Path $RepoRoot "template.dashboard.service"
    $tmplKiosk = Join-Path $RepoRoot "template.kiosk.service"
    scp $tmplDash $tmplKiosk "$($TargetUser)@$($TargetIP):/tmp/"
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo al copiar los archivos de servicio mediante SCP."
    }

    $setupScript = "set -e; " +
    "trap 'rm -f /tmp/template.dashboard.service /tmp/template.kiosk.service' EXIT; " +
    "sudo install -m 644 -o root -g root /tmp/template.dashboard.service /etc/systemd/system/dashboard.service; " +
    "sudo install -m 644 -o root -g root /tmp/template.kiosk.service /etc/systemd/system/kiosk.service; " +
    "sudo systemctl daemon-reload; " +
    "sudo systemctl enable --now dashboard.service kiosk.service"
    
    ssh "$($TargetUser)@$($TargetIP)" $setupScript
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo en la configuración remota de systemd."
    }
    Write-Host "Configuración de servicios completada." -ForegroundColor Green
}

Write-Host "Compilando backend Go para ARMv6..." -ForegroundColor Cyan

$prevGOOS = $env:GOOS
$prevGOARCH = $env:GOARCH
$prevGOARM = $env:GOARM

Push-Location (Join-Path $RepoRoot "src")
try {
    $env:GOOS = "linux"
    $env:GOARCH = "arm"
    $env:GOARM = "6"

    $outDir = Join-Path $RepoRoot "build"
    if (-not (Test-Path $outDir)) {
        New-Item -ItemType Directory -Path $outDir | Out-Null
    }
    go build -ldflags="-s -w" -o (Join-Path $outDir "dashboard") .
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
    $binaryPath = Join-Path $RepoRoot "build/dashboard"
    scp $binaryPath "$($TargetUser)@$($TargetIP):/home/$($TargetUser)/dashboard.new"
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Fallo al transferir el binario mediante scp." -ForegroundColor Red
        exit 1
    }

    $remoteScript = "set -e; " +
    "sudo systemctl stop dashboard.service kiosk.service 2>/dev/null || true; " +
    "sudo fuser -k $($HttpPort)/tcp $($UdpPort)/udp >/dev/null 2>&1 || true; " +
    "chmod +x `"/home/$TargetUser/dashboard.new`"; " +
    "mv -f `"/home/$TargetUser/dashboard.new`" `"/home/$TargetUser/dashboard`"; " +
    "sudo systemctl start dashboard.service kiosk.service"

    Write-Host "Reiniciando servicios remotos..." -ForegroundColor Cyan
    ssh "$($TargetUser)@$($TargetIP)" $remoteScript

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
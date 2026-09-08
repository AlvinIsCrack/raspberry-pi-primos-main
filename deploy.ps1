$IP = "192.168.1.11"
$USER = "dietpi"

$env:GOOS = "linux"
$env:GOARCH = "arm"
$env:GOARM = "6"

Write-Host "Compilando para ARMv6..." -ForegroundColor Cyan
Push-Location src
go build -ldflags="-s -w" -o ../build/dashboard .
Pop-Location

if ($LASTEXITCODE -eq 0) {
    Write-Host "Subiendo binario a la Raspberry Pi..." -ForegroundColor Cyan
    scp build/dashboard "$($USER)@$($IP):/home/$USER/dashboard.new"

    Write-Host "Deteniendo servicio, reemplazando ejecutable y reiniciando..." -ForegroundColor Cyan
    $remoteCmd = "sudo systemctl stop dashboard.service kiosk.service 2>/dev/null; " +
    "sudo fuser -k 8080/tcp 1883/tcp >/dev/null 2>&1 || true; " +
    "chmod +x /home/$USER/dashboard.new && " +
    "mv /home/$USER/dashboard.new /home/$USER/dashboard && " +
    "sudo systemctl start dashboard.service kiosk.service"

    ssh "$($USER)@$($IP)" $remoteCmd

    Write-Host "Despliegue completado exitosamente." -ForegroundColor Green
}
else {
    Write-Host "Error en la compilacion de Go." -ForegroundColor Red
}
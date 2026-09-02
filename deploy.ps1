$IP = "192.168.1.11"
$USER = "dietpi"

$env:GOOS = "linux"
$env:GOARCH = "arm"
$env:GOARM = "6"

Write-Host "Compilando para ARMv6..." -ForegroundColor Cyan
go build -ldflags="-s -w" -o build/dashboard src/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Subiendo binario a la Raspberry Pi..." -ForegroundColor Cyan
    scp build/dashboard "$($USER)@$($IP):/home/$USER/dashboard.new"
    
    Write-Host "Asignando permisos, reemplazando y reiniciando servicio..." -ForegroundColor Cyan
    ssh "$($USER)@$($IP)" "chmod +x /home/$USER/dashboard.new && mv /home/$USER/dashboard.new /home/$USER/dashboard && sudo systemctl restart dashboard"
    
    Write-Host "Deploy completado exitosamente." -ForegroundColor Green
}
else {
    Write-Host "Error en la compilación de Go." -ForegroundColor Red
}
# run_services.ps1
# Script para iniciar los servicios localmente en Windows

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host "Iniciando Servicios de Meloop" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

# Iniciar test-service en segundo plano (puerto 8081)
Write-Host "Iniciando test-service en puerto 8081..." -ForegroundColor Yellow
$testProc = Start-Process go -ArgumentList "run ./services/test-service" -NoNewWindow -PassThru

# Iniciar user-service en segundo plano (puerto 8082)
Write-Host "Iniciando user-service en puerto 8082..." -ForegroundColor Yellow
$userProc = Start-Process go -ArgumentList "run ./services/user-service" -NoNewWindow -PassThru

# Iniciar auth-service en segundo plano (puerto 8083)
Write-Host "Iniciando auth-service en puerto 8083..." -ForegroundColor Yellow
$authProc = Start-Process go -ArgumentList "run ./services/auth-service" -NoNewWindow -PassThru

# Esperar 2 segundos para dar tiempo a los servicios de iniciar
Start-Sleep -Seconds 2

# Iniciar api-gateway en primer plano (puerto 8080)
Write-Host "Iniciando api-gateway en puerto 8080..." -ForegroundColor Yellow
Write-Host "Presiona Ctrl+C para detener todos los servicios." -ForegroundColor Gray
Write-Host "----------------------------------------------" -ForegroundColor Gray

try {
    go run ./services/api-gateway
} finally {
    # Al detener el gateway, apagamos los microservicios en segundo plano
    Write-Host "`nDeteniendo microservicios de dominio y pruebas..." -ForegroundColor Gray
    Stop-Process -Id $testProc.Id -ErrorAction SilentlyContinue
    Stop-Process -Id $userProc.Id -ErrorAction SilentlyContinue
    Stop-Process -Id $authProc.Id -ErrorAction SilentlyContinue
    Write-Host "Servicios detenidos con éxito." -ForegroundColor Green
}

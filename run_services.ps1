# run_services.ps1
# Script para iniciar los microservicios localmente en Windows

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host "Iniciando Servicios de Meloop" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

# 1. Cargar variables de entorno desde .env si existe
if (Test-Path ".env") {
    Write-Host "Cargando variables de entorno desde .env..." -ForegroundColor Gray
    Get-Content ".env" | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
            $parts = $line.Split("=", 2)
            $name = $parts[0].Trim()
            $val = $parts[1].Trim()
            [System.Environment]::SetEnvironmentVariable($name, $val, "Process")
        }
    }
}

# 2. Función para liberar puertos en caso de ejecuciones anteriores
function Free-Port {
    param ([int]$Port)
    try {
        $conns = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
        foreach ($c in $conns) {
            if ($c.OwningProcess -gt 0) {
                Write-Host "Liberando puerto $Port (PID $($c.OwningProcess))..." -ForegroundColor DarkYellow
                Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
            }
        }
    } catch {}
}

@(8080, 8081, 8082, 8083, 8086) | ForEach-Object { Free-Port -Port $_ }

# 3. Iniciar microservicios en segundo plano
Write-Host "Iniciando test-service en puerto 8081..." -ForegroundColor Yellow
$testProc = Start-Process go -ArgumentList "run ./services/test-service" -NoNewWindow -PassThru

Write-Host "Iniciando user-service en puerto 8082..." -ForegroundColor Yellow
$userProc = Start-Process go -ArgumentList "run ./services/user-service" -NoNewWindow -PassThru

Write-Host "Iniciando auth-service en puerto 8083..." -ForegroundColor Yellow
$authProc = Start-Process go -ArgumentList "run ./services/auth-service" -NoNewWindow -PassThru

Write-Host "Iniciando social-service en puerto 8086..." -ForegroundColor Yellow
$socialProc = Start-Process go -ArgumentList "run ./services/social-service" -NoNewWindow -PassThru

# Esperar 2.5 segundos para inicialización
Start-Sleep -Milliseconds 2500

# 4. Iniciar api-gateway en primer plano (puerto 8080)
Write-Host "Iniciando api-gateway en puerto 8080..." -ForegroundColor Yellow
Write-Host "Presiona Ctrl+C para detener todos los servicios." -ForegroundColor Gray
Write-Host "----------------------------------------------" -ForegroundColor Gray

try {
    go run ./services/api-gateway
} finally {
    Write-Host "`nDeteniendo microservicios de dominio y pruebas..." -ForegroundColor Gray
    Stop-Process -Id $testProc.Id -ErrorAction SilentlyContinue
    Stop-Process -Id $userProc.Id -ErrorAction SilentlyContinue
    Stop-Process -Id $authProc.Id -ErrorAction SilentlyContinue
    Stop-Process -Id $socialProc.Id -ErrorAction SilentlyContinue
    Write-Host "Servicios detenidos con éxito." -ForegroundColor Green
}

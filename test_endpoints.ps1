# test_endpoints.ps1
# Script para probar los endpoints y verificar la estructura unificada de respuestas y errores

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host "Iniciando Pruebas de Endpoints Unificados" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan

# Función auxiliar para enviar peticiones HTTP compatible con PowerShell 5.1
function Send-Request {
    param (
        [string]$Uri,
        [string]$Method = "GET",
        [string]$Body = $null
    )
    $result = @{
        StatusCode = 0
        Content = ""
        Error = $null
    }
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            UseBasicParsing = $true
            ErrorAction = "Stop"
        }
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }
        $response = Invoke-WebRequest @params
        $result.StatusCode = $response.StatusCode
        $result.Content = $response.Content
    } catch {
        $webResp = $_.Exception.Response
        if ($webResp) {
            $result.StatusCode = [int]$webResp.StatusCode
            $stream = $webResp.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $result.Content = $reader.ReadToEnd()
            $reader.Close()
        } else {
            $result.Error = $_.Exception.Message
        }
    }
    return $result
}

$endpoints = @(
    @{ Name = "1. Health Check - Gateway (Success)"; URL = "http://localhost:8080/health"; ExpectedStatus = 200; Method = "GET" },
    @{ Name = "2. Health Check - Test Service (Success)"; URL = "http://localhost:8080/test/health"; ExpectedStatus = 200; Method = "GET" },
    @{ Name = "3. Listar Usuarios - User Service (Success)"; URL = "http://localhost:8080/users"; ExpectedStatus = 200; Method = "GET" },
    @{ Name = "4. Crear Usuario - Validación Vacía (400 - VALIDATION_ERROR)"; URL = "http://localhost:8080/users"; ExpectedStatus = 400; Method = "POST"; Body = '{}' },
    @{ Name = "5. Crear Usuario Exitoso (201 - Success)"; URL = "http://localhost:8080/users"; ExpectedStatus = 201; Method = "POST"; Body = '{"username": "nuevo_usuario", "email": "nuevo@meloop.com"}' },
    @{ Name = "6. Auth Login - Validación Vacía (400 - VALIDATION_ERROR)"; URL = "http://localhost:8080/auth/login"; ExpectedStatus = 400; Method = "POST"; Body = '{}' },
    @{ Name = "7. Auth Login - Credenciales Incorrectas (401 - UNAUTHORIZED)"; URL = "http://localhost:8080/auth/login"; ExpectedStatus = 401; Method = "POST"; Body = '{"email": "alan@meloop.com", "password": "incorrect_password"}' },
    @{ Name = "8. Auth Login - Éxito (200 - Success)"; URL = "http://localhost:8080/auth/login"; ExpectedStatus = 200; Method = "POST"; Body = '{"email": "alan@meloop.com", "password": "meloop123"}' },
    @{ Name = "9. Pánico Recuperado - Test Service (500 - INTERNAL_ERROR via Recovery)"; URL = "http://localhost:8080/test/panic"; ExpectedStatus = 500; Method = "GET" }
)

foreach ($ep in $endpoints) {
    Write-Host "`n--------------------------------------------" -ForegroundColor Gray
    Write-Host "Prueba: $($ep.Name)" -ForegroundColor White
    Write-Host "URL: $($ep.URL)" -ForegroundColor Gray
    
    $res = Send-Request -Uri $ep.URL -Method $ep.Method -Body $ep.Body
    
    if ($res.Error) {
        Write-Host "Error al realizar la petición: $($res.Error)" -ForegroundColor Red
        continue
    }
    
    $statusCode = $res.StatusCode
    $json = $res.Content | ConvertFrom-Json | ConvertTo-Json -Depth 5
    
    if ($statusCode -eq $ep.ExpectedStatus) {
        Write-Host "Status: $statusCode (Correcto)" -ForegroundColor Green
    } else {
        Write-Host "Status: $statusCode (Esperado: $($ep.ExpectedStatus))" -ForegroundColor Red
    }
    
    Write-Host "Respuesta:"
    Write-Host $json -ForegroundColor Cyan
}

Write-Host "`n==============================================" -ForegroundColor Cyan
Write-Host "Prueba de Caída de Servicio (502 - BAD_GATEWAY)" -ForegroundColor Cyan
Write-Host "Instrucciones: Detén el 'user-service' (o presiona Ctrl+C en la otra ventana) y pulsa enter para continuar..." -ForegroundColor Yellow
Read-Host

Write-Host "Haciendo petición a un servicio fuera de línea..." -ForegroundColor Gray
$res = Send-Request -Uri "http://localhost:8080/users" -Method "GET"

if ($res.Error) {
    Write-Host "Error al realizar la petición (Gateway o servicio inaccesible): $($res.Error)" -ForegroundColor Red
} else {
    $json = $res.Content | ConvertFrom-Json | ConvertTo-Json -Depth 5
    $color = if ($res.StatusCode -eq 502) { "Green" } else { "Red" }
    Write-Host "Status: $($res.StatusCode) (Esperado: 502)" -ForegroundColor $color
    Write-Host "Respuesta:"
    Write-Host $json -ForegroundColor Cyan
}

Write-Host "`nPruebas completadas." -ForegroundColor Green

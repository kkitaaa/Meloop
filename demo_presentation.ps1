# demo_presentation.ps1
# Script de ejecucion tecnica de endpoints - Meloop Architecture

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "`n============================================================" -ForegroundColor Cyan
Write-Host "          EJECUCION DE MICROSERVICIOS Y ENDPOINTS           " -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "Gateway URL: http://localhost:8080`n" -ForegroundColor Gray

# Funcion para enviar y desplegar peticiones HTTP detalladas
function Execute-Endpoint {
    param (
        [string]$Title,
        [string]$Method = "GET",
        [string]$Url,
        [hashtable]$Headers = @{},
        [string]$BodyJson = $null
    )

    Write-Host "------------------------------------------------------------" -ForegroundColor Yellow
    Write-Host "$Title" -ForegroundColor Yellow
    Write-Host "------------------------------------------------------------" -ForegroundColor Yellow
    
    Write-Host "[PETICION HTTP]" -ForegroundColor White
    Write-Host "> $Method $Url" -ForegroundColor Cyan
    
    if ($Headers.Count -gt 0) {
        foreach ($k in $Headers.Keys) {
            Write-Host "> Header: $k = $($Headers[$k])" -ForegroundColor Gray
        }
    }
    
    if ($BodyJson) {
        Write-Host "> Payload (JSON):" -ForegroundColor Gray
        Write-Host ($BodyJson | ConvertFrom-Json | ConvertTo-Json -Depth 5) -ForegroundColor DarkCyan
    }

    $result = @{
        StatusCode = 0
        Content = ""
        Json = $null
        Error = $null
    }

    try {
        $params = @{
            Uri = $Url
            Method = $Method
            UseBasicParsing = $true
            ErrorAction = "Stop"
        }
        if ($Headers.Count -gt 0) {
            $params.Headers = $Headers
        }
        if ($BodyJson) {
            $params.Body = $BodyJson
            $params.ContentType = "application/json; charset=utf-8"
        }
        
        $sw = [System.Diagnostics.Stopwatch]::StartNew()
        $response = Invoke-WebRequest @params
        $sw.Stop()

        $result.StatusCode = [int]$response.StatusCode
        $result.Content = $response.Content
        if ($response.Content) {
            try { $result.Json = $response.Content | ConvertFrom-Json } catch {}
        }
        $timeMs = $sw.ElapsedMilliseconds
    } catch {
        $webResp = $_.Exception.Response
        if ($webResp) {
            $result.StatusCode = [int]$webResp.StatusCode
            $stream = $webResp.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $result.Content = $reader.ReadToEnd()
            $reader.Close()
            if ($result.Content) {
                try { $result.Json = $result.Content | ConvertFrom-Json } catch {}
            }
        } else {
            $result.Error = $_.Exception.Message
        }
        $timeMs = 0
    }

    Write-Host "`n[RESPUESTA HTTP]" -ForegroundColor White
    $statusColor = if ($result.StatusCode -ge 200 -and $result.StatusCode -lt 300) { "Green" } else { "Red" }
    Write-Host "< Codigo de Estado: $($result.StatusCode) ($timeMs ms)" -ForegroundColor $statusColor
    
    if ($result.Content) {
        Write-Host "< Cuerpo:" -ForegroundColor Gray
        try {
            Write-Host ($result.Content | ConvertFrom-Json | ConvertTo-Json -Depth 6) -ForegroundColor $(if ($statusColor -eq "Green") { "Cyan" } else { "DarkYellow" })
        } catch {
            Write-Host $result.Content -ForegroundColor Gray
        }
    }
    
    if ($result.Error) {
        Write-Host "< Error de red / transporte: $($result.Error)" -ForegroundColor Red
    }

    Write-Host "`nPresione [Enter] para continuar..." -ForegroundColor DarkGray
    Read-Host | Out-Null
    return $result
}

# -------------------------------------------------------------
# 1. Health Checks
# -------------------------------------------------------------
$h1 = Execute-Endpoint -Title "1. Health Check - API Gateway" -Method "GET" -Url "http://localhost:8080/health"
$h2 = Execute-Endpoint -Title "2. Health Check - Microservicio de Pruebas" -Method "GET" -Url "http://localhost:8080/test/health"

# -------------------------------------------------------------
# 2. Registro de Usuarios
# -------------------------------------------------------------
$suffix = (Get-Date).ToString("mmss")
$user1_name = "alan_$suffix"
$user1_email = "alan_$suffix@meloop.com"
$user2_name = "carlos_$suffix"
$user2_email = "carlos_$suffix@meloop.com"

$bodyUser1 = @{
    username = $user1_name
    email = $user1_email
    password = "Password123!"
} | ConvertTo-Json

$reg1 = Execute-Endpoint -Title "3. Registrar Usuario Principal (Alan)" -Method "POST" -Url "http://localhost:8080/auth/register" -BodyJson $bodyUser1

$user1_id = if ($reg1.Json -and $reg1.Json.data -and $reg1.Json.data.id) { $reg1.Json.data.id } else { "u1_$suffix" }

$bodyUser2 = @{
    username = $user2_name
    email = $user2_email
    password = "Password123!"
} | ConvertTo-Json

$reg2 = Execute-Endpoint -Title "4. Registrar Segundo Usuario (Carlos)" -Method "POST" -Url "http://localhost:8080/auth/register" -BodyJson $bodyUser2

$user2_id = if ($reg2.Json -and $reg2.Json.data -and $reg2.Json.data.id) { $reg2.Json.data.id } else { "u2_$suffix" }

# -------------------------------------------------------------
# 3. Configuracion de Privacidad y Perfil (user-service)
# -------------------------------------------------------------
$headersUser1 = @{ "X-User-ID" = $user1_id }

$privBody = @{
    visibilidad_perfil = "PRIVADO"
    recepcion_mensajes = "AMIGOS"
} | ConvertTo-Json

$priv = Execute-Endpoint -Title "5. Actualizar Configuracion de Privacidad" -Method "PATCH" -Url "http://localhost:8080/users/me/privacy" -Headers $headersUser1 -BodyJson $privBody

$profBody = @{
    biografia = "Desarrollador y musico entusiasta"
    tema = "neon_dark"
} | ConvertTo-Json

$prof = Execute-Endpoint -Title "6. Actualizar Datos de Perfil" -Method "PUT" -Url "http://localhost:8080/users/me/profile" -Headers $headersUser1 -BodyJson $profBody

$getAcc = Execute-Endpoint -Title "7. Consultar Cuenta del Usuario Autenticado" -Method "GET" -Url "http://localhost:8080/users/me" -Headers $headersUser1

# -------------------------------------------------------------
# 4. Solicitud y Aceptacion de Amistad (social-service)
# -------------------------------------------------------------
$friendReqBody = @{
    id_usuario = $user2_id
} | ConvertTo-Json

$sendReq = Execute-Endpoint -Title "8. Enviar Solicitud de Amistad (Alan -> Carlos)" -Method "POST" -Url "http://localhost:8080/friends/requests" -Headers $headersUser1 -BodyJson $friendReqBody

$requestId = if ($sendReq.Json -and $sendReq.Json.data -and $sendReq.Json.data.id_amistad) { $sendReq.Json.data.id_amistad } else { 1 }

$headersUser2 = @{ "X-User-ID" = $user2_id }

$recvReq = Execute-Endpoint -Title "9. Listar Solicitudes de Amistad Recibidas (Carlos)" -Method "GET" -Url "http://localhost:8080/friends/requests/received" -Headers $headersUser2

$acceptReq = Execute-Endpoint -Title "10. Aceptar Solicitud de Amistad (ID: $requestId)" -Method "POST" -Url "http://localhost:8080/friends/requests/$requestId/accept" -Headers $headersUser2

$listFriends = Execute-Endpoint -Title "11. Listar Amigos Confirmados (Alan)" -Method "GET" -Url "http://localhost:8080/friends" -Headers $headersUser1

# -------------------------------------------------------------
# 5. Bloqueo de Usuario (social-service)
# -------------------------------------------------------------
$bodyUser3 = @{
    username = "spam_bot_$suffix"
    email = "spam_$suffix@meloop.com"
    password = "Password123!"
} | ConvertTo-Json

# Registrar un usuario para bloquearlo validamente en base de datos
$user3_id = $null
try {
    $res3 = Invoke-RestMethod -Uri "http://localhost:8080/auth/register" -Method POST -Body $bodyUser3 -ContentType "application/json; charset=utf-8" -ErrorAction Stop
    if ($res3 -and $res3.data -and $res3.data.id) {
        $user3_id = $res3.data.id
    }
} catch {}

if (-not $user3_id) {
    $user3_id = $user2_id
}

$blockBody = @{
    id_usuario = $user3_id
} | ConvertTo-Json

$blockResp = Execute-Endpoint -Title "12. Bloquear Usuario por Seguridad" -Method "POST" -Url "http://localhost:8080/friends/blocks" -Headers $headersUser1 -BodyJson $blockBody

Write-Host "============================================================" -ForegroundColor Green
Write-Host "          FLUJO COMPLETO FINALIZADO CON EXITO               " -ForegroundColor Green
Write-Host "============================================================`n" -ForegroundColor Green

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"
$BaseUrl = $BaseUrl.TrimEnd("/")

$script:Passed = 0
$script:Failed = 0
$script:Skipped = 0

function Get-JsonPathValue {
    param(
        [object]$Object,
        [string]$Path
    )

    $current = $Object
    foreach ($part in $Path.Split(".")) {
        if ($null -eq $current) {
            return $null
        }
        $property = $current.PSObject.Properties[$part]
        if ($null -eq $property) {
            return $null
        }
        $current = $property.Value
    }
    return $current
}

function Send-ApiRequest {
    param(
        [string]$Path,
        [string]$Method = "GET",
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )

    $result = @{
        StatusCode = 0
        Content = ""
        Json = $null
        InvalidJson = $false
    }
    $request = @{
        Uri = "$BaseUrl$Path"
        Method = $Method
        UseBasicParsing = $true
        ErrorAction = "Stop"
        TimeoutSec = 15
    }

    if ($null -ne $Body) {
        $request.Body = ConvertTo-Json -InputObject $Body -Depth 10 -Compress
        $request.ContentType = "application/json"
    }
    if ($Headers.Count -gt 0) {
        $request.Headers = $Headers
    }

    try {
        $response = Invoke-WebRequest @request
        $result.StatusCode = [int]$response.StatusCode
        $result.Content = [string]$response.Content
    }
    catch {
        $webResponse = $_.Exception.Response
        if ($null -eq $webResponse) {
            return $result
        }

        $result.StatusCode = [int]$webResponse.StatusCode
        if ($null -ne $_.ErrorDetails -and -not [string]::IsNullOrWhiteSpace($_.ErrorDetails.Message)) {
            $result.Content = $_.ErrorDetails.Message
        }
        else {
            $stream = $webResponse.GetResponseStream()
            if ($null -ne $stream) {
                $reader = New-Object System.IO.StreamReader($stream)
                try {
                    $result.Content = $reader.ReadToEnd()
                }
                finally {
                    $reader.Dispose()
                }
            }
        }
    }

    if (-not [string]::IsNullOrWhiteSpace($result.Content)) {
        try {
            $result.Json = ConvertFrom-Json -InputObject $result.Content -ErrorAction Stop
        }
        catch {
            $result.InvalidJson = $true
        }
    }

    return $result
}

function Record-ApiCase {
    param(
        [string]$Name,
        [hashtable]$Response,
        [int]$ExpectedStatus,
        [object]$ExpectedSuccess = $null,
        [string]$ExpectedErrorCode = $null,
        [string]$ExpectedErrorIssue = $null,
        [hashtable]$ExpectedData = @{},
        [string[]]$RequiredData = @()
    )

    $failedChecks = @()
    if ($Response.StatusCode -ne $ExpectedStatus) {
        $failedChecks += "HTTP status"
    }
    if ($Response.InvalidJson) {
        $failedChecks += "JSON format"
    }

    if ($null -ne $ExpectedSuccess) {
        $actualSuccess = Get-JsonPathValue -Object $Response.Json -Path "success"
        if ($null -eq $actualSuccess -or [bool]$actualSuccess -ne [bool]$ExpectedSuccess) {
            $failedChecks += "success field"
        }
    }

    if (-not [string]::IsNullOrEmpty($ExpectedErrorCode)) {
        if ((Get-JsonPathValue -Object $Response.Json -Path "error.code") -ne $ExpectedErrorCode) {
            $failedChecks += "error code"
        }
    }

    if (-not [string]::IsNullOrEmpty($ExpectedErrorIssue)) {
        if ((Get-JsonPathValue -Object $Response.Json -Path "error.details.issue") -ne $ExpectedErrorIssue) {
            $failedChecks += "error details"
        }
    }

    foreach ($key in $ExpectedData.Keys) {
        if ((Get-JsonPathValue -Object $Response.Json -Path ("data." + $key)) -ne $ExpectedData[$key]) {
            $failedChecks += "response data"
        }
    }

    foreach ($key in $RequiredData) {
        $value = Get-JsonPathValue -Object $Response.Json -Path ("data." + $key)
        if ($null -eq $value -or [string]::IsNullOrWhiteSpace([string]$value)) {
            $failedChecks += "required response data"
        }
    }

    $passed = $failedChecks.Count -eq 0
    if ($passed) {
        $script:Passed++
        Write-Host "PASS  $Name (HTTP $ExpectedStatus)" -ForegroundColor Green
    }
    else {
        $script:Failed++
        if ($Response.StatusCode -eq 0) {
            Write-Host "FAIL  $Name (sin respuesta del gateway; comprueba el port-forward)" -ForegroundColor Red
        }
        elseif ($Response.InvalidJson) {
            Write-Host "FAIL  $Name (la respuesta no es JSON valido; HTTP $($Response.StatusCode))" -ForegroundColor Red
        }
        else {
            $checks = $failedChecks -join ", "
            Write-Host "FAIL  $Name (respuesta distinta de la esperada; HTTP $($Response.StatusCode); revisar $checks)" -ForegroundColor Red
        }
    }

    return $passed
}

function Record-Skipped {
    param(
        [string]$Name,
        [string]$Reason
    )

    $script:Skipped++
    Write-Host "SKIP  $Name - $Reason" -ForegroundColor Yellow
}

$guid = [Guid]::NewGuid().ToString("N").Substring(0, 12)
$username = "auth-test-$guid"
$email = "$username@example.test"
$password = "TestPass1!"
$changedPassword = "ChangedPass2!"
$loginToken = $null
$changedLoginToken = $null

Write-Host "Pruebas funcionales de Auth Service mediante API Gateway" -ForegroundColor Cyan

$response = Send-ApiRequest -Path "/auth/register" -Method "POST" -Body @{}
Record-ApiCase -Name "Registro rechaza campos requeridos vacios" -Response $response -ExpectedStatus 400 -ExpectedSuccess $false -ExpectedErrorCode "VALIDATION_ERROR" | Out-Null

$response = Send-ApiRequest -Path "/auth/register" -Method "POST" -Body @{
    username = "auth-test-invalid-$guid"
    email = "not-an-email"
    password = $password
}
Record-ApiCase -Name "Registro rechaza correo invalido" -Response $response -ExpectedStatus 400 -ExpectedSuccess $false -ExpectedErrorCode "VALIDATION_ERROR" | Out-Null

$response = Send-ApiRequest -Path "/auth/login" -Method "POST" -Body @{}
Record-ApiCase -Name "Login valida campos requeridos" -Response $response -ExpectedStatus 400 -ExpectedSuccess $false -ExpectedErrorCode "VALIDATION_ERROR" | Out-Null

$recoveryEmail = "not-registered-$guid@example.test"
$response = Send-ApiRequest -Path "/auth/password-recovery" -Method "POST" -Body @{ email = "not-an-email" }
Record-ApiCase -Name "Recuperacion rechaza correo invalido" -Response $response -ExpectedStatus 400 -ExpectedSuccess $false -ExpectedErrorCode "VALIDATION_ERROR" | Out-Null

$response = Send-ApiRequest -Path "/auth/change-password" -Method "POST" -Body @{
    current_password = "TestPass1!"
    new_password = "ChangedPass2!"
}
Record-ApiCase -Name "Cambio de contrasena requiere autenticacion" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null

$registration = Send-ApiRequest -Path "/auth/register" -Method "POST" -Body @{
    username = $username
    email = $email
    password = $password
}
$registrationSucceeded = Record-ApiCase -Name "Registro de usuario unico" -Response $registration -ExpectedStatus 201 -ExpectedSuccess $true -ExpectedData @{
    username = $username
    email = $email
} -RequiredData @("id")

if ($registrationSucceeded) {
    $response = Send-ApiRequest -Path "/auth/register" -Method "POST" -Body @{
        username = $username
        email = $email
        password = $password
    }
    Record-ApiCase -Name "Registro duplicado devuelve conflicto" -Response $response -ExpectedStatus 409 -ExpectedSuccess $false -ExpectedErrorCode "CONFLICT" | Out-Null

    $response = Send-ApiRequest -Path "/auth/login" -Method "POST" -Body @{
        email = $email
        password = "WrongPass9!"
    }
    Record-ApiCase -Name "Login rechaza contrasena incorrecta" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null

    $login = Send-ApiRequest -Path "/auth/login" -Method "POST" -Body @{
        email = $email
        password = $password
    }
    $loginSucceeded = Record-ApiCase -Name "Login con credenciales correctas" -Response $login -ExpectedStatus 200 -ExpectedSuccess $true -ExpectedData @{
        "user.username" = $username
        "user.email" = $email
    } -RequiredData @("session_token")
    if ($loginSucceeded) {
        $loginToken = [string](Get-JsonPathValue -Object $login.Json -Path "data.session_token")
    }

    if ($null -ne $loginToken) {
        $authHeader = @{ Authorization = "Bearer $loginToken" }
        $response = Send-ApiRequest -Path "/auth/validate" -Method "GET" -Headers $authHeader
        Record-ApiCase -Name "Validacion con token valido" -Response $response -ExpectedStatus 200 -ExpectedSuccess $true -ExpectedData @{
            username = $username
            email = $email
        } | Out-Null

        $response = Send-ApiRequest -Path "/auth/validate" -Method "GET"
        Record-ApiCase -Name "Validacion sin token" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null

        $response = Send-ApiRequest -Path "/auth/validate" -Method "GET" -Headers @{ Authorization = "Bearer invalid-test-token" }
        Record-ApiCase -Name "Validacion con token invalido" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null

        $response = Send-ApiRequest -Path "/auth/change-password" -Method "POST" -Body @{
            current_password = $password
            new_password = "short"
        } -Headers $authHeader
        Record-ApiCase -Name "Cambio rechaza contrasena nueva demasiado corta" -Response $response -ExpectedStatus 400 -ExpectedSuccess $false -ExpectedErrorCode "VALIDATION_ERROR" -ExpectedErrorIssue "too_short" | Out-Null

        $response = Send-ApiRequest -Path "/auth/change-password" -Method "POST" -Body @{
            current_password = "WrongPass9!"
            new_password = $changedPassword
        } -Headers $authHeader
        Record-ApiCase -Name "Cambio rechaza contrasena actual incorrecta" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null

        $response = Send-ApiRequest -Path "/auth/change-password" -Method "POST" -Body @{
            current_password = $password
            new_password = $changedPassword
        } -Headers $authHeader
        Record-ApiCase -Name "Cambio de contrasena correcto" -Response $response -ExpectedStatus 200 -ExpectedSuccess $true | Out-Null

        $response = Send-ApiRequest -Path "/auth/login" -Method "POST" -Body @{
            email = $email
            password = $password
        }
        Record-ApiCase -Name "Login rechaza la contrasena anterior" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null

        $changedLogin = Send-ApiRequest -Path "/auth/login" -Method "POST" -Body @{
            email = $email
            password = $changedPassword
        }
        $changedLoginSucceeded = Record-ApiCase -Name "Login acepta la contrasena nueva" -Response $changedLogin -ExpectedStatus 200 -ExpectedSuccess $true -RequiredData @("session_token")
        if ($changedLoginSucceeded) {
            $changedLoginToken = [string](Get-JsonPathValue -Object $changedLogin.Json -Path "data.session_token")
        }

        if ($null -ne $changedLoginToken) {
            $changedAuthHeader = @{ Authorization = "Bearer $changedLoginToken" }
            $response = Send-ApiRequest -Path "/auth/logout" -Method "POST" -Headers $changedAuthHeader
            Record-ApiCase -Name "Logout con sesion valida" -Response $response -ExpectedStatus 200 -ExpectedSuccess $true | Out-Null

            $response = Send-ApiRequest -Path "/auth/validate" -Method "GET" -Headers $changedAuthHeader
            Record-ApiCase -Name "Token rechazado despues del logout" -Response $response -ExpectedStatus 401 -ExpectedSuccess $false -ExpectedErrorCode "UNAUTHORIZED" | Out-Null
        }
        else {
            Record-Skipped -Name "Logout y rechazo posterior al logout" -Reason "no se obtuvo una sesion valida con la contrasena nueva."
        }
    }
    else {
        Record-Skipped -Name "Validacion, cambio de contrasena y logout" -Reason "el login valido no devolvio una sesion utilizable."
    }
}
else {
    Record-Skipped -Name "Flujos que requieren el usuario de prueba" -Reason "el registro unico no se completo."
}

$response = Send-ApiRequest -Path "/auth/password-recovery" -Method "POST" -Body @{ email = $recoveryEmail }
Record-ApiCase -Name "Solicitud de recuperacion no revela si el correo existe" -Response $response -ExpectedStatus 200 -ExpectedSuccess $true | Out-Null

$response = Send-ApiRequest -Path "/auth/password-recovery/reset" -Method "POST" -Body @{
    token = "invalid-test-recovery-token"
    new_password = "ValidPass9!"
}
Record-ApiCase -Name "Restablecimiento rechaza token invalido" -Response $response -ExpectedStatus 400 -ExpectedSuccess $false -ExpectedErrorCode "VALIDATION_ERROR" -ExpectedErrorIssue "invalid" | Out-Null

Record-Skipped -Name "Recuperacion de cuenta registrada por correo" -Reason "requiere SMTP externo; el script no inspecciona el correo ni extrae tokens."
Record-Skipped -Name "Restablecimiento exitoso con token real" -Reason "el token solo se entrega por correo y no existe una interfaz de prueba segura para obtenerlo."

Write-Host ""
Write-Host "Resumen: $script:Passed aprobadas, $script:Failed fallidas, $script:Skipped omitidas." -ForegroundColor Cyan

if ($script:Failed -gt 0) {
    exit 1
}
exit 0

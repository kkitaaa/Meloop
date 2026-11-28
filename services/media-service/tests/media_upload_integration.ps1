param(
    [string]$GatewayUrl = "http://localhost:8080",
    [string]$PostApiUrl = "",
    [Parameter(Mandatory = $true)]
    [string]$BearerToken,
    [Parameter(Mandatory = $true)]
    [string]$UserID,
    [Parameter(Mandatory = $true)]
    [string]$OtherUserBearerToken,
    [Parameter(Mandatory = $true)]
    [string]$OtherUserID,
    [string]$PublicMediaEndpoint = "http://localhost:9000",
    [string]$ImagePath = ""
)

$ErrorActionPreference = "Stop"
$GatewayUrl = $GatewayUrl.TrimEnd("/")
$PublicMediaEndpoint = $PublicMediaEndpoint.TrimEnd("/")
if ([string]::IsNullOrWhiteSpace($PostApiUrl)) {
    $PostApiUrl = "$GatewayUrl/api/v1/posts"
}
$PostApiUrl = $PostApiUrl.TrimEnd("/")

if ([string]::IsNullOrWhiteSpace($ImagePath)) {
    $repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
    $ImagePath = Join-Path $repoRoot "app\flutter_app\assets\imagenes\meloop.png"
}
if (-not (Test-Path -LiteralPath $ImagePath -PathType Leaf)) {
    throw "No se encontró el archivo de prueba: $ImagePath"
}
if ($UserID -eq $OtherUserID) {
    throw "UserID y OtherUserID deben identificar usuarios distintos."
}

$extension = [IO.Path]::GetExtension($ImagePath).ToLowerInvariant()
switch ($extension) {
    ".jpg" { $contentType = "image/jpeg" }
    ".jpeg" { $contentType = "image/jpeg" }
    ".png" { $contentType = "image/png" }
    default { throw "El archivo debe ser JPEG o PNG; extensión recibida: $extension" }
}
$fileSize = (Get-Item -LiteralPath $ImagePath).Length
if ($fileSize -le 0 -or $fileSize -gt 5MB) {
    throw "La imagen válida debe ocupar entre 1 byte y 5 MB; tamaño: $fileSize"
}
$expectedPublicHost = ([Uri]$PublicMediaEndpoint).Host

function Invoke-Http {
    param(
        [string]$Uri,
        [string]$Method = "GET",
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [string]$ContentType = "application/json",
        [string]$InFile = ""
    )

    $request = @{
        Uri             = $Uri
        Method          = $Method
        Headers         = $Headers
        UseBasicParsing = $true
        ErrorAction     = "Stop"
    }
    if ($null -ne $Body) {
        $request.Body = $Body
        $request.ContentType = $ContentType
    }
    if (-not [string]::IsNullOrWhiteSpace($InFile)) {
        $request.InFile = $InFile
        $request.ContentType = $ContentType
    }

    try {
        $response = Invoke-WebRequest @request
        return @{
            StatusCode = [int]$response.StatusCode
            Content    = $response.Content
            ByteLength = [long]$response.RawContentLength
        }
    } catch {
        $webResponse = $_.Exception.Response
        if ($null -eq $webResponse) {
            throw "Falló $Method $Uri sin respuesta HTTP: $($_.Exception.Message)"
        }
        $reader = New-Object IO.StreamReader($webResponse.GetResponseStream())
        try {
            $content = $reader.ReadToEnd()
        } finally {
            $reader.Dispose()
        }
        return @{
            StatusCode = [int]$webResponse.StatusCode
            Content    = $content
            ByteLength = [long]$webResponse.ContentLength
        }
    }
}

function Assert-Status {
    param($Response, [int]$ExpectedStatus, [string]$Operation)
    if ($Response.StatusCode -ne $ExpectedStatus) {
        throw "$Operation devolvió HTTP $($Response.StatusCode), se esperaba $ExpectedStatus. Respuesta: $($Response.Content)"
    }
}

function Get-ResponseData {
    param($Response, [string]$Operation)
    $parsed = $Response.Content | ConvertFrom-Json
    if ($parsed.PSObject.Properties.Name -contains "success") {
        if (-not $parsed.success) {
            throw "$Operation no tuvo éxito: $($Response.Content)"
        }
        return $parsed.data
    }
    return $parsed
}

function Assert-PublicHost {
    param([string]$Url, [string]$Operation)
    $hostUri = [Uri]$Url
    if ($hostUri.Host -ne $expectedPublicHost -or $hostUri.Scheme -ne ([Uri]$PublicMediaEndpoint).Scheme) {
        throw "$Operation usa '$($hostUri.Host)' en vez del host público configurado '$expectedPublicHost'."
    }
    if ($hostUri.Host -eq "minio") {
        throw "$Operation expone el nombre interno minio."
    }
}

function New-PresignedUpload {
    param(
        [string]$ResourceType,
        [string]$MimeType,
        [long]$Size,
        [string]$Token
    )
    $headers = @{ Authorization = "Bearer $Token" }
    $body = @{
        resource_type = $ResourceType
        content_type  = $MimeType
        file_name     = [IO.Path]::GetFileName($ImagePath)
        file_size     = $Size
    } | ConvertTo-Json -Compress
    return Invoke-Http -Uri "$GatewayUrl/media/upload" -Method "POST" -Headers $headers -Body $body
}

function Upload-AndVerify {
    param([string]$ResourceType, [string]$Token)

    $response = New-PresignedUpload -ResourceType $ResourceType -MimeType $contentType -Size $fileSize -Token $Token
    Assert-Status $response 200 "Solicitar URL prefirmada para $ResourceType"
    $presigned = Get-ResponseData $response "Solicitud de URL prefirmada"

    Assert-PublicHost -Url $presigned.upload_url -Operation "La URL PUT prefirmada"

    $put = Invoke-Http -Uri $presigned.upload_url -Method "PUT" -InFile $ImagePath -ContentType $contentType
    Assert-Status $put 200 "Subir objeto $($presigned.object_key) con PUT"

    $verifyBody = @{ object_key = $presigned.object_key } | ConvertTo-Json -Compress
    $verified = Invoke-Http -Uri "$GatewayUrl/media/verify" -Method "POST" `
        -Headers @{ Authorization = "Bearer $Token" } -Body $verifyBody
    Assert-Status $verified 200 "Verificar objeto $($presigned.object_key)"
    $verification = Get-ResponseData $verified "Verificación de objeto"
    if (-not $verification.exists) {
        throw "media-service informó que no existe el objeto $($presigned.object_key)."
    }
    if ([long]$verification.size_bytes -ne $fileSize) {
        throw "El objeto verificado mide $($verification.size_bytes) bytes; se esperaban $fileSize."
    }
    return $presigned
}

$headers = @{ Authorization = "Bearer $BearerToken" }
$otherHeaders = @{ Authorization = "Bearer $OtherUserBearerToken" }

foreach ($identity in @(
    @{ Token = $BearerToken; UserID = $UserID; Headers = $headers },
    @{ Token = $OtherUserBearerToken; UserID = $OtherUserID; Headers = $otherHeaders }
)) {
    $profile = Invoke-Http -Uri "$GatewayUrl/v1/users/me/profile" -Headers $identity.Headers
    Assert-Status $profile 200 "Validar sesión del usuario $($identity.UserID)"
    $profileData = Get-ResponseData $profile "Consultar perfil autenticado"
    if ($profileData.id_usuario -ne $identity.UserID) {
        throw "El token proporcionado no pertenece al usuario indicado ($($identity.UserID))."
    }
}

Write-Host "1/6 Rechazar tipo MIME no permitido..."
$invalidType = New-PresignedUpload -ResourceType "profile_photo" -MimeType "application/pdf" -Size 1024 -Token $BearerToken
Assert-Status $invalidType 400 "Validación de tipo MIME"

Write-Host "2/6 Rechazar tamaño superior a 5 MB..."
$oversized = New-PresignedUpload -ResourceType "profile_photo" -MimeType $contentType -Size (5MB + 1) -Token $BearerToken
Assert-Status $oversized 400 "Validación de tamaño"

Write-Host "3/6 Subir y verificar foto de perfil..."
$profileUpload = Upload-AndVerify -ResourceType "profile_photo" -Token $BearerToken
$profileUpdate = Invoke-Http -Uri "$GatewayUrl/v1/users/me/profile" -Method "PUT" -Headers $headers `
    -Body (@{ foto_perfil = $profileUpload.object_key } | ConvertTo-Json -Compress)
Assert-Status $profileUpdate 200 "Asociar foto al perfil del propietario"
$profileDownload = Invoke-Http -Uri "$GatewayUrl/media/download-url?key=$([Uri]::EscapeDataString($profileUpload.object_key))" `
    -Headers $headers
Assert-Status $profileDownload 200 "Solicitar descarga privada de la foto"
$profileDownloadData = Get-ResponseData $profileDownload "URL de descarga de foto"
Assert-PublicHost -Url $profileDownloadData.download_url -Operation "La URL GET prefirmada de la foto"
$profileBytes = Invoke-Http -Uri $profileDownloadData.download_url
Assert-Status $profileBytes 200 "Descargar foto con URL prefirmada"
if ($profileBytes.ByteLength -ne $fileSize) {
    throw "La descarga de la foto mide $($profileBytes.ByteLength) bytes; se esperaban $fileSize."
}

Write-Host "4/6 Rechazar asociación de objeto de otro usuario..."
$foreignProfileUpdate = Invoke-Http -Uri "$GatewayUrl/v1/users/me/profile" -Method "PUT" -Headers $otherHeaders `
    -Body (@{ foto_perfil = $profileUpload.object_key } | ConvertTo-Json -Compress)
Assert-Status $foreignProfileUpdate 400 "Asociar foto al perfil de otro usuario"
$foreignError = $foreignProfileUpdate.Content | ConvertFrom-Json
if ($foreignError.error.details.issue -ne "not_owned") {
    throw "El objeto ajeno se rechazó por una causa inesperada: $($foreignProfileUpdate.Content)"
}

Write-Host "5/6 Subir y verificar adjunto de publicación..."
$postUpload = Upload-AndVerify -ResourceType "post_attachment" -Token $BearerToken
$postBody = @{
    content    = "Prueba de integración de adjunto multimedia"
    media_urls = @($postUpload.object_key)
} | ConvertTo-Json -Compress
$createdPost = Invoke-Http -Uri $PostApiUrl -Method "POST" -Headers $headers -Body $postBody
Assert-Status $createdPost 201 "Crear publicación con adjunto"
$postData = $createdPost.Content | ConvertFrom-Json
if ($postData.url_multimedia -ne $postUpload.object_key) {
    throw "La publicación no persistió la clave estable del adjunto: $($createdPost.Content)"
}

Write-Host "6/6 Confirmar URL visible públicamente..."
$postDownload = Invoke-Http -Uri "$GatewayUrl/media/download-url?key=$([Uri]::EscapeDataString($postUpload.object_key))" `
    -Headers $headers
Assert-Status $postDownload 200 "Solicitar descarga privada del adjunto"
$postDownloadData = Get-ResponseData $postDownload "URL de descarga del adjunto"
Assert-PublicHost -Url $postDownloadData.download_url -Operation "La URL GET prefirmada del adjunto"
$postBytes = Invoke-Http -Uri $postDownloadData.download_url
Assert-Status $postBytes 200 "Descargar adjunto con URL prefirmada"
if ($postBytes.ByteLength -ne $fileSize) {
    throw "La descarga del adjunto mide $($postBytes.ByteLength) bytes; se esperaban $fileSize."
}

Write-Host "Integración completada: perfil, adjunto de publicación, validaciones y propiedad del objeto."
Write-Host "Clave de perfil: $($profileUpload.object_key)"
Write-Host "Clave del adjunto: $($postUpload.object_key)"
Write-Host "La app solicita URLs GET temporales para ambas claves."

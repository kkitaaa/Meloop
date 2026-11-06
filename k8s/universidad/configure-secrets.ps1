$ErrorActionPreference = 'Stop'

$namespace = 'student-acofre'
$secretNames = @(
    'meloop-postgres-credentials',
    'meloop-redis-credentials',
    'meloop-auth-connection'
)

$existingSecrets = @()
foreach ($secretName in $secretNames) {
    $existing = & kubectl get secret $secretName --namespace $namespace --ignore-not-found --output name
    if ($LASTEXITCODE -ne 0) {
        throw "No se pudo consultar el Secret '$secretName'."
    }
    if ($existing) {
        $existingSecrets += $secretName
    }
}

if ($existingSecrets.Count -eq $secretNames.Count) {
    Write-Output 'Los tres Secrets ya existen; no se rotaron las credenciales.'
    return
}
if ($existingSecrets.Count -gt 0) {
    throw 'Hay Secrets parciales. Revísalos antes de continuar; no se generaron credenciales nuevas.'
}

function New-RandomSecretValue {
    $bytes = New-Object byte[] 32
    $generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $generator.GetBytes($bytes)
    }
    finally {
        $generator.Dispose()
    }

    return [Convert]::ToBase64String($bytes).TrimEnd('=').Replace('+', '-').Replace('/', '_')
}

function Apply-Secret {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string[]]$Literals
    )

    $createArgs = @(
        'create', 'secret', 'generic', $Name,
        '--namespace', $namespace
    ) + $Literals + @('--dry-run=client', '--output=yaml')

    $secretYaml = & kubectl @createArgs
    if ($LASTEXITCODE -ne 0) {
        throw "No se pudo generar el manifiesto del Secret '$Name'."
    }

    $secretYaml | & kubectl apply --filename -
    if ($LASTEXITCODE -ne 0) {
        throw "No se pudo aplicar el Secret '$Name'."
    }
}

$postgresPassword = New-RandomSecretValue
$redisPassword = New-RandomSecretValue
$databaseUrl = "postgres://meloop_auth:${postgresPassword}@meloop-postgres-service:5432/meloop?sslmode=disable"
$redisUrl = "redis://:${redisPassword}@meloop-redis-service:6379/0"

Apply-Secret -Name 'meloop-postgres-credentials' -Literals @(
    '--from-literal=POSTGRES_DB=meloop',
    '--from-literal=POSTGRES_USER=meloop_auth',
    "--from-literal=POSTGRES_PASSWORD=$postgresPassword"
)
Apply-Secret -Name 'meloop-redis-credentials' -Literals @(
    "--from-literal=REDIS_PASSWORD=$redisPassword"
)
Apply-Secret -Name 'meloop-auth-connection' -Literals @(
    "--from-literal=DATABASE_URL=$databaseUrl",
    "--from-literal=REDIS_URL=$redisUrl"
)

Write-Output 'Credenciales generadas y guardadas en Kubernetes Secrets; no se imprimieron ni escribieron en archivos.'

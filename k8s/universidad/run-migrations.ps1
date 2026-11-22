$ErrorActionPreference = 'Stop'

$namespace = 'student-acofre'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$migrationsPath = Join-Path $repoRoot 'supabase\migrations'
$migrationFiles = @(Get-ChildItem -Path $migrationsPath -Filter '*.sql' -File | Sort-Object Name)

if ($migrationFiles.Count -ne 6) {
    throw "Se esperaban exactamente seis migraciones; se encontraron $($migrationFiles.Count)."
}

$configMapArgs = @(
    'create', 'configmap', 'meloop-auth-migrations',
    '--namespace', $namespace
)
foreach ($migrationFile in $migrationFiles) {
    $configMapArgs += "--from-file=$($migrationFile.FullName)"
}
$configMapArgs += @('--dry-run=client', '--output=yaml')

$configMapYaml = & kubectl @configMapArgs
if ($LASTEXITCODE -ne 0) {
    throw 'No se pudo construir el ConfigMap con las migraciones.'
}

$configMapYaml | & kubectl apply --filename -
if ($LASTEXITCODE -ne 0) {
    throw 'No se pudo aplicar el ConfigMap de migraciones.'
}

& kubectl delete job meloop-auth-migrations --namespace $namespace --ignore-not-found --wait=true
if ($LASTEXITCODE -ne 0) {
    throw 'No se pudo eliminar una ejecución anterior del Job de migraciones.'
}

& kubectl apply --filename (Join-Path $PSScriptRoot 'migrations-job.yaml')
if ($LASTEXITCODE -ne 0) {
    throw 'No se pudo crear el Job de migraciones.'
}

Write-Output 'Job creado. Sigue el resultado con kubectl wait y consulta los logs si falla.'

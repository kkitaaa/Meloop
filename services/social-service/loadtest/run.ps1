$ErrorActionPreference = 'Stop'

$projectName = 'meloop-social-loadtest'
$repositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$composeFile = Join-Path $PSScriptRoot 'docker-compose.yml'
$scriptFile = Join-Path $PSScriptRoot 'social.js'
$artifactDirectory = Join-Path $env:TEMP 'meloop-social-loadtest'
$runId = Get-Date -Format 'yyyyMMdd-HHmmss'
$tokenFile = Join-Path $artifactDirectory "sessions-$runId.json"
$summaryFile = Join-Path $artifactDirectory "summary-$runId.json"
$metricsFile = Join-Path $artifactDirectory "metrics-$runId.json"
$gatewayUrl = 'http://127.0.0.1:18080'

$environmentNames = @(
    'LOADTEST_USER_PASSWORD',
    'LOADTEST_DATABASE_URL',
    'LOADTEST_AUTH_URL',
    'LOADTEST_TOKEN_FILE'
)
$previousEnvironment = @{}
foreach ($name in $environmentNames) {
    $previousEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}

$secretPointer = [IntPtr]::Zero
$k6ExitCode = 1

try {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw 'Docker CLI is required.'
    }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw 'Go is required to prepare the isolated dataset and sessions.'
    }
    if (-not (Get-Command k6 -ErrorAction SilentlyContinue)) {
        throw 'k6 is required to run the load test.'
    }

    New-Item -ItemType Directory -Path $artifactDirectory -Force | Out-Null

    & docker compose --project-name $projectName --file $composeFile up --detach --build --wait
    if ($LASTEXITCODE -ne 0) {
        throw 'The dedicated load-test Compose environment failed to start.'
    }

    $gatewayReady = $false
    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        try {
            $health = Invoke-WebRequest -Uri "$gatewayUrl/health" -UseBasicParsing -TimeoutSec 5
            if ($health.StatusCode -eq 200) {
                $gatewayReady = $true
                break
            }
        } catch {
            Start-Sleep -Seconds 2
        }
    }
    if (-not $gatewayReady) {
        throw 'The dedicated API Gateway did not become ready.'
    }

    $plainPassword = [Environment]::GetEnvironmentVariable('LOADTEST_USER_PASSWORD', 'Process')
    if ([string]::IsNullOrEmpty($plainPassword)) {
        $securePassword = Read-Host 'Enter a load-test-only password (8-72 characters)' -AsSecureString
        $secretPointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
        $plainPassword = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($secretPointer)
    }
    [Environment]::SetEnvironmentVariable('LOADTEST_USER_PASSWORD', $plainPassword, 'Process')
    [Environment]::SetEnvironmentVariable(
        'LOADTEST_DATABASE_URL',
        'postgres://postgres@127.0.0.1:15432/meloop_social_loadtest?sslmode=disable',
        'Process'
    )
    [Environment]::SetEnvironmentVariable('LOADTEST_AUTH_URL', 'http://127.0.0.1:18083', 'Process')
    [Environment]::SetEnvironmentVariable('LOADTEST_TOKEN_FILE', $tokenFile, 'Process')

    Push-Location (Join-Path $repositoryRoot 'services\social-service')
    try {
        & go run .\loadtest\seed.go
        if ($LASTEXITCODE -ne 0) {
            throw 'Dataset and session preparation failed.'
        }
    } finally {
        Pop-Location
    }

    [Environment]::SetEnvironmentVariable('LOADTEST_USER_PASSWORD', $null, 'Process')
    $plainPassword = $null

    $sessions = Get-Content -LiteralPath $tokenFile -Raw | ConvertFrom-Json
    if ($sessions -isnot [System.Array]) {
        $sessions = @($sessions)
    }
    if ($sessions.Count -ne 200) {
        throw 'Preparation did not produce exactly 200 sessions.'
    }
    $sample = $sessions[0]
    $authorization = @{ Authorization = "Bearer $($sample.token)" }

    try {
        $friends = Invoke-RestMethod -Uri "$gatewayUrl/friends" -Headers $authorization -TimeoutSec 15
        $suggestions = Invoke-RestMethod -Uri "$gatewayUrl/friends/suggestions" -Headers $authorization -TimeoutSec 30
    } catch {
        throw 'Functional preflight through the Gateway failed; no load test was started.'
    }

    $actualFriendIDs = @($friends.data | ForEach-Object { $_.id_usuario } | Sort-Object)
    $expectedFriendIDs = @($sample.expected_friend_ids | Sort-Object)
    if ($friends.success -ne $true -or $actualFriendIDs.Count -ne 30 -or
        (($actualFriendIDs -join ',') -ne ($expectedFriendIDs -join ','))) {
        throw 'The friends endpoint did not match the seeded fixture during preflight.'
    }
    $suggestionList = @($suggestions.data)
    $knownCandidate = $suggestionList |
        Where-Object { $_.id_usuario -eq $sample.expected_suggestion_id } |
        Select-Object -First 1
    if ($suggestions.success -ne $true -or $suggestionList.Count -ne 469 -or
        $null -eq $knownCandidate -or
        $knownCandidate.amigos_en_comun -ne $sample.expected_mutual_friends) {
        throw 'Suggestions or the expected amigos_en_comun value failed preflight.'
    }

    Write-Host 'Preflight passed: Gateway, authenticated friends, and suggestions verified.'
    Write-Host 'Starting k6 with 200 constant VUs for 2 minutes.'

    & k6 run `
        -e "TOKEN_FILE=$tokenFile" `
        --summary-export $summaryFile `
        --out "json=$metricsFile" `
        $scriptFile
    $k6ExitCode = $LASTEXITCODE

    Write-Host "k6 summary: $summaryFile"
    Write-Host "k6 metrics: $metricsFile"
    if ($k6ExitCode -ne 0) {
        Write-Warning "k6 returned exit code $k6ExitCode; inspect the saved thresholds and checks."
    }
} finally {
    $plainPassword = $null
    if ($secretPointer -ne [IntPtr]::Zero) {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($secretPointer)
    }
    foreach ($name in $environmentNames) {
        if ($null -eq $previousEnvironment[$name]) {
            Remove-Item "Env:\$name" -ErrorAction SilentlyContinue
        } else {
            [Environment]::SetEnvironmentVariable($name, $previousEnvironment[$name], 'Process')
        }
    }
    if (Test-Path -LiteralPath $tokenFile) {
        Remove-Item -LiteralPath $tokenFile -Force
    }
}

exit $k6ExitCode

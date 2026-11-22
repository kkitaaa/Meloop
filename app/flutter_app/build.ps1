# Flutter build script for different environments (PowerShell)
# Usage: .\build.ps1 -Environment dev -Platform android
# Usage: .\build.ps1 -Environment staging -Platform android

param(
    [Parameter(Position = 0)]
    [ValidateSet('dev', 'development', 'staging', 'prod', 'production')]
    [string]$Environment = 'dev',
    
    [Parameter(Position = 1)]
    [ValidateSet('android', 'windows', 'ios', 'all')]
    [string]$Platform = 'android'
)

$ErrorActionPreference = 'Stop'

# Configuration mapping
switch ($Environment) {
    { $_ -in 'dev', 'development' } {
        $flutterEnv = 'development'
        $apiBaseUrl = 'http://localhost:8080'
        $k8sApiGatewayUrl = ''
        $k8sNamespace = 'meloop'
        $k8sIngressHost = 'meloop.local'
        $buildName = '1.0.0'
        $buildNumber = '1'
        break
    }
    'staging' {
        $flutterEnv = 'staging'
        $apiBaseUrl = ''
        $k8sApiGatewayUrl = $env:K8S_API_GATEWAY_URL_STAGING ?? 'http://staging.meloop.com'
        $k8sNamespace = 'meloop'
        $k8sIngressHost = $env:K8S_INGRESS_HOST_STAGING ?? 'staging.meloop.com'
        $buildName = '1.0.0'
        $buildNumber = '2'
        break
    }
    { $_ -in 'prod', 'production' } {
        $flutterEnv = 'production'
        $apiBaseUrl = ''
        $k8sApiGatewayUrl = $env:K8S_API_GATEWAY_URL_PROD ?? 'https://api.meloop.com'
        $k8sNamespace = 'meloop'
        $k8sIngressHost = $env:K8S_INGRESS_HOST_PROD ?? 'api.meloop.com'
        $buildName = '1.0.0'
        $buildNumber = '3'
        break
    }
    default {
        Write-Host "Unknown environment: $Environment" -ForegroundColor Red
        exit 1
    }
}

# Display configuration
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "Flutter Build Configuration" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "Environment:       $flutterEnv"
Write-Host "Platform:          $Platform"
Write-Host "API Base URL:      $apiBaseUrl"
Write-Host "K8s API Gateway:   $k8sApiGatewayUrl"
Write-Host "K8s Namespace:     $k8sNamespace"
Write-Host "K8s Ingress Host:  $k8sIngressHost"
Write-Host "Build Name:        $buildName"
Write-Host "Build Number:      $buildNumber"
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan

# Build dart defines
$dartDefines = @(
    "--dart-define=FLUTTER_ENV=$flutterEnv",
    "--dart-define=K8S_NAMESPACE=$k8sNamespace",
    "--dart-define=K8S_INGRESS_HOST=$k8sIngressHost"
)

if ($apiBaseUrl) {
    $dartDefines += "--dart-define=API_BASE_URL=$apiBaseUrl"
}

if ($k8sApiGatewayUrl) {
    $dartDefines += "--dart-define=K8S_API_GATEWAY_URL=$k8sApiGatewayUrl"
}

# Build based on platform
switch ($Platform) {
    'android' {
        Write-Host "Building for Android..." -ForegroundColor Green
        & flutter build apk `
            --build-name=$buildName `
            --build-number=$buildNumber `
            @dartDefines `
            --release
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ APK built: build\app\outputs\flutter-apk\app-release.apk" -ForegroundColor Green
        }
        else {
            Write-Host "✗ Android build failed" -ForegroundColor Red
            exit 1
        }
        break
    }
    'windows' {
        Write-Host "Building for Windows..." -ForegroundColor Green
        & flutter build windows @dartDefines
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Windows build completed: build\windows\runner\Release\" -ForegroundColor Green
        }
        else {
            Write-Host "✗ Windows build failed" -ForegroundColor Red
            exit 1
        }
        break
    }
    'ios' {
        Write-Host "Building for iOS..." -ForegroundColor Green
        & flutter build ios `
            --build-name=$buildName `
            --build-number=$buildNumber `
            @dartDefines `
            --release
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ iOS build completed: build\ios\iphoneos\" -ForegroundColor Green
        }
        else {
            Write-Host "✗ iOS build failed" -ForegroundColor Red
            exit 1
        }
        break
    }
    'all' {
        Write-Host "Building for all platforms..." -ForegroundColor Green
        
        Write-Host "`n1. Building for Android..." -ForegroundColor Yellow
        & flutter build apk `
            --build-name=$buildName `
            --build-number=$buildNumber `
            @dartDefines `
            --release
        if ($LASTEXITCODE -ne 0) {
            Write-Host "✗ Android build failed" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "`n2. Building for Windows..." -ForegroundColor Yellow
        & flutter build windows @dartDefines
        if ($LASTEXITCODE -ne 0) {
            Write-Host "✗ Windows build failed" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "`n✓ All builds completed!" -ForegroundColor Green
        break
    }
    default {
        Write-Host "Unknown platform: $Platform" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "Build completed successfully!" -ForegroundColor Green

# Flutter build script for different environments (PowerShell)
# Compatible with PowerShell 5.1 and 7
# Usage: .\build.ps1 -Environment dev -Platform android
# Usage: .\build.ps1 -Environment staging -Platform windows

param(
    [Parameter(Position = 0)]
    [ValidateSet('dev', 'development', 'staging', 'prod', 'production')]
    [string]$Environment = 'dev',
    
    [Parameter(Position = 1)]
    [ValidateSet('android', 'windows', 'ios', 'all')]
    [string]$Platform = 'android'
)

$ErrorActionPreference = 'Stop'

function GetEnvironmentConfig {
    param([string]$env)
    
    switch ($env) {
        'dev' {
            return @{
                flutterEnv = 'development'
                apiBaseUrl = 'http://localhost:8080'
                k8sApiGatewayUrl = ''
                k8sNamespace = 'meloop'
                k8sIngressHost = 'meloop.local'
                buildName = '1.0.0'
                buildNumber = '1'
            }
        }
        'staging' {
            $k8sUrl = if ([string]::IsNullOrEmpty($env:K8S_API_GATEWAY_URL_STAGING)) { 'http://staging.meloop.com' } else { $env:K8S_API_GATEWAY_URL_STAGING }
            $ingressHost = if ([string]::IsNullOrEmpty($env:K8S_INGRESS_HOST_STAGING)) { 'staging.meloop.com' } else { $env:K8S_INGRESS_HOST_STAGING }
            return @{
                flutterEnv = 'staging'
                apiBaseUrl = ''
                k8sApiGatewayUrl = $k8sUrl
                k8sNamespace = 'meloop'
                k8sIngressHost = $ingressHost
                buildName = '1.0.0'
                buildNumber = '2'
            }
        }
        'prod' {
            $k8sUrl = if ([string]::IsNullOrEmpty($env:K8S_API_GATEWAY_URL_PROD)) { 'https://api.meloop.com' } else { $env:K8S_API_GATEWAY_URL_PROD }
            $ingressHost = if ([string]::IsNullOrEmpty($env:K8S_INGRESS_HOST_PROD)) { 'api.meloop.com' } else { $env:K8S_INGRESS_HOST_PROD }
            return @{
                flutterEnv = 'production'
                apiBaseUrl = ''
                k8sApiGatewayUrl = $k8sUrl
                k8sNamespace = 'meloop'
                k8sIngressHost = $ingressHost
                buildName = '1.0.0'
                buildNumber = '3'
            }
        }
        default {
            Write-Host "Unknown environment: $env" -ForegroundColor Red
            exit 1
        }
    }
}

function BuildAndroid {
    param($config, [string]$dartDefines)
    
    Write-Host "Building for Android..." -ForegroundColor Green
    & flutter build apk `
        --build-name=$($config.buildName) `
        --build-number=$($config.buildNumber) `
        @dartDefines `
        --release
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "OK: APK built: build/app/outputs/flutter-apk/app-release.apk" -ForegroundColor Green
        return $true
    } else {
        Write-Host "ERROR: Android build failed" -ForegroundColor Red
        return $false
    }
}

function BuildWindows {
    param($config, [string]$dartDefines)
    
    Write-Host "Building for Windows..." -ForegroundColor Green
    & flutter build windows @dartDefines
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "OK: Windows build completed: build/windows/runner/Release/" -ForegroundColor Green
        return $true
    } else {
        Write-Host "ERROR: Windows build failed" -ForegroundColor Red
        return $false
    }
}

function BuildiOS {
    param($config, [string]$dartDefines)
    
    Write-Host "Building for iOS..." -ForegroundColor Green
    & flutter build ios `
        --build-name=$($config.buildName) `
        --build-number=$($config.buildNumber) `
        @dartDefines `
        --release
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "OK: iOS build completed: build/ios/iphoneos/" -ForegroundColor Green
        return $true
    } else {
        Write-Host "ERROR: iOS build failed" -ForegroundColor Red
        return $false
    }
}

# Main script
$normalizedEnv = if ($Environment -in 'prod', 'production') { 'prod' } else { $Environment }
$config = GetEnvironmentConfig $normalizedEnv

# Display configuration
Write-Host "`n===============================================" -ForegroundColor Cyan
Write-Host "Flutter Build Configuration" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "Environment:       $($config.flutterEnv)"
Write-Host "Platform:          $Platform"
Write-Host "API Base URL:      $($config.apiBaseUrl)"
Write-Host "K8s API Gateway:   $($config.k8sApiGatewayUrl)"
Write-Host "K8s Namespace:     $($config.k8sNamespace)"
Write-Host "K8s Ingress Host:  $($config.k8sIngressHost)"
Write-Host "Build Name:        $($config.buildName)"
Write-Host "Build Number:      $($config.buildNumber)"
Write-Host "===============================================" -ForegroundColor Cyan

# Build dart defines
$dartDefines = @(
    "--dart-define=FLUTTER_ENV=$($config.flutterEnv)",
    "--dart-define=K8S_NAMESPACE=$($config.k8sNamespace)",
    "--dart-define=K8S_INGRESS_HOST=$($config.k8sIngressHost)"
)

if ($config.apiBaseUrl) {
    $dartDefines += "--dart-define=API_BASE_URL=$($config.apiBaseUrl)"
}

if ($config.k8sApiGatewayUrl) {
    $dartDefines += "--dart-define=K8S_API_GATEWAY_URL=$($config.k8sApiGatewayUrl)"
}

# Execute build
$success = $true
switch ($Platform) {
    'android' {
        $success = BuildAndroid $config $dartDefines
    }
    'windows' {
        $success = BuildWindows $config $dartDefines
    }
    'ios' {
        $success = BuildiOS $config $dartDefines
    }
    'all' {
        Write-Host "`nBuilding all platforms..." -ForegroundColor Cyan
        Write-Host "`n[1/3] Building for Android..." -ForegroundColor Yellow
        $success = BuildAndroid $config $dartDefines
        
        if ($success) {
            Write-Host "`n[2/3] Building for Windows..." -ForegroundColor Yellow
            $success = BuildWindows $config $dartDefines
        }
        
        if ($success) {
            Write-Host "`n[3/3] Done!" -ForegroundColor Green
        }
    }
    default {
        Write-Host "Unknown platform: $Platform" -ForegroundColor Red
        exit 1
    }
}

if ($success) {
    Write-Host "`nBuild completed successfully!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}

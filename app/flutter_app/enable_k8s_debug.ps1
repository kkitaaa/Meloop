# Script para activar/desactivar acceso a KubernetesConnectivityScreen
# Uso: .\enable_k8s_debug.ps1 -Action [enable|disable|status]

param(
    [ValidateSet("enable", "disable", "status")]
    [string]$Action = "status"
)

$flutterAppDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$mainFile = Join-Path $flutterAppDir "lib\main.dart"
$backupFile = Join-Path $flutterAppDir ".main.dart.backup"

# Color codes
$ColorGreen = "`e[32m"
$ColorRed = "`e[31m"
$ColorYellow = "`e[33m"
$ColorReset = "`e[0m"

function Enable-K8sDebug {
    Write-Host "${ColorYellow}Activando acceso a KubernetesConnectivityScreen...${ColorReset}"
    
    # Crear backup si no existe
    if (-not (Test-Path $backupFile)) {
        Copy-Item $mainFile $backupFile
        Write-Host "${ColorGreen}✓ Backup creado${ColorReset}"
    }
    
    # Leer archivo
    $content = Get-Content $mainFile -Raw
    
    # Verificar si ya está modificado
    if ($content -match "home: const KubernetesConnectivityScreen\(\),") {
        Write-Host "${ColorYellow}Ya está activado${ColorReset}"
        return
    }
    
    # Reemplazar AuthScreen() con KubernetesConnectivityScreen()
    $content = $content -replace "home: const AuthScreen\(\),", "home: const KubernetesConnectivityScreen(),"
    
    # Agregar import si no está
    if ($content -notmatch "import 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';") {
        # Encontrar la línea de auth_screen import y agregar después
        $content = $content -replace "(import 'package:flutter_app/screens/auth_screen.dart';)", "`$1`nimport 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';"
    }
    
    # Guardar archivo
    Set-Content $mainFile $content
    
    Write-Host "${ColorGreen}✓ KubernetesConnectivityScreen activado como home${ColorReset}"
    Write-Host "${ColorYellow}Próximo paso: ejecuta 'flutter run' o presiona 'r' en DevTools${ColorReset}"
}

function Disable-K8sDebug {
    Write-Host "${ColorYellow}Desactivando acceso a KubernetesConnectivityScreen...${ColorReset}"
    
    # Restaurar desde backup
    if (-not (Test-Path $backupFile)) {
        Write-Host "${ColorRed}✗ No hay backup, no se puede restaurar${ColorReset}"
        return
    }
    
    Copy-Item $backupFile $mainFile -Force
    Remove-Item $backupFile
    
    Write-Host "${ColorGreen}✓ AuthScreen restaurado como home${ColorReset}"
    Write-Host "${ColorYellow}Próximo paso: ejecuta 'flutter run' o presiona 'r' en DevTools${ColorReset}"
}

function Show-Status {
    $content = Get-Content $mainFile -Raw
    
    if ($content -match "home: const KubernetesConnectivityScreen\(\),") {
        Write-Host "${ColorGreen}✓ Debug activado (KubernetesConnectivityScreen es home)${ColorReset}"
    } else {
        Write-Host "${ColorYellow}○ Debug desactivado (AuthScreen es home)${ColorReset}"
    }
}

# Main
switch ($Action) {
    "enable" {
        Enable-K8sDebug
    }
    "disable" {
        Disable-K8sDebug
    }
    "status" {
        Show-Status
    }
    default {
        Write-Host "Uso: .\enable_k8s_debug.ps1 -Action [enable|disable|status]"
        Write-Host ""
        Write-Host "Ejemplos:"
        Write-Host "  .\enable_k8s_debug.ps1 -Action enable    # Activa KubernetesConnectivityScreen"
        Write-Host "  .\enable_k8s_debug.ps1 -Action disable   # Restaura AuthScreen"
        Write-Host "  .\enable_k8s_debug.ps1 -Action status    # Muestra estado actual"
        exit 1
    }
}

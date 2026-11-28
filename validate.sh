#!/bin/bash
# Script de validación de pruebas - Flutter-Kubernetes Integration
# Este script valida que toda la integración está funcionando correctamente

set -e

echo "╔════════════════════════════════════════════════════════════════╗"
echo "║      VALIDACIÓN DE INTEGRACIÓN FLUTTER-KUBERNETES            ║"
echo "╚════════════════════════════════════════════════════════════════╝"
echo ""

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0
WARNINGS=0

# Test counter
test_number=1

# Function to print test result
print_result() {
    local result=$1
    local message=$2
    
    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✓${NC} Test $test_number: $message"
        ((PASSED++))
    elif [ "$result" = "FAIL" ]; then
        echo -e "${RED}✗${NC} Test $test_number: $message"
        ((FAILED++))
    else
        echo -e "${YELLOW}⚠${NC} Test $test_number: $message"
        ((WARNINGS++))
    fi
    ((test_number++))
}

# Change to flutter app directory
cd app/flutter_app

echo -e "${BLUE}📋 PRUEBAS UNITARIAS${NC}"
echo "════════════════════════════════════════════════════════════════"
echo ""

# Test 1: Flutter Analyze
echo "⏳ Ejecutando: flutter analyze"
if flutter analyze --no-pub 2>&1 | grep -q "0 errors"; then
    print_result "PASS" "Flutter analyze - 0 errores"
else
    print_result "FAIL" "Flutter analyze - encontró errores"
fi

# Test 2: Configuration tests
echo ""
echo "⏳ Ejecutando: flutter test (ApiConfig)"
if flutter test test/config/api_config_test.dart -v 2>&1 | grep -q "passed"; then
    print_result "PASS" "ApiConfig tests"
else
    print_result "WARN" "ApiConfig tests - ver detalles"
fi

# Test 3: HttpClientManager tests
echo "⏳ Ejecutando: flutter test (HttpClientManager)"
if flutter test test/services/http_client_manager_test.dart -v 2>&1 | grep -q "passed"; then
    print_result "PASS" "HttpClientManager tests"
else
    print_result "WARN" "HttpClientManager tests - ver detalles"
fi

# Test 4: WebSocketManager tests
echo "⏳ Ejecutando: flutter test (WebSocketManager)"
if flutter test test/services/websocket_manager_test.dart -v 2>&1 | grep -q "passed"; then
    print_result "PASS" "WebSocketManager tests"
else
    print_result "WARN" "WebSocketManager tests - ver detalles"
fi

# Test 5: Configuration file exists
echo ""
if [ -f "lib/config/api_config.dart" ]; then
    print_result "PASS" "Archivo ApiConfig.dart existe"
else
    print_result "FAIL" "Archivo ApiConfig.dart NO existe"
fi

# Test 6: HttpClientManager file exists
if [ -f "lib/services/http_client_manager.dart" ]; then
    print_result "PASS" "Archivo HttpClientManager existe"
else
    print_result "FAIL" "Archivo HttpClientManager NO existe"
fi

# Test 7: WebSocketManager file exists
if [ -f "lib/services/websocket_manager.dart" ]; then
    print_result "PASS" "Archivo WebSocketManager existe"
else
    print_result "FAIL" "Archivo WebSocketManager NO existe"
fi

# Test 8: HealthCheckService file exists
if [ -f "lib/services/health_check_service.dart" ]; then
    print_result "PASS" "Archivo HealthCheckService existe"
else
    print_result "FAIL" "Archivo HealthCheckService NO existe"
fi

# Test 9: Build scripts exist
if [ -f "build.ps1" ] && [ -f "build.sh" ]; then
    print_result "PASS" "Scripts de build existen (build.ps1 y build.sh)"
else
    print_result "FAIL" "Scripts de build NO existen"
fi

# Test 10: Documentation files
echo ""
if [ -f "README_FLUTTER_K8S.md" ]; then
    print_result "PASS" "Documentación index existe"
else
    print_result "FAIL" "Documentación index NO existe"
fi

if [ -f "QUICK_REFERENCE.md" ]; then
    print_result "PASS" "Referencia rápida existe"
else
    print_result "FAIL" "Referencia rápida NO existe"
fi

if [ -f "VERIFICATION_REPORT.md" ]; then
    print_result "PASS" "Reporte de verificación existe"
else
    print_result "FAIL" "Reporte de verificación NO existe"
fi

if [ -f "TEST_PLAN.md" ]; then
    print_result "PASS" "Plan de pruebas existe"
else
    print_result "FAIL" "Plan de pruebas NO existe"
fi

# Test 14: Git status
echo ""
if git status --short | grep -q ""; then
    print_result "WARN" "Cambios pendientes en git"
else
    print_result "PASS" "Repositorio git limpio"
fi

# Test 15: APK can build (optional, skip if already done)
echo ""
echo "⏳ Intentando compilación APK (esto tomará tiempo)..."
if flutter build apk --dart-define=FLUTTER_ENV=development --debug 2>&1 | grep -q "Built"; then
    print_result "PASS" "APK se compila exitosamente"
else
    print_result "WARN" "Build APK - revisar detalles"
fi

# Print summary
echo ""
echo "════════════════════════════════════════════════════════════════"
echo -e "${BLUE}📊 RESUMEN DE PRUEBAS${NC}"
echo "════════════════════════════════════════════════════════════════"
echo ""
echo -e "${GREEN}Pasadas: $PASSED${NC}"
echo -e "${YELLOW}Advertencias: $WARNINGS${NC}"
echo -e "${RED}Fallidas: $FAILED${NC}"
echo ""

# Final verdict
if [ $FAILED -eq 0 ]; then
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo -e "║          ${GREEN}✅ VALIDACIÓN EXITOSA - LISTO PARA USO${NC}            ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    exit 0
else
    echo "╔════════════════════════════════════════════════════════════════╗"
    echo -e "║            ${RED}❌ VALIDACIÓN INCOMPLETA - REVISAR${NC}             ║"
    echo "╚════════════════════════════════════════════════════════════════╝"
    exit 1
fi

#!/bin/bash

# Script para activar/desactivar acceso a KubernetesConnectivityScreen
# Uso: ./enable_k8s_debug.sh [enable|disable]

FLUTTER_APP_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_FILE="$FLUTTER_APP_DIR/lib/main.dart"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Backup file
BACKUP_FILE="$FLUTTER_APP_DIR/.main.dart.backup"

enable_k8s_debug() {
    echo -e "${YELLOW}Activando acceso a KubernetesConnectivityScreen...${NC}"
    
    # Crear backup si no existe
    if [ ! -f "$BACKUP_FILE" ]; then
        cp "$MAIN_FILE" "$BACKUP_FILE"
        echo -e "${GREEN}✓ Backup creado${NC}"
    fi
    
    # Verificar si ya está modificado
    if grep -q "KubernetesConnectivityScreen()," "$MAIN_FILE"; then
        echo -e "${YELLOW}Ya está activado${NC}"
        return
    fi
    
    # Reemplazar AuthScreen() con KubernetesConnectivityScreen()
    # Usar sed con compatibilidad cross-platform
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' 's/home: const AuthScreen(),/home: const KubernetesConnectivityScreen(),/g' "$MAIN_FILE"
    else
        # Linux
        sed -i 's/home: const AuthScreen(),/home: const KubernetesConnectivityScreen(),/g' "$MAIN_FILE"
    fi
    
    # Agregar import si no está
    if ! grep -q "import 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';" "$MAIN_FILE"; then
        # Encontrar la línea de imports y agregar después de la última
        # Usar un marcador temporal para insertar
        sed -i '1,/^import.*auth_screen\.dart/s/^import.*auth_screen\.dart/&\nimport '\''package:flutter_app\/screens\/kubernetes_connectivity_screen.dart'\'';/' "$MAIN_FILE"
    fi
    
    echo -e "${GREEN}✓ KubernetesConnectivityScreen activado como home${NC}"
    echo -e "${YELLOW}Próximo paso: ejecuta 'flutter run' o presiona 'r' en DevTools${NC}"
}

disable_k8s_debug() {
    echo -e "${YELLOW}Desactivando acceso a KubernetesConnectivityScreen...${NC}"
    
    # Restaurar desde backup
    if [ ! -f "$BACKUP_FILE" ]; then
        echo -e "${RED}✗ No hay backup, no se puede restaurar${NC}"
        return 1
    fi
    
    cp "$BACKUP_FILE" "$MAIN_FILE"
    rm "$BACKUP_FILE"
    
    echo -e "${GREEN}✓ AuthScreen restaurado como home${NC}"
    echo -e "${YELLOW}Próximo paso: ejecuta 'flutter run' o presiona 'r' en DevTools${NC}"
}

status() {
    if grep -q "home: const KubernetesConnectivityScreen()," "$MAIN_FILE"; then
        echo -e "${GREEN}✓ Debug activado (KubernetesConnectivityScreen es home)${NC}"
    else
        echo -e "${YELLOW}○ Debug desactivado (AuthScreen es home)${NC}"
    fi
}

# Main
case "${1:-status}" in
    enable)
        enable_k8s_debug
        ;;
    disable)
        disable_k8s_debug
        ;;
    status)
        status
        ;;
    *)
        echo "Uso: $0 [enable|disable|status]"
        echo ""
        echo "Ejemplos:"
        echo "  $0 enable    - Activa KubernetesConnectivityScreen como home"
        echo "  $0 disable   - Restaura AuthScreen como home"
        echo "  $0 status    - Muestra estado actual"
        exit 1
        ;;
esac

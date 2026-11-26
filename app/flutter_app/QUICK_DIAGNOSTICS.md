# 🎯 Guía Rápida: Acceder a Diagnósticos de Kubernetes en 2 Minutos

## ⚡ Opción MÁS FÁCIL (Recomendada)

### Paso 1: Usar el script automatizado
```powershell
# En Windows (PowerShell)
cd app\flutter_app
.\enable_k8s_debug.ps1 -Action enable
```

```bash
# En Linux/macOS
cd app/flutter_app
chmod +x enable_k8s_debug.sh
./enable_k8s_debug.sh enable
```

**¿Qué hace?** Cambia automáticamente `main.dart` para que `KubernetesConnectivityScreen` sea la pantalla de inicio.

### Paso 2: Ejecutar la app
```bash
cd app/flutter_app
flutter run -d windows --debug
# o: flutter run -d android, flutter run -d linux, etc.
```

### Paso 3: Ver los diagnósticos
1. ✅ La app abre directamente en la pantalla de diagnósticos
2. ✅ Haz clic en el botón azul **"Full Diagnostics"**
3. ✅ Espera ~5 segundos a que termine
4. ✅ Lee el reporte

### Paso 4: Volver a lo normal
```powershell
# Windows
.\enable_k8s_debug.ps1 -Action disable
```

```bash
# Linux/macOS
./enable_k8s_debug.sh disable
```

---

## 📸 Screenshots de lo que verás:

```
┌─────────────────────────────────────────┐
│   🔵 Kubernetes Connectivity Diagnostics │
├─────────────────────────────────────────┤
│                                         │
│  Environment: Development               │
│  Base URL: http://localhost:8080        │
│                                         │
│  ┌─────────────────────────────────────┐│
│  │ ● Full Diagnostics                  ││
│  │ ● Test Health Endpoint               ││
│  │ ● Test WebSocket Connection          ││
│  │ ● Verify All Endpoints               ││
│  └─────────────────────────────────────┘│
│                                         │
│  Results:                              │
│  ✓ Health: OK                          │
│  ✓ Endpoints: 5/5 reachable            │
│  ✓ WebSocket: Connected                │
│  ✓ Network: Stable                     │
│  ✓ Response time: 234ms                │
│  ✓ Certificates: Valid                 │
│                                         │
└─────────────────────────────────────────┘
```

---

## 🔄 Flujo Completo en Terminal

### Terminal 1: Preparar
```bash
cd app/flutter_app

# Windows
.\enable_k8s_debug.ps1 -Action enable

# Linux/macOS
./enable_k8s_debug.sh enable
```

**Output esperado:**
```
✓ Backup creado
✓ KubernetesConnectivityScreen activado como home
Próximo paso: ejecuta 'flutter run' o presiona 'r' en DevTools
```

### Terminal 2: Ejecutar app
```bash
cd app/flutter_app
flutter run -d windows --debug
```

**Output esperado:**
```
Flutter run key commands.
h List all available commands.
c Clear the screen.
q Quit (terminate the app on exit).
s Save a screenshot to flutter.png.
z Run with unoptimized snapshots.
Z Run with VM service asserts.
p Toggle performance overlay.
I Display the widget inspector.
L Show the layers inspector.
t Trace widget rebuilds.
...
```

### Browser o DevTools: Ver la app
La app se abre automáticamente en la pantalla de diagnósticos.

---

## ✅ Checklist de Verificación

Después de ver el reporte, verifica esto:

```
Checklist:
□ ¿Muestra "✓" en Health?
□ ¿Muestra "5/5 reachable" en Endpoints?
□ ¿Muestra "Connected" en WebSocket?
□ ¿Muestra "Stable" en Network?
□ ¿Response time < 1000ms?
□ ¿Certificates: Valid?

Si todos los puntos tienen ✓:
   ✅ TAREA VERIFICADA - Integración funcionando correctamente

Si alguno tiene ✗:
   ⚠️  PROBLEMA - Revisa la sección de Troubleshooting abajo
```

---

## 🛠️ Troubleshooting Rápido

### "Network connection refused"
```bash
# Verificar que Kubernetes está corriendo
kubectl cluster-info

# Verificar que el API Gateway está corriendo
kubectl get pods

# Hacer port-forward para pruebas locales
kubectl port-forward svc/api-gateway 8080:8080
```

### "WebSocket timeout"
```
1. Verifica en lib/config/api_config.dart que websocketUrl sea correcto
2. Debe empezar con ws:// (o wss:// si tienes SSL)
3. Verifica que tu Ingress está configurado para WebSocket
```

### "Certificate validation failed"
```
Esto es NORMAL en desarrollo. 
En lib/services/http_client_manager.dart busca:
  // httpClient.badCertificateCallback = (_, __, ___) => true;

Descomenta esa línea SOLO en desarrollo. ⚠️ NUNCA en producción.
```

### "No app running" error
```bash
# Asegúrate de ejecutar con --debug
flutter run -d windows --debug

# NO ejecutes:
flutter build apk
```

---

## 🎓 Qué significa cada sección del reporte:

| Campo | Qué es | Esperado |
|-------|--------|----------|
| **Health** | Estado del backend | `✓ OK` |
| **Endpoints** | Servicios disponibles | `5/5 reachable` |
| **WebSocket** | Conexión en tiempo real | `Connected` |
| **Network** | Calidad de conexión | `Stable` |
| **Response Time** | Velocidad de respuesta | `< 500ms` |
| **Certificates** | Validez de SSL/TLS | `Valid` (OK en dev si desactivas validación) |

---

## 📱 Para Diferentes Plataformas:

### Windows (Desktop)
```bash
flutter run -d windows --debug
```

### Android (Emulador)
```bash
flutter run -d emulator-5554 --debug
# o:
flutter run -d android
```

### Linux (Desktop)
```bash
flutter run -d linux --debug
```

### iOS (Simulador en macOS)
```bash
flutter run -d iPhone --debug
# o:
flutter run -d ios
```

---

## 🎯 Resumen de Comandos

```bash
# Activar pantalla de diagnósticos
.\enable_k8s_debug.ps1 -Action enable          # Windows
./enable_k8s_debug.sh enable                   # Linux/macOS

# Ejecutar app
flutter run -d windows --debug

# Ver estado actual
.\enable_k8s_debug.ps1 -Action status          # Windows
./enable_k8s_debug.sh status                   # Linux/macOS

# Restaurar app normal
.\enable_k8s_debug.ps1 -Action disable         # Windows
./enable_k8s_debug.sh disable                  # Linux/macOS
```

---

## 🚀 Próximos Pasos

✅ **Si todo funciona:**
1. Desactiva el debug: `enable_k8s_debug.ps1 -Action disable`
2. Lanza a producción con confianza

⚠️ **Si hay problemas:**
1. Revisa el Troubleshooting arriba
2. Lee `README_FLUTTER_K8S.md` para más detalles
3. Consulta `VERIFICATION_REPORT.md` para configuraciones avanzadas

---

**¿Preguntas?** Lee [HOW_TO_RUN_DIAGNOSTICS.md](./HOW_TO_RUN_DIAGNOSTICS.md) para más detalles. 📖

# 🔧 Cómo Acceder a Diagnósticos de Kubernetes

## Opción 1: Desde DevTools (Recomendado para Desarrollo)

### Paso 1: Ejecutar la app en modo debug
```bash
cd app/flutter_app
flutter run -d windows  # o -d android, -d linux, etc.
```

### Paso 2: Abrir DevTools
En la terminal donde está ejecutándose flutter, verás una URL como:
```
The Dart DevTools debugger and profiler is available at: http://localhost:8888
```

Haz clic en esa URL o cópiala en tu navegador.

### Paso 3: Navegar a la pantalla de diagnósticos
Una vez en la app, puedes acceder de dos maneras:

**Opción A: Desde el código (debuging rápido)**
En DevTools → Console, ejecuta:
```dart
Navigator.of(context).push(
  MaterialPageRoute(
    builder: (_) => const KubernetesConnectivityScreen(),
  ),
);
```

**Opción B: Modificar temporalmente la app**
Edita `lib/main.dart` y cambia la línea de home:
```dart
home: const KubernetesConnectivityScreen(),  // en lugar de AuthScreen()
```

Luego ejecuta hot reload (presiona `r` en la terminal de Flutter).

---

## Opción 2: Integración Manual en la App

### Si quieres acceder desde el menú de la app:

#### Paso 1: Abrir archivo de menú/perfil
Busca donde está el menú principal (generalmente en ProfileScreen o un drawer):
```bash
# Buscar archivo que tenga opciones de menú
grep -r "menu\|debug\|settings" lib/screens lib/features
```

#### Paso 2: Agregar botón de diagnósticos
En el archivo del menú, agrega:

```dart
import 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';

// En tu lista de opciones del menú:
ListTile(
  leading: const Icon(Icons.bug_report),
  title: const Text('Kubernetes Diagnostics'),
  onTap: () {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => const KubernetesConnectivityScreen(),
      ),
    );
  },
),
```

#### Paso 3: Hot reload
En la terminal de Flutter, presiona `r` para hot reload.

---

## Opción 3: Crear una Ruta de Debug Rápida

### Paso 1: Editar `lib/main.dart`
Importa la pantalla:
```dart
import 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';
```

### Paso 2: Agregar ruta especial
En el `MaterialApp`, agrega `onGenerateRoute`:

```dart
MaterialApp(
  navigatorKey: _navigatorKey,
  title: 'Meloop',
  onGenerateRoute: (settings) {
    // Acceso rápido a diagnósticos con URL especial
    if (settings.name == '/k8s-diagnostics') {
      return MaterialPageRoute(
        builder: (_) => const KubernetesConnectivityScreen(),
      );
    }
    return null;
  },
  home: const AuthScreen(),
  // ... resto del código
);
```

### Paso 3: Usar la ruta en DevTools Console
```dart
Navigator.of(context).pushNamed('/k8s-diagnostics');
```

---

## 📊 Una vez en la pantalla de diagnósticos:

### Botón "Full Diagnostics"
Haz clic en este botón para ejecutar todas las pruebas:

✅ **Pruebas que ejecuta:**
1. Health check endpoint
2. API endpoints verification
3. WebSocket connection
4. Network connectivity
5. Response times
6. Certificate validation

### Resultados esperados:
```
✓ Health: OK
✓ Endpoints: 5/5 reachable
✓ WebSocket: Connected
✓ Network: Stable
✓ Response time: <500ms
✓ Certificates: Valid
```

### Si algo falla:
- Lee el mensaje de error específico
- Haz clic en "Test Health Endpoint" para aislar problemas
- Verifica tu configuración en `lib/config/api_config.dart`
- Consulta el troubleshooting abajo

---

## 🛠️ Troubleshooting

### "Network connection refused"
- Verifica que tu Kubernetes está corriendo
- Verifica la IP del cluster (comando: `kubectl cluster-info`)
- En `lib/config/api_config.dart`, actualiza la URL base

### "Certificate validation failed"
En modo debug, puedes desactivar verificación temporal:
1. En `lib/services/http_client_manager.dart`
2. Busca `badCertificateCallback`
3. Descomenta línea: `httpClient.badCertificateCallback = (_, __, ___) => true;`
4. ⚠️ NUNCA hagas esto en producción

### "WebSocket connection timeout"
- Verifica que tu LoadBalancer/Ingress esté configurado para WebSocket
- En `lib/config/api_config.dart`, verifica que `websocketUrl` es correcto
- Usa `wss://` si tienes SSL, `ws://` si no

### "No app running" en DevTools
Asegúrate de ejecutar:
```bash
flutter run -d windows --debug
```

---

## 🎯 Verificación paso a paso

### 1️⃣ Verificar que todo está corriendo
```bash
# Terminal 1: ejecutar la app
cd app/flutter_app
flutter run -d windows --debug

# Terminal 2: verificar Kubernetes
kubectl get pods
kubectl port-forward svc/api-gateway 8080:8080
```

### 2️⃣ Acceder a la pantalla
- Opción: Edita `lib/main.dart` y pon `KubernetesConnectivityScreen()` como home
- Hot reload: presiona `r`

### 3️⃣ Ejecutar diagnósticos
- Haz clic en "Full Diagnostics"
- Espera ~5 segundos
- Revisa los resultados

### 4️⃣ Interpretar resultados
```
✅ Verde = OK, puedes continuar
⚠️  Amarillo = Advertencia, revisar
❌ Rojo = Error, revisar configuración
```

---

## 📱 Para pruebas en diferentes plataformas:

### Windows
```bash
flutter run -d windows --debug
```

### Android (emulador)
```bash
flutter run -d emulator-5554 --debug
```

### Linux
```bash
flutter run -d linux --debug
```

### iOS (macOS)
```bash
flutter run -d iPhone --debug
```

---

## 🚀 Resumen rápido:

1. **Ejecutar app:** `flutter run -d windows --debug`
2. **Acceder pantalla:** Edita `main.dart` y pon `KubernetesConnectivityScreen()` como home
3. **Hot reload:** Presiona `r`
4. **Hacer diagnósticos:** Clic en "Full Diagnostics"
5. **Revisar resultados:** Espera a que termine y lee los reportes

¡Listo! 🎉

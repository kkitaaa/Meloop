# Verificación de Integración Flutter-Kubernetes

Este documento detalla cómo verificar que la integración de Flutter con el backend desplegado en Kubernetes funciona correctamente.

## 1. Verificación de Configuración

### 1.1 Verificar que ApiConfig está correctamente configurada

```dart
import 'package:flutter_app/config/api_config.dart';

void main() {
  print('Environment: ${ApiConfig.environment}');
  print('Base URL: ${ApiConfig.baseUrl}');
  print('WebSocket URL: ${ApiConfig.webSocketUrl}');
}
```

**Outputs esperados por entorno:**
- **Development**: `baseUrl: http://localhost:8080` (Windows) o `http://10.0.2.2:8080` (Android)
- **Staging**: `baseUrl: http://staging.meloop.com` (o URL del Ingress)
- **Production**: `baseUrl: https://api.meloop.com`

## 2. Compilación por Entorno

### 2.1 Desarrollo (Local)

```bash
# Windows
cd app\flutter_app
flutter run -d windows --dart-define=FLUTTER_ENV=development

# Android (Emulador)
flutter run -d emulator-5554 --dart-define=FLUTTER_ENV=development
```

### 2.2 Staging (K8s)

```bash
# Con Ingress hostname
flutter build apk --dart-define=FLUTTER_ENV=staging \
                   --dart-define=K8S_INGRESS_HOST=api.staging.meloop.com

# O con PowerShell
.\build.ps1 -Environment staging -Platform android
```

### 2.3 Producción

```bash
# Con URL de producción
flutter build apk --dart-define=FLUTTER_ENV=production \
                   --dart-define=K8S_API_GATEWAY_URL=https://api.meloop.com

# O con PowerShell
.\build.ps1 -Environment prod -Platform android
```

## 3. Testing de Conectividad

### 3.1 Pantalla de Kubernetes Connectivity Test

La app incluye una pantalla de debugging integrada. Accésible mediante:

```dart
// En tu main.dart o AppBar
Navigator.push(
  context,
  MaterialPageRoute(
    builder: (context) => const KubernetesConnectivityScreen(),
  ),
);
```

Esta pantalla proporciona:
- ✓ Estado actual de configuración
- ✓ Test de health check del backend
- ✓ Verificación de endpoints individuales
- ✓ Test de conexión WebSocket
- ✓ Reporte completo de diagnostico

### 3.2 Testing manual desde terminal

#### Verificar health del backend

```bash
# Local
curl -v http://localhost:8080/health

# K8s con port-forward
kubectl -n meloop port-forward svc/backend 8080:8080
curl -v http://localhost:8080/health

# A través de Ingress
curl -v -H "Host: meloop.local" http://<ingress-ip>/health
```

#### Verificar endpoints del API

```bash
# Demo endpoint
curl -v http://localhost:8080/api/demo

# Feed endpoint
curl -v -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/feed?category=music&page=0

# WebSocket
websocat ws://localhost:8080/api/notifications
```

## 4. Verificación desde Plataformas

### 4.1 Android

```bash
# Compilar APK para staging
flutter build apk --dart-define=FLUTTER_ENV=staging

# Instalar en dispositivo
adb install -r build/app/outputs/flutter-apk/app-release.apk

# Ver logs en tiempo real
adb logcat | grep flutter

# Verificar conectividad
adb shell dumpsys connectivitymanager
```

### 4.2 Windows

```bash
# Compilar para Windows
flutter build windows --dart-define=FLUTTER_ENV=development

# Ejecutar
start build/windows/runner/Release/flutter_app.exe

# Debug mode con hot reload
flutter run -d windows --dart-define=FLUTTER_ENV=development
```

### 4.3 iOS (si aplica)

```bash
# Compilar para staging
flutter build ios --dart-define=FLUTTER_ENV=staging --release

# Instalar en dispositivo
open -a Simulator
flutter run -d <device-id> --dart-define=FLUTTER_ENV=staging
```

## 5. Validación de Autenticación

### 5.1 Test de login

```dart
import 'package:flutter_app/services/api_client.dart';

final apiClient = ApiClient();

try {
  // Login
  final response = await apiClient.post(
    '/api/auth/login',
    {
      'username': 'testuser',
      'password': 'testpass',
    },
    (json) => json,
  );
  
  print('Login successful: ${response.data}');
  
  // Extractar token
  final token = response.data['token'];
  
  // Usar token en requests posteriores
  httpClientManager.setAuthToken(token);
  
} catch (e) {
  print('Login failed: $e');
}
```

### 5.2 Test de refresh de token

```dart
// Si el token expira, el interceptor debería refrescarlo automáticamente
// Verificar en logs que se ejecuta _TokenRefreshInterceptor
```

## 6. Verificación de WebSocket

### 6.1 Test manual con websocat

```bash
# Instalar websocat (si no lo tienes)
# En Windows con choco: choco install websocat
# En macOS: brew install websocat

# Conectarse al WebSocket
websocat ws://localhost:8080/api/notifications

# Deberías recibir mensajes de bienvenida
```

### 6.2 Test desde Flutter

```dart
import 'package:flutter_app/services/websocket_manager.dart';

final wsManager = WebSocketManager('/api/notifications');

wsManager.onMessage((message) {
  print('Received: $message');
});

wsManager.onError((error) {
  print('Error: $error');
});

wsManager.onClose(() {
  print('Connection closed');
});

await wsManager.connect();

// Enviar mensaje
wsManager.send('{"action": "subscribe", "channel": "notifications"}');

// Desconectar después de usar
await wsManager.disconnect();
```

## 7. Verificación de Kubernetes

### 7.1 Confirmar que el backend está disponible

```bash
# Ver pods
kubectl -n meloop get pods

# Ver servicios
kubectl -n meloop get svc

# Ver ingress
kubectl -n meloop get ingress

# Ver logs del API Gateway
kubectl -n meloop logs deployment/backend -c api-gateway -f

# Exec dentro del pod
kubectl -n meloop exec -it deployment/backend -- sh

# Desde dentro: curl localhost:8080/health
```

### 7.2 Test de DNS desde pod

```bash
# Si la app está en K8s, debería poder usar DNS interno
http://backend.meloop.svc.cluster.local:8080/health
```

## 8. Checklist de Validación

- [ ] ApiConfig retorna URL correcta para el entorno
- [ ] Health endpoint responde con 200
- [ ] Feed endpoint retorna datos con autenticación
- [ ] WebSocket se conecta y mantiene conexión
- [ ] Token refresh funciona automáticamente en 401
- [ ] Android puede conectarse (10.0.2.2:8080)
- [ ] Windows puede conectarse (localhost:8080)
- [ ] Reintentos automáticos funcionan en timeout
- [ ] Logs muestran información de debugging en desarrollo
- [ ] Endpoints individuales son accesibles según reporte

## 9. Troubleshooting

### "Cannot connect to backend"

```bash
# Verificar que el backend está corriendo
kubectl -n meloop get pods -l app.kubernetes.io/name=backend

# Verificar health endpoint
kubectl -n meloop port-forward svc/backend 8080:8080
curl http://localhost:8080/health

# Verificar networking
kubectl -n meloop logs deployment/backend
```

### "WebSocket connection failed"

```bash
# Verificar que WebSocket está habilitado
# Verificar que el puerto es accesible
netstat -an | grep 8080

# Si usa Ingress, verificar que soporta WebSocket
kubectl -n meloop get ingress backend -o yaml
```

### "Token refresh failing"

```dart
// Verificar que el token está siendo guardado
print(dio.options.headers['Authorization']);

// Verificar que el endpoint de refresh existe
curl -X POST http://localhost:8080/api/auth/refresh
```

### "Android no puede conectar (10.0.2.2)"

```bash
# Usar 10.0.2.2 para acceder a host desde Android emulador
# Verificar que la URL se configura correctamente
adb logcat | grep "10.0.2.2"
```

## 10. Performance Benchmarking

### Medir tiempo de respuesta

```dart
final stopwatch = Stopwatch()..start();
final response = await apiClient.getDemo();
stopwatch.stop();
print('Response time: ${stopwatch.elapsedMilliseconds}ms');
```

### Monitorear consumo de memoria

```bash
# Android
adb shell dumpsys meminfo | grep flutter_app

# Windows
tasklist | findstr flutter_app
# Luego usar Performance Monitor
```

## Próximos Pasos

1. ✅ Configuración por entorno
2. ✅ HTTP con retry y exponential backoff
3. ✅ WebSocket con reconexión automática
4. ✅ Health checks y diagnostico
5. ⏳ Integración en CI/CD pipeline
6. ⏳ Documentar endpoints específicos del negocio
7. ⏳ Implementar caching y synchronización offline
8. ⏳ Performance optimization

## Referencias

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Flutter Networking](https://flutter.dev/docs/cookbook/networking)
- [Dio HTTP Client](https://pub.dev/packages/dio)
- [WebSocket API](https://dart.dev/guides/libraries/library-tour#dartio---io-library-for-server-side-code)

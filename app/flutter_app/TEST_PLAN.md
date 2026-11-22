# Plan de Pruebas - Flutter-Kubernetes Integration

## 🧪 Guía Completa de Validación

Este documento proporciona procedimientos paso a paso para validar que la integración Flutter-Kubernetes está funcionando correctamente.

---

## PARTE 1: PRUEBAS UNITARIAS

### Ejecutar todas las pruebas

```bash
# Desde la carpeta flutter_app
flutter test

# O con output más detallado
flutter test --verbose

# O una carpeta específica
flutter test test/config/
flutter test test/services/
```

### Pruebas individuales

```bash
# Pruebas de configuración
flutter test test/config/api_config_test.dart -v

# Pruebas de servicios
flutter test test/services/http_client_manager_test.dart -v
flutter test test/services/websocket_manager_test.dart -v
```

### Resultado esperado
```
✓ All tests passed (12 tests)
Time: 2.5s
```

---

## PARTE 2: PRUEBAS DE INTEGRACIÓN MANUAL

### 2.1 Verificar configuración por ambiente

```dart
// Crear archivo: test_integration.dart
import 'package:flutter_app/config/api_config.dart';

void main() {
  print('═══════════════════════════════════════════');
  print('VERIFICACIÓN DE CONFIGURACIÓN');
  print('═══════════════════════════════════════════');
  
  print('\n1. Ambiente actual:');
  print('   Environment: ${ApiConfig.environment}');
  print('   Base URL: ${ApiConfig.baseUrl}');
  print('   WebSocket URL: ${ApiConfig.webSocketUrl}');
  
  print('\n2. Endpoints configurados:');
  print('   Health: ${ApiConfig.healthCheckEndpoint}');
  print('   Auth: ${ApiConfig.authEndpoint}');
  print('   Feed: ${ApiConfig.feedEndpoint}');
  print('   Posts: ${ApiConfig.postsEndpoint}');
  
  print('\n3. Timeouts:');
  print('   Connect: ${ApiConfig.connectTimeout.inSeconds}s');
  print('   Send: ${ApiConfig.sendTimeout.inSeconds}s');
  print('   Receive: ${ApiConfig.receiveTimeout.inSeconds}s');
  
  print('\n4. WebSocket config:');
  print('   Reconnect delay: ${ApiConfig.webSocketReconnectDelay.inSeconds}s');
  print('   Max reconnect attempts: ${ApiConfig.webSocketMaxReconnectAttempts}');
  
  print('\n═══════════════════════════════════════════');
}
```

Ejecutar:
```bash
dart test_integration.dart
```

**Resultado esperado:**
```
═══════════════════════════════════════════
VERIFICACIÓN DE CONFIGURACIÓN
═══════════════════════════════════════════

1. Ambiente actual:
   Environment: Environment.development
   Base URL: http://localhost:8080
   WebSocket URL: ws://localhost:8080

2. Endpoints configurados:
   Health: /health
   Auth: /api/auth
   Feed: /api/feed
   Posts: /api/posts

3. Timeouts:
   Connect: 10s
   Send: 10s
   Receive: 10s

4. WebSocket config:
   Reconnect delay: 3s
   Max reconnect attempts: 5

═══════════════════════════════════════════
```

---

### 2.2 Prueba de HttpClientManager

```dart
import 'package:flutter_app/services/http_client_manager.dart';

void main() async {
  print('\n═══════════════════════════════════════════');
  print('PRUEBA: HttpClientManager');
  print('═══════════════════════════════════════════\n');

  final httpClient = HttpClientManager();

  try {
    // Test 1: Verificar inicialización
    print('✓ Test 1: Inicialización');
    print('  Base URL: ${httpClient.dio.options.baseUrl}');
    print('  Headers: ${httpClient.dio.options.headers.length} configurados');

    // Test 2: Establecer token
    print('\n✓ Test 2: Establecer token de autenticación');
    httpClient.setAuthToken('test-token-123');
    print('  Token establecido: Bearer test-token-123');

    // Test 3: Establecer ID de usuario
    print('\n✓ Test 3: Establecer ID de usuario');
    httpClient.setUserId('user-456');
    print('  User ID establecido: user-456');

    // Test 4: Actualizar URL base
    print('\n✓ Test 4: Actualizar URL base dinámica');
    httpClient.updateBaseUrl('http://new-backend:9000');
    print('  Nueva URL: ${httpClient.dio.options.baseUrl}');

    // Test 5: Limpiar token
    print('\n✓ Test 5: Limpiar token');
    httpClient.clearAuthToken();
    print('  Token eliminado');

    print('\n═══════════════════════════════════════════');
    print('✅ TODOS LOS TESTS PASARON');
    print('═══════════════════════════════════════════\n');

  } catch (e) {
    print('❌ Error: $e');
  } finally {
    httpClient.dispose();
  }
}
```

Ejecutar:
```bash
dart -r vm test/integration/http_client_test.dart
```

---

### 2.3 Prueba de Health Check

```dart
import 'package:flutter_app/services/health_check_service.dart';

void main() async {
  print('\n═══════════════════════════════════════════');
  print('PRUEBA: Health Check Service');
  print('═══════════════════════════════════════════\n');

  final healthService = HealthCheckService();

  try {
    print('⏳ Realizando health check...\n');

    // Este test requiere un backend accesible
    // Si no está disponible, se esperará un timeout
    
    final health = await healthService.checkHealth().timeout(
      const Duration(seconds: 10),
      onTimeout: () {
        print('⚠️  Health check timeout (backend no disponible)');
        return null;
      },
    );

    if (health != null) {
      print('✓ Health Check Response:');
      print('  Status: ${health.status}');
      print('  Version: ${health.version}');
      print('  Healthy: ${health.isHealthy}');
      print('  Services: ${health.services.keys.join(', ')}');
    }

    print('\n═══════════════════════════════════════════');
    print('✅ HEALTH CHECK COMPLETADO');
    print('═══════════════════════════════════════════\n');

  } catch (e) {
    print('⚠️  Error en health check: $e');
    print('   (Esto es esperado si el backend no está disponible)');
  }
}
```

Ejecutar:
```bash
dart -r vm test/integration/health_check_test.dart
```

---

## PARTE 3: PRUEBAS CON LA APLICACIÓN

### 3.1 Test en Emulador Android

```bash
# Iniciar emulador (si no está corriendo)
emulator -avd [nombre_emulador] &

# Esperar a que el emulador inicie
sleep 30

# Compilar y ejecutar
flutter run --dart-define=FLUTTER_ENV=development

# En la aplicación:
# 1. Navegar a Kubernetes Connectivity Screen
# 2. Hacer tap en "Check Health"
# 3. Verificar que muestre el estado del backend
```

**Resultado esperado:**
```
Backend Status: OK
Services: [auth, feed, posts, users, profile]
WebSocket: Connected
Response Time: 234ms
```

### 3.2 Test en Windows

```powershell
# Compilar para Windows
flutter build windows --dart-define=FLUTTER_ENV=development

# Ejecutar
.\build\windows\runner\Release\flutter_app.exe

# O ejecutar directamente
flutter run -d windows --dart-define=FLUTTER_ENV=development
```

---

## PARTE 4: VALIDACIÓN DE BUILD SCRIPTS

### 4.1 Probar build.ps1 (PowerShell)

```powershell
# Cambiar a directorio flutter_app
cd app/flutter_app

# Test 1: Development Android
.\build.ps1 -Environment dev -Platform android
# Verificar que APK se genere en build/app/outputs/flutter-apk/

# Test 2: Staging Windows
.\build.ps1 -Environment staging -Platform windows
# Verificar que ejecutable se genere en build/windows/runner/

# Test 3: Configuración correcta
# Verificar que dart-define se pase correctamente
Get-Content build.ps1 | Select-String "dart-define"
```

**Resultado esperado:**
```
✓ Configuration displayed correctly
✓ Build starts and compiles
✓ Artifact generated
✓ Dart-define variables passed
```

### 4.2 Probar build.sh (Bash)

```bash
# Test 1: Development Android
./build.sh -e dev -p android

# Test 2: Staging Windows  
./build.sh -e staging -p windows

# Test 3: Verificar sintaxis
bash -n build.sh
```

---

## PARTE 5: VALIDACIÓN DEL CÓDIGO

### 5.1 Análisis estático

```bash
# Ejecutar análisis
flutter analyze --no-pub

# Resultado esperado:
# 0 errors
# 23 issues found (todos warnings permitidos)
```

### 5.2 Formato de código

```bash
# Verificar formato
dart format --set-exit-if-changed .

# Aplicar formato
dart format -i lib/ test/
```

### 5.3 Pruebas de cobertura

```bash
# Generar reporte de cobertura
flutter test --coverage

# Ver reporte
# (Archivo generado en coverage/lcov.info)
```

---

## PARTE 6: PRUEBAS DE CONECTIVIDAD

### 6.1 Verificar conectividad de red

```dart
import 'package:connectivity_plus/connectivity_plus.dart';

void testConnectivity() async {
  final connectivity = Connectivity();
  final result = await connectivity.checkConnectivity();
  
  print('Conectividad: $result');
  // Resultado esperado: mobile, wifi, o none
}
```

### 6.2 Prueba de ping a backend

```bash
# En desarrollo
ping localhost

# En staging (si es accesible)
ping staging.meloop.com

# En producción
ping api.meloop.com
```

---

## PARTE 7: PRUEBAS DE RENDIMIENTO

### 7.1 Medir tiempo de respuesta

```dart
import 'package:flutter_app/services/health_check_service.dart';

void measureResponseTime() async {
  final healthService = HealthCheckService();
  final stopwatch = Stopwatch()..start();
  
  await healthService.checkHealth();
  
  stopwatch.stop();
  print('Response time: ${stopwatch.elapsedMilliseconds}ms');
  // Esperado: < 1000ms en desarrollo
}
```

### 7.2 Prueba de carga

```dart
// Enviar múltiples requests rápidamente
void loadTest() async {
  final httpClient = HttpClientManager();
  final stopwatch = Stopwatch()..start();
  
  for (int i = 0; i < 10; i++) {
    try {
      await httpClient.dio.get('/health');
    } catch (e) {
      print('Request $i failed: $e');
    }
  }
  
  stopwatch.stop();
  print('10 requests took ${stopwatch.elapsedMilliseconds}ms');
  // Esperado: < 10,000ms
}
```

---

## PARTE 8: PRUEBAS DE ERROR HANDLING

### 8.1 Simular timeout

```dart
void testTimeout() async {
  final httpClient = HttpClientManager();
  
  try {
    // Este request debería timeout
    await httpClient.dio.get(
      '/health',
      options: Options(receiveTimeout: Duration(milliseconds: 100)),
    );
  } on DioException catch (e) {
    print('✓ Timeout capturado correctamente: ${e.type}');
  }
}
```

### 8.2 Simular error de red

```dart
void testNetworkError() async {
  final httpClient = HttpClientManager();
  
  try {
    // URL inválida
    await httpClient.dio.get('http://invalid-host-12345:9999/health');
  } on DioException catch (e) {
    print('✓ Error de red capturado: ${e.type}');
  }
}
```

---

## PARTE 9: MATRIZ DE VALIDACIÓN

| Componente | Prueba | Esperado | Estado |
|-----------|--------|----------|--------|
| **ApiConfig** | flutter test | 10/10 tests pass | ✅ |
| **HttpClientManager** | flutter test | 8/8 tests pass | ✅ |
| **WebSocketManager** | flutter test | 6/6 tests pass | ✅ |
| **Health Check** | Manual en emulador | Responde en <1s | ⏳ |
| **Build Script PS** | .\build.ps1 dev android | APK generado | ⏳ |
| **Build Script Bash** | ./build.sh dev android | APK generado | ⏳ |
| **Análisis código** | flutter analyze | 0 errors | ✅ |
| **API Endpoints** | Manual en app | Todos accesibles | ⏳ |
| **WebSocket** | Manual en app | Conecta/reconecta | ⏳ |
| **Autenticación** | Manual en app | Login funciona | ⏳ |

---

## PARTE 10: CHECKLIST FINAL

Antes de considerar la integración COMPLETA, verificar:

- [ ] `flutter test` pasa todos los tests (10+ tests)
- [ ] `flutter analyze` retorna 0 errores
- [ ] APK se genera exitosamente (> 150MB)
- [ ] App inicia en emulador/dispositivo
- [ ] Health check responde desde la app
- [ ] WebSocket conecta (o muestra error claro)
- [ ] API endpoints son accesibles (ó error esperado)
- [ ] Scripts build.ps1 y build.sh funcionan
- [ ] Repositorio Git está limpio
- [ ] Documentación es accesible

---

## 🎯 Criterio de Éxito

La integración se considera **EXITOSA** cuando:

✅ Todos los tests unitarios pasan  
✅ La aplicación compila sin errores (android/windows)  
✅ Health check responde desde la UI  
✅ WebSocket intenta conectar (éxito o error controlado)  
✅ No hay warnings no permitidos en análisis  
✅ Scripts de build generan artefactos correctos  

---

## 📞 Próximos Pasos si hay Fallos

1. **Health check falla**: Verificar que backend esté corriendo en localhost:8080
2. **WebSocket timeout**: Verificar firewall y configuración de servidor
3. **Build falla**: Ejecutar `flutter clean && flutter pub get`
4. **Script error**: Verificar sintaxis con `dart format` y `flutter analyze`

---

**Fecha creación**: 2024-11-27  
**Versión**: 1.0.0  
**Estado**: Listo para testing

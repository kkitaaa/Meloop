# Configuración de Flutter - Integración con Kubernetes

## Descripción General

Esta configuración permite a la aplicación Flutter comunicarse con los microservicios de Meloop desplegados en Kubernetes. La aplicación soporta múltiples entornos: desarrollo, staging y producción.

## Estructura de Configuración

### ApiConfig (lib/config/api_config.dart)

Clase centralizada que gestiona toda la configuración de API:

- **Entornos soportados**: 
  - `development`: Conexión local (localhost o 10.0.2.2 para Android)
  - `staging`: Conexión a cluster K8s
  - `production`: Conexión a servidor de producción

- **Variables de entorno**:
  - `FLUTTER_ENV`: Define el entorno (`development`, `staging`, `production`)
  - `API_BASE_URL`: URL base personalizada del API
  - `K8S_API_GATEWAY_URL`: URL del API Gateway en K8s
  - `K8S_NAMESPACE`: Namespace de K8s (default: `meloop`)
  - `K8S_INGRESS_HOST`: Hostname del Ingress de K8s (default: `meloop.local`)
  - `CURRENT_USER_ID`: ID del usuario actual para headers

### HttpClientManager (lib/services/http_client_manager.dart)

Gestor de cliente HTTP con:
- **Retry automático** con exponential backoff para errores de conexión
- **Logging** en desarrollo
- **Gestión de tokens** de autenticación
- **Refresh automático** de tokens (401 Unauthorized)

### WebSocketManager (lib/services/websocket_manager.dart)

Gestor de conexiones WebSocket con:
- **Reconexión automática** con backoff exponencial
- **Manejo de errores** robusto
- **Handlers** para mensajes, errores y cierre de conexión
- Soporte para múltiples listeners

### HealthCheckService (lib/services/health_check_service.dart)

Servicio para verificar la conectividad del backend:
- Verifica health del backend
- Valida accesibilidad de endpoints individuales
- Prueba conexiones WebSocket
- Genera reportes de diagnostico

## Compilación por Entorno

### Desarrollo (Local)

```bash
cd app/flutter_app

# Android - Emulador (usa 10.0.2.2 como host)
flutter run -d emulator-5554

# Windows - Local (usa localhost:8080)
flutter run -d windows

# Con configuración personalizada
flutter run --dart-define=FLUTTER_ENV=development \
            --dart-define=API_BASE_URL=http://localhost:8080
```

### Staging (Kubernetes Cluster)

```bash
# En cluster K8s con Ingress
flutter build apk --dart-define=FLUTTER_ENV=staging \
                   --dart-define=K8S_INGRESS_HOST=api.staging.meloop.com \
                   --dart-define=K8S_NAMESPACE=meloop

# O con URL de API Gateway explícita
flutter build apk --dart-define=FLUTTER_ENV=staging \
                   --dart-define=K8S_API_GATEWAY_URL=http://staging-api.meloop.com
```

### Producción

```bash
# Producción
flutter build apk --dart-define=FLUTTER_ENV=production \
                   --dart-define=K8S_API_GATEWAY_URL=https://api.meloop.com

# Para iOS/Android release
flutter build ios --dart-define=FLUTTER_ENV=production \
                   --dart-define=K8S_API_GATEWAY_URL=https://api.meloop.com \
                   --release
```

## Configuración de Kubernetes

### URL de acceso según tipo de ingreso:

1. **Desde dentro del cluster** (ej: otros pods):
   ```
   http://backend.meloop.svc.cluster.local:8080
   ```

2. **A través de Ingress** (desde fuera):
   ```
   http://meloop.local/   (requiere /etc/hosts)
   https://api.meloop.com (producción con DNS)
   ```

3. **A través de LoadBalancer** (si está configurado):
   ```
   http://<EXTERNAL-IP>:8080
   ```

### Verificar endpoints de K8s

```bash
# Ver servicios disponibles
kubectl -n meloop get services

# Ver ingress
kubectl -n meloop get ingress

# Verificar health del backend
kubectl -n meloop exec -it deployment/backend -- curl localhost:8080/health

# Ver logs del API Gateway
kubectl -n meloop logs deployment/backend -c api-gateway -f
```

## Validación de Conectividad

### Usando HealthCheckService

```dart
final healthCheck = HealthCheckService();
final report = await healthCheck.runDiagnostics();

print(report);
print('Backend Healthy: ${report.backendHealthy}');
print('All Endpoints: ${report.endpoints}');
print('WebSocket Available: ${report.webSocketHealthy}');
```

### Verificación manual desde Flutter

```dart
import 'package:flutter_app/services/health_check_service.dart';

// En tu app
final healthCheck = HealthCheckService();

// Chequear health
try {
  final health = await healthCheck.checkHealth();
  print('Backend is healthy: ${health.isHealthy}');
  print('Version: ${health.version}');
} catch (e) {
  print('Backend is not accessible: $e');
}

// Chequear endpoints específicos
final endpointOk = await healthCheck.checkEndpoint('/api/feed');
print('Feed endpoint accessible: $endpointOk');
```

## Flujo de Autenticación

1. **Login**: 
   - POST `/api/auth/login` con credenciales
   - Recibe JWT token
   - HttpClientManager almacena token en headers

2. **Requests autenticados**:
   - Todo request incluye `Authorization: Bearer <token>`

3. **Token expirado** (401):
   - _TokenRefreshInterceptor detecta 401
   - Llama a `/api/auth/refresh`
   - Reintentar request original

4. **Logout**:
   - Eliminar token local
   - POST `/api/auth/logout` (opcional)

## Troubleshooting

### "Cannot connect to backend"

1. Verificar que el backend está corriendo:
   ```bash
   kubectl -n meloop get pods -l app.kubernetes.io/name=backend
   ```

2. Verificar health endpoint:
   ```bash
   curl http://localhost:8080/health
   ```

3. Para Android: Verificar que la URL es `10.0.2.2:8080` (no `localhost`)

4. Para K8s: Verificar hosts file si usa `meloop.local`:
   ```
   # /etc/hosts
   127.0.0.1  meloop.local
   ```

### "WebSocket connection failed"

1. Verificar que WebSocket está habilitado en backend
2. Verificar firewall permite WebSocket (puerto 8080)
3. Revisar logs del backend

### "Timeout connecting to backend"

1. Verificar latencia de red: `ping backend-url`
2. Aumentar timeout en ApiConfig si necesario
3. Verificar que el backend no está sobrecargado

## Testing

### Unit Tests

```bash
flutter test test/

# Test específico
flutter test test/services/health_check_service_test.dart
```

### Integration Tests

```bash
flutter test integration_test/

# Desde emulador
flutter drive --target=integration_test/api_gateway_test.dart
```

### Verificación manual desde línea de comandos

```bash
# Verificar health
curl -v http://localhost:8080/health

# Verificar con headers
curl -v -H "Authorization: Bearer token" \
     -H "X-User-ID: user123" \
     http://localhost:8080/api/demo

# WebSocket test
websocat ws://localhost:8080/api/notifications
```

## Variables de Entorno en CI/CD

Para GitHub Actions o similar:

```yaml
env:
  FLUTTER_ENV: staging
  K8S_API_GATEWAY_URL: ${{ secrets.K8S_API_GATEWAY_URL }}
  K8S_NAMESPACE: meloop

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: subosito/flutter-action@v2
      - name: Build APK
        run: |
          cd app/flutter_app
          flutter build apk --dart-define=FLUTTER_ENV=${{ env.FLUTTER_ENV }} \
                            --dart-define=K8S_API_GATEWAY_URL=${{ env.K8S_API_GATEWAY_URL }}
```

## Próximos Pasos

1. ✅ Configuración de entornos (realizado)
2. ✅ Gestión HTTP con retry (realizado)
3. ✅ WebSocket con reconexión (realizado)
4. ✅ Health checks (realizado)
5. ⏳ Implementar verificación de autenticación
6. ⏳ Validar desde plataformas (Android, Windows)
7. ⏳ Integrar en CI/CD pipeline
8. ⏳ Documentar endpoints específicos del negocio

# Referencia rápida

## URLs

| Entorno | Windows | Android Emulator |
|---|---|---|
| Desarrollo | `http://localhost:8080` | `http://10.0.2.2:8080` |

Ambas URLs requieren `kubectl port-forward svc/backend 8080:8080 -n meloop`.
Un teléfono físico requiere una URL accesible en la red local.

Staging/production deben configurar un endpoint real:

```powershell
flutter run -d windows `
  --dart-define=FLUTTER_ENV=staging `
  --dart-define=API_BASE_URL=https://<gateway-staging-real>
```

Production exige HTTPS. La app no inventa dominios de entorno.

## Rutas principales

- Salud: `GET /health`
- Autenticación: `/auth/login`, `/auth/register`, `/auth/refresh`,
  `/auth/validate`, `/auth/logout`
- Prueba Gateway: `/api/demo`
- Feed: `/api/feed`
- Recomendaciones: `/recommendations/...` cuando el servicio está desplegado

WebSocket no está desplegado en la topología actual. Solo configúralo mediante
`WEBSOCKET_PATH` cuando exista una ruta real.

## Comandos

Desde `app\flutter_app`:

```powershell
flutter test
flutter analyze
flutter build apk --debug --dart-define=FLUTTER_ENV=development
flutter build windows --debug --dart-define=FLUTTER_ENV=development
flutter test integration_test -d windows --dart-define=FLUTTER_ENV=development
```

La prueba de integración necesita el API Gateway activo. Para instrucciones
completas de Kubernetes y Android Emulator, consulta
[KUBERNETES_INTEGRATION.md](./KUBERNETES_INTEGRATION.md) y
[VERIFICATION_GUIDE.md](./VERIFICATION_GUIDE.md).

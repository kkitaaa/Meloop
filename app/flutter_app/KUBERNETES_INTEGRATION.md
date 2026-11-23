# Configuración Flutter para Kubernetes

La app Flutter se conecta al API Gateway; no usa DNS interno de Kubernetes
porque Android y Windows ejecutan fuera del clúster.

## Desarrollo local

Desde la raíz del repositorio, mantén este comando activo:

```powershell
kubectl port-forward svc/backend 8080:8080 -n meloop
```

Inicia la app desde `app\flutter_app`:

```powershell
flutter run -d windows --dart-define=FLUTTER_ENV=development
flutter run -d <id-del-emulador-android> --dart-define=FLUTTER_ENV=development
```

Flutter selecciona `localhost:8080` en Windows y `10.0.2.2:8080` en Android
Emulator. Para un dispositivo Android físico, configura `API_BASE_URL` con una
dirección alcanzable desde ese dispositivo; `localhost` y `10.0.2.2` no son
direcciones para el host de un teléfono físico.

## Staging

Para probar el Gateway de staging desde Windows:

```powershell
cd app\flutter_app
flutter run -d windows `
  --dart-define=FLUTTER_ENV=staging `
  --dart-define=API_BASE_URL=https://meloop-acofre.dev.censei.cl
```

En Android Emulator usa la misma URL HTTPS. `API_BASE_URL` debe ser solo el
origen del Gateway; no agregues `/api` al final. El cliente define la ruta
completa de cada servicio.

El Gateway enruta autenticación por `/auth`, usuarios por `/users`, amistades
por `/friends` y gamificación por `/api/v1/gamification`. Otras rutas de la app,
como `/api/demo`, `/api/feed` y `/v1/social/users/.../mutual-friends`, conservan
sus prefijos propios.

## Staging y producción

Pasa la URL real del Ingress o LoadBalancer del API Gateway:

```powershell
flutter build apk --release `
  --dart-define=FLUTTER_ENV=staging `
  --dart-define=API_BASE_URL=https://<gateway-staging-real>
```

Production también requiere `https://`. No se incluyen nombres DNS ficticios
como valores predeterminados. Los scripts `build.ps1` y `build.sh` exigen
configuración explícita para staging/production y aceptan sus variables de
entorno específicas; consulta el propio script para la lista completa.

## Rutas y capacidades

| Uso | Ruta del Gateway | Estado |
|---|---|---|
| Salud | `GET /health` | Configurada |
| Login / registro | `POST /auth/login`, `POST /auth/register` | Requiere auth-service, Redis y base de datos |
| Renovación / logout / validar | `/auth/refresh`, `/auth/logout`, `/auth/validate` | Requiere sesión válida |
| Prueba HTTP | `GET, POST /api/demo` | Gateway |
| Feed | `GET /api/feed` | Feed de demostración del Gateway |
| Recomendaciones | `/recommendations/...` | Depende de recommendation-service y sus dependencias |
| WebSocket | Ruta opcional `WEBSOCKET_PATH` | Sin servidor WebSocket en los manifests actuales |

El token es una sesión de servidor, no un refresh-token OAuth. El cliente
solicita extensión de su TTL antes del vencimiento. Si ya expiró o fue revocada,
el usuario debe volver a iniciar sesión.

La lista completa de recursos, pre-requisitos y operaciones del clúster está en
[k8s/README.md](../../k8s/README.md).

# Plan de pruebas Flutter-Kubernetes

## Pruebas locales (sin clúster)

Desde `app\flutter_app`:

```powershell
flutter test
flutter analyze
flutter build apk --debug --dart-define=FLUTTER_ENV=development
flutter build windows --debug --dart-define=FLUTTER_ENV=development
```

Estas pruebas verifican código, compilación y contratos mockeados; no prueban
DNS, Ingress, secretos, base de datos ni servicios de un clúster.

## Pruebas contra el Gateway

1. Desplegar siguiendo [k8s/README.md](../../k8s/README.md).
2. Iniciar `kubectl port-forward svc/backend 8080:8080 -n meloop`.
3. Comprobar `curl.exe http://localhost:8080/health` y `/api/demo`.
4. Ejecutar `flutter test integration_test -d windows`.
5. Ejecutar la misma prueba con `-d <id-del-emulador-android>`.
6. Iniciar sesión con una cuenta de prueba desde ambas plataformas.
7. Comprobar persistencia, logout y, cuando aplique, renovación antes de
   vencimiento.
8. Revisar la pantalla **Diagnóstico de conexión** en debug.

La integración test automatiza GET/POST de `/api/demo` y lectura del feed. Para
la cuenta real y renovación de sesión se necesita un entorno de auth operativo.

## Matriz de aceptación

| Criterio | Evidencia necesaria |
|---|---|
| URL correcta por ambiente | Configuración explícita de Gateway e inspección de la URL seleccionada |
| HTTP por Gateway | Test de integración en vivo desde Windows y Android |
| Login/renovación/logout | Cuenta de prueba en auth-service con PostgreSQL y Redis |
| Funcionalidades principales | Servicios correspondientes listos y flujo probado |
| WebSocket | Servidor/ruta desplegados y prueba en vivo; actualmente no configurado |

No se declara la tarea plenamente verificada hasta ejecutar las comprobaciones
en un clúster real con su configuración y secretos.

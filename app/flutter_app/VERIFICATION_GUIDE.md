# Verificación Flutter ↔ API Gateway

Las pruebas unitarias no necesitan un clúster. Las pruebas de integración y el
inicio de sesión real sí requieren un Gateway desplegado y accesible.

## 1. Preparar el entorno local

Sigue primero [k8s/README.md](../../k8s/README.md). Con el clúster listo, deja
activo el port-forward en una terminal:

```powershell
kubectl port-forward svc/backend 8080:8080 -n meloop
```

En otra terminal verifica:

```powershell
curl.exe http://localhost:8080/health
curl.exe http://localhost:8080/api/demo
```

La autenticación requiere además que `auth-service` esté listo y tenga acceso
al Secret `meloop-secrets` con `DATABASE_URL`, a PostgreSQL/Supabase y a Redis.

## 2. Ejecutar validaciones automatizadas

Desde `app\flutter_app`:

```powershell
flutter test
flutter analyze
flutter build apk --debug --dart-define=FLUTTER_ENV=development
flutter build windows --debug --dart-define=FLUTTER_ENV=development
```

Las pruebas que hacen llamadas al clúster se ejecutan por separado:

```powershell
flutter devices
flutter test integration_test -d windows --dart-define=FLUTTER_ENV=development
flutter test integration_test -d <id-del-emulador-android> --dart-define=FLUTTER_ENV=development
```

`integration_test/api_gateway_test.dart` verifica las operaciones GET/POST de
`/api/demo` y la carga del feed. Ejecútala desde Windows y Android Emulator para
probar ambas rutas de red; un test compilado o una prueba unitaria no sustituye
esta comprobación en vivo.

## 3. Probar autenticación en la app

1. Inicia Flutter en Windows o Android Emulator con el port-forward activo.
2. Usa la pantalla de login para autenticar una cuenta de prueba real.
3. Confirma que las credenciales inválidas muestran el error del servidor y no
   abren la app autenticada.
4. Cierra y vuelve a abrir la app para comprobar restauración de sesión.
5. Cierra sesión y comprueba que la sesión ya no se restaura.

La renovación de sesión ocurre antes de que expire una sesión válida. No
pruebes contra cuentas productivas ni compartas sus contraseñas en los comandos.

## 4. Diagnósticos desde la interfaz

En una compilación debug, abre **Diagnóstico de conexión** desde el login. La
pantalla realiza comprobaciones HTTP de health, demo, feed, recomendaciones y
enrutamiento de auth. Un WebSocket sin `WEBSOCKET_PATH` debe aparecer como no
configurado; no significa que la conexión se haya probado.

## 5. Criterios que requieren un despliegue real

- Login, registro, validación, renovación y logout a través del Ingress/Gateway.
- Rutas de las funcionalidades principales hacia los microservicios desplegados.
- Integración HTTP desde Windows y Android Emulator (y dispositivo físico, si
  forma parte del soporte esperado).
- WebSocket solo si se despliega/configura un servidor y una ruta real.

Si no hay kubeconfig, Ingress o credenciales de despliegue disponibles, registra
esas comprobaciones como pendientes; no marques el criterio como cumplido por
el hecho de que Flutter compile.

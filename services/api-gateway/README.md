# API Gateway

El API Gateway es el punto de entrada HTTP para la aplicación Flutter. Escucha
en el puerto definido por `PORT` (por defecto, `8080`) y reenvía las
solicitudes al microservicio correspondiente.

## Rutas

| Acceso | Gateway | Destino | Variable de entorno |
| --- | --- | --- | --- |
| Pública | `GET /health` | Respuesta local del gateway | Ninguna |
| Pública | `GET /api/demo` | Respuesta JSON de prueba del gateway | Ninguna |
| Pública | `POST /api/demo` | Valida y devuelve el mensaje recibido | Ninguna |
| Pública | `GET /api/feed` | Feed paginado de demostración por categoría | Ninguna |
| Pública | `ANY /test` y `ANY /test/*any` | `test-service`, quitando `/test` | `TEST_SERVICE_URL` |
| Privada | `ANY /users` y `ANY /users/*any` | `user-service`, conservando `/users` | `USER_SERVICE_URL` |
| Pública | `ANY /auth` y `ANY /auth/*any` | `auth-service`, conservando `/auth` | `AUTH_SERVICE_URL` |

Las rutas privadas reciben y reenvían el encabezado `Authorization`; la
validación del token debe realizarse mediante la política de autenticación del
servicio correspondiente.

Las variables de destino aceptan una URL completa. Si no están definidas, el
gateway usa los valores locales `http://localhost:8081`,
`http://localhost:8082` y `http://localhost:8083`, respectivamente.

## Verificación local

Con los servicios levantados, la comunicación con `user-service` puede
verificarse con:

```powershell
Invoke-WebRequest -Uri http://localhost:8080/users -Method Get
```

En Docker Compose, el gateway usa los nombres de servicio de la red interna.
La prueba automatizada del proxy cubre reenvío de método, ruta, query y body,
además del error `502 Bad Gateway` cuando el destino no está disponible.

## Prueba de integración Flutter

Inicia el gateway local con `go run .` desde este directorio y abre la pantalla
**Probar conexión con API** desde el icono de nube del feed Flutter. La pantalla
consume `GET /api/demo` al abrirse y envía el texto ingresado mediante
`POST /api/demo`. Ambas respuestas usan el envelope JSON compartido en
`services/common/httpresponse`.

Verifica el gateway con `go test ./...`. Para probar la comunicación de Flutter
contra el gateway en ejecución, ejecuta desde `app/flutter_app`:

```powershell
flutter test integration_test/api_gateway_test.dart -d windows
flutter test integration_test/api_gateway_test.dart -d <id-del-emulador>
```

La configuración del cliente está en `app/flutter_app/lib/config/api_config.dart`.
Usa `localhost:8080` en desktop/simulador iOS, `10.0.2.2:8080` en emuladores
Android, o `--dart-define=API_BASE_URL=http://<ip-del-equipo>:8080` para un
dispositivo físico. El manifiesto de depuración Android habilita HTTP local;
usa HTTPS en producción.

## Feed de demostración

`GET /api/feed?category=recent&page=1&pageSize=5` devuelve publicaciones en
el envelope JSON compartido, junto con `page`, `pageSize` y `hasMore`.
`category` admite `recent`, `friends` y `suggested`; `pageSize` acepta valores
entre 1 y 20. Este endpoint entrega datos de demostración del Gateway: el
`post-service` actual aún no expone endpoints HTTP de publicaciones ni almacena
el feed. No debe usarse como fuente persistente de contenido en producción.

Prueba las categorías y la paginación con:

```powershell
Invoke-RestMethod "http://localhost:8080/api/feed?category=recent&page=1&pageSize=5"
Invoke-RestMethod "http://localhost:8080/api/feed?category=friends&page=2&pageSize=5"
Invoke-RestMethod "http://localhost:8080/api/feed?category=suggested&page=1&pageSize=5"
```
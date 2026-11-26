# API Gateway

El Gateway escucha en el puerto `8080`. La especificación OpenAPI está en
[`openapi.yaml`](./openapi.yaml) y su copia accesible para documentación en
[`docs/api/openapi.yaml`](../../docs/api/openapi.yaml).

## Inventario de rutas registradas

Los prefijos con `ANY` reenvían el método HTTP y la ruta al servicio; las
operaciones que realmente existen en cada destino están descritas en OpenAPI.

| Gateway | Destino | Prefijo conservado/recortado | Variable de entorno |
| --- | --- | --- | --- |
| `GET /health` | Respuesta local del Gateway | — | — |
| `ANY /test`, `/test/*` | `test-service` | Recorta `/test` | `TEST_SERVICE_URL` |
| `ANY /users`, `/users/*` | `user-service` | Conserva `/users` | `USER_SERVICE_URL` |
| `ANY /auth`, `/auth/*` | `auth-service` | Conserva `/auth` | `AUTH_SERVICE_URL` |
| `ANY /friends`, `/friends/*` | `social-service` | Conserva `/friends` | `SOCIAL_SERVICE_URL` |
| `ANY /recommendations`, `/recommendations/*` | `recommendation-service` | Conserva `/recommendations` | `RECOMMENDATION_SERVICE_URL` |
| `ANY /media`, `/media/*` | `media-service` | Conserva `/media` | `MEDIA_SERVICE_URL` |
| `ANY /v1/users`, `/v1/users/*` | `user-service` | Conserva `/v1/users` | `USER_SERVICE_URL` |
| `ANY /v1/media`, `/v1/media/*` | `media-service` | Conserva `/v1/media` | `MEDIA_SERVICE_URL` |
| Pública | `GET /health` | Respuesta local del gateway | Ninguna |
| Pública | `GET /api/demo` | Respuesta JSON de prueba del gateway | Ninguna |
| Pública | `POST /api/demo` | Valida y devuelve el mensaje recibido | Ninguna |
| Pública | `GET /api/feed` | Feed paginado de demostración por categoría | Ninguna |
| Pública | `ANY /test` y `ANY /test/*any` | `test-service`, quitando `/test` | `TEST_SERVICE_URL` |
| Privada | `ANY /users` y `ANY /users/*any` | `user-service`, conservando `/users` | `USER_SERVICE_URL` |
| Pública | `ANY /auth` y `ANY /auth/*any` | `auth-service`, conservando `/auth` | `AUTH_SERVICE_URL` |

El enrutador todavía no publica rutas `/v1` para Auth, Social ni Recomendaciones,
ni rutas para Publicaciones, Música, Chat, Notificaciones generales o
Gamificación. Las preferencias de notificación y el inventario existentes son
operaciones del prefijo de Usuarios.

La ruta versionada de social-service reenvía `/v1/social/...` sin reescrituras;
Social Service exige un Bearer token: su middleware consulta la sesión a Auth
Service y obtiene la identidad del usuario de la respuesta validada. El
encabezado `X-User-ID` enviado por el cliente no se utiliza como identidad
confiable. El sistema utiliza tokens de sesión opacos, no JWT.

Las variables de destino aceptan una URL completa. Si no están definidas, el
gateway usa los valores locales `http://localhost:8081`,
`http://localhost:8082` y `http://localhost:8083`, respectivamente.
## Seguridad y propagación de identidad

El Gateway valida las peticiones protegidas de `/users`, `/friends`,
`/recommendations` y `/media` llamando a `GET /auth/validate` con el Bearer
`session_token`. Antes del proxy elimina `X-User-ID`, `X-User-Id` y
`X-User-Role`, y solo propaga la identidad devuelta por Auth.

El Gateway permite CORS únicamente desde `http://localhost:8090` para habilitar
el botón “Try it out” de Swagger UI en desarrollo.
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
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
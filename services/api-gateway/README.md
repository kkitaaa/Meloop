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

## Compatibilidad de seguridad pendiente

El proxy reenvía `Authorization` pero no autentica por sí mismo. En su estado
actual, `user-service` y `social-service` esperan `X-User-ID` (y algunas rutas
administrativas `X-User-Role`), mientras que `auth-service` emite
`session_token`. El contrato declara Bearer/JWT como objetivo, pero esa
integración debe completarse antes de considerar operativas las pruebas
protegidas de Swagger UI.

El Gateway permite CORS únicamente desde `http://localhost:8090` para habilitar
el botón “Try it out” de Swagger UI en desarrollo.

La especificación marca las rutas protegidas con Bearer y `x-required-roles`;
esto aún no está aplicado uniformemente por los servicios. En particular,
`GET/POST /users`, las verificaciones de Media, Social y Recomendaciones no
validan hoy ese esquema de forma consistente.

Además, el Compose raíz no define `social-service` ni
`recommendation-service`, y el Gateway no configura sus URLs en ese Compose.
Esas rutas están registradas en el código, pero no son ejercitables con la
configuración Compose actual.
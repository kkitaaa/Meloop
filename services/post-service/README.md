# Post service: guardados y compartición

Estas rutas requieren una sesión válida mediante el middleware de autenticación
del servicio (PST-02). El usuario se obtiene del token; no se acepta un
`user_id` proporcionado por el cliente.

| Método | Ruta | Comportamiento |
|---|---|---|
| `POST` | `/v1/posts/{id}/save` | Guarda la publicación; repetir la petición no duplica el registro. |
| `DELETE` | `/v1/posts/{id}/save` | Quita el guardado; repetir la petición también es válido. |
| `GET` | `/v1/me/saved?limit=20&offset=0` | Devuelve publicaciones guardadas y `hasMore`. |
| `POST` | `/v1/posts/{id}/share` | Envía una referencia de publicación al chat del amigo. |

Para compartir, el cuerpo es `{"friend_id":"..."}`. Social service valida que el
destinatario sea amigo aceptado y que no exista bloqueo en ninguna dirección.
Post service envía luego el mensaje al endpoint interno de chat-service y
responde con su `message_id`; no genera un enlace fijo.

La tabla `saved_posts` se crea con la migración
[`20261128000000_saved_posts.sql`](../../supabase/migrations/20261128000000_saved_posts.sql).
El detalle relacional también está en
[`Diagrama_MER_guardados.md`](../../docs/diagrams/Diagrama_MER_guardados.md).

Configuración necesaria para compartir:

- `SOCIAL_SERVICE_URL`: URL de social-service (por defecto `http://localhost:8086`).
- `CHAT_SERVICE_URL`: URL de chat-service (por defecto `http://localhost:8091`).
- `CHAT_SERVICE_INTERNAL_TOKEN`: secreto compartido entre post-service y chat-service.
- `POST_SHARING_TIMEOUT`: timeout de las solicitudes de validación y entrega.

El cliente devuelve `503 SHARING_UNAVAILABLE` si un servicio de dependencia no
está disponible; no confirma una compartición que no haya sido persistida.

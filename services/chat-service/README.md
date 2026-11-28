# Chat service

El servicio persiste mensajes en `mensaje_chat`. Post service usa el endpoint
interno protegido para enviar referencias a publicaciones:

```http
POST /internal/messages/share
Authorization: Bearer <CHAT_SERVICE_INTERNAL_TOKEN>
Content-Type: application/json

{"sender_id":"...","recipient_id":"...","post_id":"..."}
```

La respuesta `201` contiene el `message_id` persistido. El contenido almacenado
es una referencia JSON (`type: "post_share"` y `post_id`), no un enlace fijo.
La amistad y los bloqueos se validan antes en social-service mediante
post-service.

Variables de entorno:

- `DATABASE_URL`: conexión PostgreSQL.
- `CHAT_SERVICE_INTERNAL_TOKEN`: token requerido por el endpoint interno.
- `CHAT_SERVICE_PORT`: puerto HTTP (por defecto `8091`).

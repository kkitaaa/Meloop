# Social Service

Microservicio encargado de la gestión de relaciones sociales en Meloop:
- Solicitudes de amistad (RF-10, RF-11).
- Amigos y perfiles de amistad (RF-12).
- Bloqueo bidireccional y validación de interacciones (RF-13).
- Sugerencias de amistad con motor de compatibilidad (RF-14, RF-15).
- **Centralización de exclusión de usuarios bloqueados** (Endpoint interno y cliente Go).

## Endpoints Principales

### Rutas Protegidas (`/friends`) - Requiere sesión / Header `X-User-ID`:
- `POST /friends/requests`: Enviar solicitud de amistad.
- `GET /friends/requests/received`: Listar solicitudes recibidas.
- `GET /friends/requests/sent`: Listar solicitudes enviadas.
- `POST /friends/requests/:id/accept`: Aceptar solicitud.
- `POST /friends/requests/:id/reject`: Rechazar solicitud.
- `POST /friends/requests/:id/cancel`: Cancelar solicitud.
- `GET /friends`: Listar amigos.
- `DELETE /friends/:id`: Eliminar amigo.
- `GET /friends/:id/profile`: Consultar perfil de amigo.
- `POST /friends/blocks`: Bloquear a un usuario.
- `GET /friends/validate-interaction`: Validar interacción bidireccional entre usuarios.
- `GET /friends/suggestions`: Obtener sugerencias de amistad ordenadas por compatibilidad.
- `POST /friends/blocks/exclude`: Excluir usuarios bloqueados para el usuario autenticado.

### Endpoint Interno (`/internal`) - Comunicación entre Microservicios:
- `POST /internal/blocks/exclude`:
  Recibe `user_id` y `candidate_ids` en JSON.
  Retorna `candidate_ids` (conservados) y `excluded_ids` (bloqueados).

Consulta la documentación completa en [`docs/centralized_block_exclusion.md`](../../docs/centralized_block_exclusion.md).

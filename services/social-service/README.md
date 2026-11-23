# Social Service

Microservicio encargado de la gestión de relaciones sociales en Meloop:
- Solicitudes de amistad (RF-10, RF-11).
- Amigos y perfiles de amistad (RF-12).
- Bloqueo bidireccional y validación de interacciones (RF-13).
- Sugerencias de amistad con motor de compatibilidad (RF-14, RF-15).
- **Centralización de exclusión de usuarios bloqueados** (Endpoint interno y cliente Go).

## Endpoints Principales

### Rutas protegidas de amistades (`/friends`)
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

### Amigos en común
- `GET /v1/social/users/:id/mutual-friends?limit=20&cursor=`: Consultar amigos en común.

### Endpoint Interno (`/internal`) - Comunicación entre Microservicios:
- `POST /internal/blocks/exclude`:
  Recibe `user_id` y `candidate_ids` en JSON.
  Retorna `candidate_ids` (conservados) y `excluded_ids` (bloqueados).

Consulta la documentación completa en [`docs/centralized_block_exclusion.md`](../../docs/centralized_block_exclusion.md).

## Pruebas y cobertura

Desde la raíz del repositorio, ejecuta `make test-social` para correr `go test ./... -cover` en este servicio.
La cobertura medida para la lógica de negocio en el paquete `services` es **90.5 %**, por encima del mínimo de 60 %.

Las pruebas de integración de `tests/` requieren `TEST_DATABASE_URL` (base de datos cuyo nombre incluya `test`).
Sin esa variable se omiten para evitar conectarse a datos no destinados a pruebas.

La invalidación de caché de bloqueos (S-01) no aplica a la arquitectura actual: `social-service` no utiliza
una caché Redis para bloqueos. Por ese motivo no se implementa ni se pudo verificar invalidación de claves.

La consulta de amigos en común calcula la intersección en PostgreSQL, excluye cuentas con acciones de moderación
`BAN` o `SUSPENSION` y filtra candidatos bloqueados con el usuario consultante. La visibilidad de la lista se
configura mediante `visibilidad_amigos` en la configuración existente de privacidad.

Las rutas protegidas de social-service requieren `Authorization: Bearer <token>` y validan la sesión mediante
`GET /auth/validate` de auth-service. La identidad del contexto se toma únicamente de la respuesta validada;
`X-User-ID` no se acepta como prueba de identidad. Configura `AUTH_SERVICE_URL` con la URL interna de auth-service
(por defecto `http://localhost:8083`). Si auth-service no está disponible, las rutas protegidas responden
`503 Service Unavailable` y no permiten el acceso.

RN-05 se aplica exclusivamente a `GET /v1/social/users/:id/mutual-friends`: se considera inactivo un usuario cuyo
último inicio de sesión, publicación o interacción sea anterior a seis meses calendario. La fecha de creación
también protege a las cuentas recién registradas. Para cuentas existentes sin fecha de creación ni actividad
registrada, el filtrado es conservador y no las clasifica como inactivas. Una actividad posterior vuelve a
considerar activa la cuenta dentro de este endpoint.
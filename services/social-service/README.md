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
- `GET /friends`: Listar amigos con paginación por cursor.
- `DELETE /friends/:id`: Eliminar amigo.
- `GET /friends/:id/profile`: Consultar perfil de amigo.
- `POST /friends/blocks`: Bloquear a un usuario.
- `DELETE /friends/blocks/:id`: Desbloquear a un usuario.
- `GET /friends/validate-interaction`: Validar interacción bidireccional entre usuarios.
- `GET /friends/suggestions`: Obtener sugerencias de amistad ordenadas por compatibilidad.
- `POST /friends/blocks/exclude`: Excluir usuarios bloqueados para el usuario autenticado.

### Endpoint Interno (`/internal`) - Comunicación entre Microservicios:
- `POST /internal/blocks/exclude`:
  Recibe `user_id` y `candidate_ids` en JSON.
  Retorna `candidate_ids` (conservados) y `excluded_ids` (bloqueados).
- `GET /internal/users/:id/friends/count`: Retorna `{ "count": n }` y requiere
  `Authorization: Bearer <SOCIAL_SERVICE_INTERNAL_TOKEN>`.

Configura el mismo `SOCIAL_SERVICE_INTERNAL_TOKEN` en `social-service` y `user-service`.
El endpoint de conteo falla cerrado si el token no está configurado; los perfiles muestran HTTP 502
si `user-service` no puede obtener el conteo, en vez de devolver un valor local potencialmente incorrecto.

Consulta la documentación completa en [`docs/centralized_block_exclusion.md`](../../docs/centralized_block_exclusion.md).

## Eventos de cancelación por bloqueo

Al bloquear, `social-service` crea el bloqueo, marca las amistades aceptadas como `ELIMINADA` y cambia a `CANCELADA` todas las solicitudes pendientes entre ambos usuarios, en cualquiera de las direcciones, dentro de una transacción PostgreSQL. Repetir el mismo bloqueo no vuelve a modificar datos ni publica eventos.

Después del `COMMIT`, publica un evento por cada solicitud que cambió efectivamente a `CANCELADA` en el exchange `meloop.events`, con routing key `friend.request.cancelled` y este JSON:

```json
{
  "event_id": "uuid-v4",
  "event_type": "friend.request.cancelled",
  "request_id": "id-amistad",
  "sender_id": "id-emisor",
  "receiver_id": "id-receptor",
  "reason": "user_blocked"
}
```

La publicación es best-effort: si RabbitMQ falla después del `COMMIT`, el bloqueo permanece confirmado y el evento puede perderse. `notification-service` debe asociar la notificación a `request_id` para retirar solo la notificación de la solicitud cancelada.

### Paginación de amigos (RF-12)

`GET /friends?limit=20&cursor=...&q=...` acepta `limit` (predeterminado 20, máximo 50; cualquier valor positivo superior a 50 se ajusta automáticamente a 50), `cursor` (base64 devuelto por la página anterior) y `q` (búsqueda parcial por nombre de usuario). Mantén el mismo `q` al enviar cada `cursor`. El mismo endpoint está disponible tanto en el servicio social como a través del gateway.

Ejemplo: `GET /friends?limit=20&cursor=eyJmZWNoYV9hbWlzdGFkIjoi...&q=usuario`.

La migración que incorpora `fecha_amistad` asigna una fecha de referencia a las amistades aceptadas existentes que no tenían este campo. Esa fecha de referencia no representa necesariamente la fecha real en que fueron aceptadas.

La respuesta conserva `data` como la lista de amigos y agrega `next_cursor` al envoltorio existente:

```json
{
  "success": true,
  "data": [
    {
      "id_amistad": "friendship-id",
      "id_usuario": "user-id",
      "username": "usuario",
      "estado": "ACEPTADA"
    }
  ],
  "error": null,
  "next_cursor": "eyJmZWNoYV9hbWlzdGFkIjoi..."
}
```

Cuando no hay otra página, `next_cursor` es `""` y `data` es `[]`. Un `limit` menor o igual a cero devuelve HTTP 400; los cursores inválidos devuelven HTTP 400 con código `INVALID_CURSOR`.

## Pruebas y cobertura

Desde la raíz del repositorio, ejecuta `make test-social` para correr `go test ./... -cover` en este servicio.
La cobertura medida para la lógica de negocio en el paquete `services` es **90.5 %**, por encima del mínimo de 60 %.

Las pruebas de integración de `tests/` requieren `TEST_DATABASE_URL` (base de datos cuyo nombre incluya `test`).
Sin esa variable se omiten para evitar conectarse a datos no destinados a pruebas.
Configura `TEST_DATABASE_URL` con una base PostgreSQL de pruebas separada (por ejemplo, `social_test`) y aplica antes las migraciones de `supabase/migrations` a esa base.

Para levantar PostgreSQL local con Supabase CLI y ejecutar la suite de integración en una base aislada:

```powershell
supabase start
supabase status
$env:TEST_DATABASE_URL = "postgresql://postgres:<DB_PASSWORD>@localhost:15422/social_test?sslmode=disable"
psql "postgresql://postgres:<DB_PASSWORD>@localhost:15422/postgres" -c "CREATE DATABASE social_test"
Get-ChildItem supabase\migrations\*.sql | Sort-Object Name | ForEach-Object {
    psql $env:TEST_DATABASE_URL -v ON_ERROR_STOP=1 -f $_.FullName
    if ($LASTEXITCODE -ne 0) { throw "Falló la migración $($_.Name)" }
}
Push-Location services\social-service
go test ./... -count=1
Pop-Location
```

Usa la contraseña mostrada por `supabase status` en `<DB_PASSWORD>`. La base `social_test` es independiente de `postgres`; no ejecutes estas pruebas sobre datos de producción.

Para levantar PostgreSQL local con Supabase CLI y ejecutar la suite de integración en una base aislada:

```powershell
supabase start
supabase status
$env:TEST_DATABASE_URL = "postgresql://postgres:<DB_PASSWORD>@localhost:15422/social_test?sslmode=disable"
psql "postgresql://postgres:<DB_PASSWORD>@localhost:15422/postgres" -c "CREATE DATABASE social_test"
Get-ChildItem supabase\migrations\*.sql | Sort-Object Name | ForEach-Object {
    psql $env:TEST_DATABASE_URL -v ON_ERROR_STOP=1 -f $_.FullName
    if ($LASTEXITCODE -ne 0) { throw "Falló la migración $($_.Name)" }
}
Push-Location services\social-service
go test ./... -count=1
Pop-Location
```

Usa la contraseña mostrada por `supabase status` en `<DB_PASSWORD>`. La base `social_test` es independiente de `postgres`; no ejecutes estas pruebas sobre datos de producción.

La caché Redis para conteos sociales se habilita con `REDIS_URL` y usa `cache-aside` con TTL de 600 s para
`social:friends:count:{uid}` y `social:blocked:count:{uid}`. Las claves se invalidan tras aceptar o eliminar
una amistad y tras bloquear/desbloquear usuarios, y se respetan los fallos de Redis sin romper la operación
persistida en PostgreSQL.

La prueba `TestSocialCountCacheAsideInvalidationFallbackAndConcurrency` usa `TEST_DATABASE_URL` para
verificar las consultas SQL e invalidaciones contra una base aislada. También puede usar Redis real mediante
`TEST_REDIS_URL`; si no se configura, usa miniredis.

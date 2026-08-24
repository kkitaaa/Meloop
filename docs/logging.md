# Convenciones de logging

El proyecto utiliza logs estructurados enviados a `stdout`, para que Docker y cualquier agregador puedan capturarlos sin configuración adicional.

## Formato común

Los servicios Go usan `log/slog` y el ML usa el módulo estándar `logging` de Python. Cada registro incluye:

- `timestamp`
- `level`: `DEBUG`, `INFO`, `WARN` o `ERROR`
- `service`: identificador del servicio
- `event`: nombre estable del evento
- campos adicionales no sensibles

Ejemplo:

```json
{"time":"2026-08-23T12:00:00Z","level":"INFO","msg":"http_request","service":"user-service","method":"GET","path":"/users","status":200,"duration_ms":4}
```

## Niveles

- `DEBUG`: información detallada para desarrollo.
- `INFO`: inicio/cierre, solicitudes, operaciones y eventos normales.
- `WARN`: situaciones anómalas recuperables.
- `ERROR`: fallos que requieren atención.

En Go y Python se selecciona el nivel con `LOG_LEVEL`. El valor por defecto es `INFO`.

## Eventos recomendados

Usar nombres consistentes como:

- `service_started`
- `service_stopped`
- `http_request`
- `request_failed`
- `event_published`
- `event_consumed`
- `event_publish_failed`
- `event_consumer_failed`

Los servicios con HTTP registran método, ruta, estado y duración. No registran cuerpos completos de solicitudes.

## Información sensible

Nunca registrar:

- contraseñas
- tokens, cookies o cabeceras de autorización
- claves de API
- secretos de conexión
- cuerpos completos que puedan contener datos personales

Para depurar, registrar IDs técnicos o cantidades, siempre que no sean secretos y que cumplan con las políticas de privacidad.

## Go

El logger común está en `services/common/logging`. Los servicios Gin usan `GinMiddleware`; los servicios basados en `net/http` usan `HTTPMiddleware`. Los servicios de mensajería deben registrar publicación, consumo y errores con el nombre del evento, sin incluir payloads sensibles.

## Machine Learning

`ml-service` escribe eventos JSON a stdout, registra el ciclo de vida, las solicitudes HTTP y errores controlados. `/predict` registra `user_id` y `limit`, pero no la lista completa de preferencias ni el cuerpo de la solicitud.

## Docker

Los contenedores ya escriben sus logs a stdout/stderr. Para consultarlos:

```powershell
docker compose logs -f user-service
docker compose logs --tail=100 api-gateway
```

Para ver todos los servicios:

```powershell
docker compose logs -f
```

El formato queda disponible para Docker con `docker logs <container>` y no debe depender de archivos locales dentro del contenedor.

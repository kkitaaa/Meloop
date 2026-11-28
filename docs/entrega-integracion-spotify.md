# Entrega: integración de Spotify y preferencias musicales

Documento de traspaso para el equipo de backend/infraestructura. Describe lo que
ya está en el repositorio y las acciones necesarias para desplegarlo y probarlo
de extremo a extremo.

## Estado

La integración está implementada en código, pero todavía requiere migrar la base
compartida, configurar Secrets y desplegar las imágenes al clúster. No se ha
validado contra credenciales reales de Spotify ni contra el clúster remoto.

## Qué quedó implementado

### Catálogo Spotify

- Flutter consulta `GET /v1/music/search` con `q`, `type`, `offset` y `limit`.
  Los tipos admitidos son `track`, `artist`, `album` y `genre`; el límite de
  resultados por llamada es 20. Spotify acepta hasta 10 por petición, por lo
  que el backend combina páginas cuando se solicitan más.
- Flutter consulta detalles por `GET /v1/music/tracks/{id}`,
  `/v1/music/artists/{id}` y `/v1/music/albums/{id}`.
- `music-service` obtiene tokens con Spotify Client Credentials, los reutiliza
  hasta su vencimiento y consulta el catálogo. La búsqueda de géneros usa el
  catálogo curado de Meloop y no llama a Spotify.
- El Gateway envía las rutas de catálogo a `music-service`. Los errores de
  Spotify se comunican con status HTTP y códigos explícitos: un `429` de
  Spotify se traduce a `503 MUSIC_RATE_LIMITED` con `Retry-After`.
- La búsqueda devuelve resultados normalizados en `data.results` y guarda
  búsquedas exitosas en Redis durante cinco minutos. Si Redis no está
  configurado o falla, se registra el problema y la búsqueda continúa.

### Preferencias musicales persistentes

- Flutter usa:
  - `GET /v1/music/preferences`
  - `POST /v1/music/preferences`
  - `DELETE /v1/music/preferences/{type}/{item_id}`
- Las rutas pasan por Gateway hacia `user-service`. El servicio valida el
  Bearer token consultando `GET /auth/validate` de `auth-service` y toma de esa
  respuesta el ID del usuario. **No debe usarse `X-User-ID` como autenticación**
  para estas rutas.
- Las preferencias se guardan en la tabla existente `PREFERENCIA_MUSICAL`:
  `track → CANCION`, `artist → ARTISTA`, `genre → GENERO`. El máximo es 30 por
  categoría. Se conservan el nombre y la URL de imagen para que Flutter pueda
  volver a mostrar las preferencias.
- La actualización del perfil conserva esos metadatos cuando el ID musical
  sigue seleccionado.
- Una preferencia duplicada responde `409 ALREADY_EXISTS`; alcanzar el máximo
  responde `409 PREFERENCE_LIMIT`. Una petición inválida responde `400`.

Las respuestas exitosas usan el envelope Meloop:

```json
{
  "success": true,
  "data": {
    "preferences": {
      "songs": [{"id": "ID", "name": "Nombre", "artwork_url": "https://..."}],
      "artists": [],
      "genres": []
    },
    "limits": {"track": 30, "artist": 30, "genre": 30}
  },
  "error": null
}
```

### Archivos de despliegue/documentación

- `docker-compose.yml` configura `user-service`, `music-service` y Gateway.
- `.github/workflows/k8s-images.yml` incluye la publicación de la imagen de
  `user-service`, además de Gateway y `music-service`.
- `k8s/user-service.yaml` y `k8s/music-service.yaml` agregan Deployment y
  Service para el clúster local. Las imágenes `:dev` son para Kind; en el
  clúster del equipo deben usarse los tags publicados en GHCR.
- `docs/openapi.yaml` documenta los contratos HTTP.

## Acciones requeridas antes de cerrar la tarea

### 1. Preparar base de datos

- Aplicar, mediante el flujo de migraciones aprobado,  
  `supabase/migrations/20261128140100_music_preference_metadata.sql`.
- Confirmar en la base compartida que `PREFERENCIA_MUSICAL` tenga las columnas
  `nombre` y `artwork_url`, además de `spotify_id` y `fecha_creacion`.
- Verificar que `DATABASE_URL` apunta a la base persistente compartida y que
  `auth-service` y `user-service` usan la misma base/esquema.
- No borrar ni recrear la tabla: ya contiene preferencias utilizadas por
  perfiles y recomendaciones.

### 2. Configurar secretos y variables

- Cargar `SPOTIFY_CLIENT_ID` y `SPOTIFY_CLIENT_SECRET` en el gestor de Secrets
  aprobado y montarlos como variables del `music-service`. No incluir sus
  valores en este repositorio, archivos YAML, logs, tickets ni ejemplos.
- **Rotar el Client Secret que se compartió anteriormente** y usar el nuevo
  valor únicamente en el gestor de secretos.
- Configurar `SPOTIFY_MARKET` como código ISO 3166-1 alpha-2 válido (por
  ejemplo, `CL`). El servicio lo exige al iniciar.
- Asegurar que `meloop-secrets` contiene `DATABASE_URL`.
- Asegurar que `meloop-config` define `USER_SERVICE_URL`,
  `MUSIC_SERVICE_URL`, `AUTH_SERVICE_URL` y `MEDIA_SERVICE_URL` con los nombres
  DNS/puertos correctos del namespace.
- Configurar `REDIS_URL` para `music-service` con credenciales desde el gestor
  de secretos si Redis exige autenticación. Sin Redis no habrá caché de búsquedas.

### 3. Publicar y desplegar

- Construir/publicar las imágenes actualizadas de `api-gateway`,
  `user-service` y `music-service`. Confirmar que el tag de cada imagen
  corresponde al mismo commit que se va a desplegar.
- En el clúster compartido, cambiar las imágenes `:dev` de los manifiestos por
  los tags GHCR aprobados; no asumir que las imágenes locales de Kind existen
  en el clúster remoto.
- Aplicar primero la configuración y los Secrets, luego desplegar
  `auth-service`, `user-service`, `music-service` y el Gateway actualizado.
- Confirmar que los Services responden en los puertos esperados: Gateway `8080`,
  `user-service` `8082`, `auth-service` `8083` y `music-service` `8091`.
- Comprobar Pods `Ready`, eventos, logs y probes antes de probar Flutter.

### 4. Prueba de extremo a extremo

Usar una sesión de prueba y mantener el token fuera de logs, capturas y tickets.

1. Consultar el catálogo:
   `GET /v1/music/search?q=radiohead&type=track&limit=5`.
   Debe devolver resultados y metadatos utilizables por Flutter.
2. Consultar `GET /v1/music/preferences` con `Authorization: Bearer <sesión>`.
   Sin preferencias, debe responder con listas vacías y límites de 30.
3. Guardar una canción con `POST /v1/music/preferences`:

   ```json
   {
     "type": "track",
     "item_id": "<ID obtenido del catálogo>",
     "name": "<nombre obtenido del catálogo>",
     "artwork_url": "<URL https obtenida del catálogo>"
   }
   ```

4. Volver a consultar las preferencias y confirmar que aparece una sola vez,
   con ID, nombre e imagen correctos.
5. Reiniciar el Pod de `user-service` y volver a consultar: la preferencia debe
   persistir en la base de datos.
6. Eliminarla con
   `DELETE /v1/music/preferences/track/<ID>` y confirmar que ya no aparece.
7. Repetir al menos una vez para `artist` y `genre`, verificando los mapeos a
   `ARTISTA` y `GENERO`.
8. Repetir la petición de alta: debe responder `409 ALREADY_EXISTS` y no crear
   una fila duplicada. Probar también el límite de 30 por categoría.

### 5. Casos de autenticación y errores

- Sin token o con token inválido/expirado: `401`; el handler no debe ejecutarse.
- Enviar un `X-User-ID` distinto al usuario del Bearer token: la petición debe
  seguir asociada al ID validado por `auth-service`, nunca al header.
- Si `auth-service` no responde: `503`, no una identidad ficticia ni acceso
  permitido.
- Si falta configuración de Spotify o Spotify no responde: comprobar logs y
  errores HTTP; nunca registrar Client Secret ni tokens.
- Verificar búsqueda y detalle con Flutter además de `curl`, para confirmar
  envelope, nombres de campos, imágenes y manejo de errores.

## Validaciones ya realizadas

- `go test ./...` pasó en `services/user-service` y `services/api-gateway`.
- `docker compose config --quiet` y `git diff --check` pasaron.
- Las pruebas del catálogo usan respuestas de prueba; no reemplazan una
  validación con Spotify real.
- No se pudo validar el despliegue Kubernetes porque el contexto local no tenía
  un clúster accesible. El equipo debe ejecutar la verificación en el clúster
  destino.

## Referencias del repositorio

- Flutter: `app/flutter_app/lib/services/api_client.dart`
- Catálogo Spotify: `services/music-service/`
- Persistencia/autenticación de preferencias: `services/user-service/`
- Enrutamiento: `services/api-gateway/controllers/proxy_controller.go`
- Contratos: `docs/openapi.yaml`
- Despliegue local: `k8s/README.md`

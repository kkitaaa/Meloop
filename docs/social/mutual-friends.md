# Amigos en común

El API Gateway expone `GET /v1/social/users/{id}/mutual-friends` y lo enruta al
servicio social. La respuesta incluye `items`, `totalCount`, `hidden`,
`hasMore` y `nextCursor`. La paginación usa un cursor opaco basado en el ID
estable del último amigo, ordenado por ID; `limit` acepta valores de 1 a 50.

## Conexión local

1. Aplica las migraciones de Supabase, incluida
   `20261005120000_mutual_friends_privacy.sql`.
2. Inicia los servicios en Docker Compose. Ya se configura
   `SOCIAL_SERVICE_URL=http://social-service:8086`; ambos microservicios usan
   `DATABASE_URL`:

   ```powershell
   docker compose up -d --build social-service api-gateway
   ```

   Para ejecutarlos fuera de Docker, inicia el social-service en el puerto
   `8086` y el Gateway con `SOCIAL_SERVICE_URL=http://localhost:8086`.
3. Ejecuta Flutter pasando la URL del Gateway y el ID del usuario autenticado:

   ```powershell
   flutter run --dart-define=API_BASE_URL=http://localhost:8080 `
     --dart-define=CURRENT_USER_ID=<id-usuario>
   ```

   Para Android Emulator usa `http://10.0.2.2:8080`; en un dispositivo físico
   usa la IP local del equipo que ejecuta el Gateway.

`CURRENT_USER_ID` solo permite conectar el prototipo con el middleware actual,
que confía en `X-User-ID`. No es un mecanismo de autenticación adecuado para
producción; debe sustituirse por la identidad de la sesión autenticada. Las
publicaciones consumidas por el perfil deben incluir `user_id` para abrir el
perfil de la persona correcta.

## Privacidad y comportamiento

`CONFIGURACION_PRIVACIDAD.visibilidad_amigos` usa `PUBLICO` por defecto. El
valor `OCULTO` devuelve el conteo sin la lista. La persona solicitante y el
perfil objetivo se ocultan entre sí ante bloqueos, respondiendo con un 404
neutral (`La persona no está disponible.`), sin revelar la causa.

El perfil Flutter muestra hasta tres avatares cuando hay amigos en común,
oculta la fila con conteo cero y abre la lista paginada. La pantalla evita
agregar IDs repetidos y permite abrir el perfil de cada amigo.

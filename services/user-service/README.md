# User Service: privacidad

La configuración de privacidad RF-52 está disponible en las rutas autenticadas:

- `GET /users/me/privacy` obtiene las preferencias del usuario autenticado.
- `PATCH /users/me/privacy` actualiza parcialmente las preferencias.
- `PUT /users/me/privacy` es un alias compatible para actualizar preferencias.
- `GET` y `PATCH /users/me/privacidad` son alias en español.
- `/v1/users/me/privacy` publica los mismos métodos bajo el prefijo versionado.

Los campos aceptados son `visibilidad_perfil`, `visibilidad_publicaciones` y
`visibilidad_interacciones` (`PUBLICO`, `AMIGOS`, `PRIVADO`), además de
`recepcion_mensajes` y `recepcion_solicitudes_amistad` (`TODOS`, `AMIGOS`,
`NADIE`). El alias legado `recepcion_solicitudes` sigue aceptándose en las
peticiones. Las actualizaciones PATCH conservan los valores de los campos que
no se envían.

La configuración se persiste en `CONFIGURACION_PRIVACIDAD`. Las migraciones
`20261123174100_privacy_value_constraints.sql` y
`20261123201500_privacy_policy_functions.sql` normalizan y restringen los
valores, y agregan las funciones SQL usadas por perfiles, publicaciones,
solicitudes de amistad y validación de mensajes. Deben aplicarse a Supabase
antes de desplegar las versiones de los servicios que las consultan.

## Prueba funcional con k6

Desde la raíz del repositorio, con API Gateway, Auth, User y Social Service y
Supabase disponibles, ejecuta:

```powershell
k6 run privacy_test.js
```

El script comprueba lectura y persistencia de preferencias, denegación de
mensajes y solicitudes con `NADIE`, y habilitación con `TODOS`. Usa el Gateway
local en `http://localhost:8080` por defecto; puedes cambiarlo con
`$env:BASE_URL` antes de ejecutar k6. Cada iteración registra dos cuentas de
prueba nuevas y no las elimina, así que conviene usar una base de desarrollo o
pruebas. Para una ejecución reproducible de una sola iteración:

```powershell
k6 run --vus 1 --iterations 1 privacy_test.js
```

`social-service` valida las solicitudes al insertarlas y expone
`POST /friends/validate-message` para comprobar bloqueo y recepción de mensajes
antes de abrir una conversación. El envío de mensajes aún no tiene un endpoint
backend: `chat-service` contiene solo el arranque vacío y la interfaz actual
mantiene los mensajes en memoria. Cuando se implemente el envío, debe invocar
esta validación con el ID del destinatario antes de persistir o entregar el
mensaje. El cliente debe autenticar la llamada con el token de sesión.


Convenciones de Organización del Proyecto Flutter (/lib):

Proyecto Base: La aplicación se creó correctamente dentro de la ruta /app/flutter_app.

Plataformas: El proyecto ya está inicializado y probado. Corre de forma nativa tanto en Windows como en Android con la misma base de código.

Arquitectura Inicial: Para evitar que todos los archivos terminen mezclados, la carpeta lib/ quedó dividida de esta forma:

/core: Configuraciones globales de la app (temas, colores, network, constantes).

/features: Cada funcionalidad fuerte tiene su propia carpeta (auth, chat, music, posts, profile). La idea es que cada uno trabaje en su feature sin chocar con el resto.

/shared: Cosas reciclables que usaremos en varios lados (botones personalizados, modelos de datos, etc).

Archivos autogenerados por Flutter
Al crear el proyecto se generaron varias carpetas base. Acá un resumen rápido:

pubspec.yaml: Es el archivo clave. Acá es donde vamos a ir agregando las dependencias (librerías externas) y los assets (imágenes, fuentes).

android/, windows/ (y otras como ios/, web/): Contienen el código nativo de cada sistema. Por regla general no las vamos a tocar, a menos que necesitemos meter mano en permisos específicos de Android o Windows.

build/ y ocultas (.dart_tool, .idea): Son carpetas que guardan caché y la compilación de la app. Git las ignora automáticamente, así que no se preocupen por eso.

main.dart: Inicializa el MaterialApp de Meloop.

## Conexión con el API Gateway

La pantalla **Probar conexión con API** (icono de nube en el feed) ejecuta
`GET /api/demo` al abrirse y permite enviar un mensaje con `POST /api/demo`.
El feed también consume `GET /api/feed` con categoría y paginación (`page`,
`pageSize`). Los filtros recientes, amigos y sugeridos consultan categorías
distintas; el scroll infinito y el botón «Cargar más blogs» solicitan la página
siguiente y respetan `hasMore`.
El cliente HTTP y la URL base están centralizados en `lib/services/api_client.dart`
y `lib/config/api_config.dart`. Las respuestas siguen el envelope
`{ "success": true, "data": ..., "error": null }`.

Los elementos del feed son datos de demostración servidos por el API Gateway.
El post-service todavía no implementa una API HTTP ni persistencia para
publicaciones, por lo que estos datos no representan publicaciones reales.

### Crear publicaciones

El editor valida hasta cuatro imágenes JPG/PNG/WEBP de 5 MB cada una, un audio
de hasta 15 MB y un texto de hasta 2000 caracteres. Envía `POST /v1/posts` como
`multipart/form-data`: `content`, campos repetidos `images`, `audio` opcional y
`song_id`/`song_title`/`artist_name` cuando se eligió una canción. Muestra el
progreso de subida, permite cancelarla y mantiene el borrador ante errores.
`MUSIC_NOT_FOUND`, `MUSIC_UNAVAILABLE` y `POSTING_RESTRICTED` se traducen a
mensajes para el usuario.

La pantalla de selección consulta
`GET /v1/music/search?q=<texto>` y espera el envelope estándar con
`data.items`, donde cada elemento incluye `id`, `title`, `artist` y
opcionalmente `artwork_url`. La publicación exitosa espera `data.post` (o el
objeto de publicación en `data`) y opcionalmente `data.xp_awarded`.

**Dependencia del backend:** en este repositorio el API Gateway todavía no
enruta `/v1/posts` ni `/v1/music`; el `post-service` no implementa HTTP ni
persistencia y el `music-service` tampoco expone búsqueda. Por eso el cliente
está cableado al contrato esperado, pero publicar y buscar canciones requieren
que P-01, MU-07 y VS-03 expongan esos endpoints. La aplicación mostrará el error
de conexión del Gateway y conservará el borrador mientras no estén disponibles.

### Búsqueda global

El feed abre `lib/features/search/presentation/global_search_screen.dart` desde
el botón de búsqueda. La pantalla espera 400 ms tras escribir al menos dos
caracteres y consulta en paralelo:

- `GET /v1/search/users?q=<texto>` con `data.items` (`id_usuario`, `username`
  o `nombre`, `avatar_url`, `nivel` y `friend_status`).
- `GET /v1/search/posts?q=<texto>&filter=<text|song|artist|genre>` con
  `data.items` en formato de publicación compatible con Post Card.
- `GET /v1/music/search?q=<texto>` con canciones `id`, `title`, `artist` y
  `artwork_url`.

Cada pestaña conserva sus propios resultados, error y reintento; una fuente
fallida no bloquea las otras. Cambiar el texto cancela las solicitudes previas.
Al vaciar el campo se muestran hasta cinco búsquedas recientes de la sesión.
La acción Agregar usa `POST /v1/social/friends/requests` con `id_usuario`.

**Dependencia del backend:** este Gateway aún no registra `/v1/search/users`,
`/v1/search/posts` ni `/v1/music/search`; por ello las pestañas muestran su
estado de error/reintento hasta que BU-01, BU-02 y MU-02 estén disponibles.
Las solicitudes de amistad sí usan la ruta existente del social-service.

### Conexión local

Inicia el gateway desde la raíz del repositorio:

```powershell
cd services/api-gateway
go run .
```

La URL predeterminada es `http://localhost:8080` en Windows, iOS simulador y
desktop; en el emulador Android es `http://10.0.2.2:8080`. Para un teléfono
físico, conecta ambos dispositivos a la misma red e indica la IP local del
equipo donde corre el gateway:

```powershell
cd app/flutter_app
flutter run --dart-define=API_BASE_URL=http://192.168.1.10:8080
```

Reemplaza `192.168.1.10` por la IP real del equipo. Permite conexiones entrantes
al puerto `8080` en el firewall. En Android, el tráfico HTTP sin TLS se permite
solo en el manifiesto de depuración; para producción configura HTTPS.

Prueba los endpoints directamente:

```powershell
Invoke-RestMethod http://localhost:8080/api/demo
Invoke-RestMethod -Uri http://localhost:8080/api/demo -Method Post `
  -ContentType 'application/json' -Body '{"message":"Hola desde Flutter"}'
Invoke-RestMethod "http://localhost:8080/api/feed?category=recent&page=1&pageSize=5"
```

### Validación automatizada

Con el API Gateway ejecutándose en otra terminal, desde `app/flutter_app`:

```powershell
flutter test
flutter test integration_test/api_gateway_test.dart -d windows
flutter test integration_test/api_gateway_test.dart -d <id-del-emulador>
```

La prueba de integración envía el `GET` y el `POST` reales desde Flutter. El
emulador Android predeterminado usa `10.0.2.2`; Windows usa `localhost`. Para
un teléfono físico, usa `--dart-define=API_BASE_URL=http://<ip-del-equipo>:8080`.

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
El cliente HTTP y la URL base están centralizados en `lib/services/api_client.dart`
y `lib/config/api_config.dart`. Las respuestas siguen el envelope
`{ "success": true, "data": ..., "error": null }`.

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
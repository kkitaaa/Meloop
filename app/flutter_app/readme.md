
Convenciones de Organización del Proyecto Flutter (/lib):

Proyecto Base: La aplicación se creó correctamente dentro de la ruta /app/flutter_app.

Plataformas: El proyecto ya está inicializado y probado. Corre de forma nativa tanto en Windows como en Android con la misma base de código.

Arquitectura Inicial: Para evitar que todos los archivos terminen mezclados, la carpeta lib/ quedó dividida de esta forma:

/core: Configuraciones globales de la app (temas, colores, network, constantes).

/features: Cada funcionalidad fuerte tiene su propia carpeta (auth, chat, music, posts, profile). La idea es que cada uno trabaje en su feature sin chocar con el resto.

/shared: Cosas reciclables que usaremos en varios lados (botones personalizados, modelos de datos, etc).

Lo demás (Archivos autogenerados por Flutter)
Al crear el proyecto se generaron varias carpetas base. Acá un resumen rápido:

pubspec.yaml: Es el archivo clave. Acá es donde vamos a ir agregando las dependencias (librerías externas) y los assets (imágenes, fuentes).

android/, windows/ (y otras como ios/, web/): Contienen el código nativo de cada sistema. Por regla general no las vamos a tocar, a menos que necesitemos meter mano en permisos específicos de Android o Windows.

build/ y ocultas (.dart_tool, .idea): Son carpetas que guardan caché y la compilación de la app. Git las ignora automáticamente, así que no se preocupen por eso.

main.dart: Ya lo dejé limpio. Tiene un MaterialApp básico con un texto de prueba para que podamos empezar a conectar las pantallas desde ahí.

Convenciones de Organización del Proyecto Flutter (/lib):

/core: Contiene la configuración global y transversal de la app. Aquí van los colores y fuentes (/theme), las variables globales (/constants), la configuración de la API (/network) y funciones de ayuda generales (/utils).

/features: El núcleo de la aplicación separado por módulos. Cada funcionalidad independiente (como el chat, la música, el feed de posts o el perfil) tiene su propia carpeta. Ningún feature debe depender directamente de otro para evitar código enredado.

/shared: Componentes reutilizables. Si un botón personalizado o un modelo de datos (como una clase "Usuario") se va a usar en varios features distintos, debe ir aquí en /widgets o /models.
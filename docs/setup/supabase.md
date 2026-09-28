Guía de Configuración Local de Base de Datos

Asegurar requisitos previos: Mantener Docker Desktop abierto y ejecutándose en segundo plano. Es obligatorio para que el sistema pueda alojar los contenedores locales de la base de datos.

Actualizar el repositorio: Ejecutar git pull en la rama de trabajo para descargar la carpeta /supabase que contiene las configuraciones base y el archivo de migración con todas las tablas del Modelo Entidad-Relación.

Iniciar los servicios locales: Abrir la terminal en la raíz del proyecto y ejecutar el siguiente comando:


npx supabase start
La primera vez tardará unos minutos en descargar las imágenes de PostgreSQL, el gateway de API y el panel de administración.

Construir la estructura de tablas: Para asegurar que la base de datos se arme correctamente con las tablas y claves foráneas en el orden adecuado, ejecutar:


npx supabase db reset
Acceder al panel visual (Supabase Studio): Abrir el navegador de preferencia e ingresar explícitamente a http://localhost:54323. En el menú lateral izquierdo, bajo "Table Editor", estarán disponibles todas las tablas listas para ser consultadas o llenadas con datos de prueba.

Actualizar variables de entorno en Go: Cada desarrollador debe modificar el archivo .env de los microservicios que tenga a cargo (Auth, Social, Post, etc.), reemplazando la conexión a la nube por la credencial local exacta:
DB_URL="postgresql://postgres:postgres@127.0.0.1:54322/postgres"

Apagar el entorno al finalizar: Para detener los contenedores y liberar memoria RAM tras la jornada de desarrollo, ejecutar en la terminal:


npx supabase stop
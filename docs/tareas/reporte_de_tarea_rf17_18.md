# Documento Explicativo de Tarea

| Apartado | Detalle |
| :--- | :--- |
| **Identificación** | Tarea: RF-17 y RF-18 - Desarrollar la edición y eliminación segura de publicaciones.<br>Responsable: Equipo de Desarrollo<br>Fecha: 07/11/2026<br>Pull Request: [Añadir enlace o número del PR al crearlo] |
| **Objetivo** | - Crear el endpoint que permita al autor modificar el contenido de su publicación (RF-17).<br>- Implementar la ruta para eliminar publicaciones verificando siempre que el usuario que realiza la petición tenga los permisos de autoría (RF-18). |
| **Solución** | Se implementó el método `verifyOwnership` en el repositorio para validar atómicamente la autoría de un post antes de modificarlo o borrarlo, previniendo manipulaciones no autorizadas. El controlador ahora enruta los errores específicos a sus códigos HTTP correspondientes (404 y 403) para mantener buenas prácticas de seguridad en la API. |
| **Cambios realizados** | **Componentes:** `post-service/repositories/post_repository.go` y `post-service/controllers/post_controller.go`.<br>**Endpoints:** `PUT /api/v1/posts/:id` y `DELETE /api/v1/posts/:id`. |
| **Ejecución** | 1. Levantar el servicio: `cd services/post-service && go run main.go`.<br>2. Ejecutar prueba: `k6 run ml-service/tests/seguridad_posts.js`. |
| **Pruebas** | **Script k6:** `seguridad_posts.js`<br>**Escenario:** default (1 VU, 2s).<br>**Checks:** Edición y Eliminación bloqueada correctamente (404/403).<br>**Resultado Real:** 100% checks pasados (4 de 4). Tiempo de respuesta promedio (avg) = 5.23ms. |
| **Pendientes** | Migrar el script de k6 desde `ml-service/tests/` hacia `post-service/tests/k6/` para mantener la organización de directorios. |
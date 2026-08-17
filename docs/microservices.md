# Microservicios y responsabilidades

## 1. Introducción

El backend del sistema está compuesto por un conjunto de microservicios independientes desarrollados en Go. Cada microservicio mantiene una responsabilidad específica dentro del sistema, posee sus propios datos y expone las operaciones correspondientes a su dominio.

La separación de responsabilidades permite mantener los servicios desacoplados, facilitar el desarrollo paralelo y simplificar el mantenimiento del sistema.

Los microservicios definidos para el sistema son:

* `api-gateway`
* `auth-service`
* `user-service`
* `social-service`
* `post-service`
* `music-service`
* `media-service`
* `chat-service`
* `gamification-service`
* `recommendation-service`
* `moderation-service`
* `notification-service`

Además, el sistema cuenta con un servicio independiente de Machine Learning:

* `ml-service`

---

# 2. API Gateway

### Propósito

Actuar como punto único de entrada al backend y enrutar las solicitudes hacia los microservicios correspondientes.

### Responsabilidades

* Recibir solicitudes provenientes de la aplicación Flutter.
* Enrutar las solicitudes hacia los microservicios correspondientes.
* Centralizar el acceso a los servicios backend.
* Gestionar la comunicación HTTP/REST con los microservicios.
* Participar como punto de entrada para las conexiones WebSocket del chat.

### Datos propios

No posee datos propios.

### Comunicación

**Entrada:**

* HTTP/REST desde Flutter.
* WebSockets cuando corresponda al chat.

**Salida:**

* HTTP/REST hacia los microservicios.
* Conexión WebSocket hacia `chat-service`.

### No debe manejar

* Lógica de negocio propia de los microservicios.
* Persistencia de datos de dominio.

---

# 3. Auth Service

### Propósito

Gestionar la autenticación y las credenciales de los usuarios.

### Responsabilidades

* Registrar usuarios.
* Gestionar inicio de sesión.
* Gestionar credenciales.
* Generar y validar tokens.
* Aplicar las reglas relacionadas con autenticación.

### Datos propios

* Cuentas.
* Credenciales.

### Comunicación

**Entrada:**

* HTTP/REST desde `api-gateway`.

**Salida:**

* Respuestas de autenticación.
* Tokens.
* Eventos relacionados con autenticación mediante RabbitMQ cuando corresponda.

### Persistencia

* PostgreSQL.
* Redis puede utilizarse para sesiones si posteriormente es necesario.

### No debe manejar

* Perfiles.
* Publicaciones.
* Amistades.
* Música.
* Recomendaciones.
* Otras responsabilidades pertenecientes a los demás microservicios.

---

# 4. User Service

### Propósito

Gestionar los perfiles y la información asociada a los usuarios.

### Responsabilidades

* Gestionar perfiles.
* Gestionar información pública del usuario.
* Gestionar preferencias del usuario.

### Datos propios

* Información del perfil.

### Comunicación

**Entrada:**

* HTTP/REST desde `api-gateway`.

**Salida:**

* Información relacionada con perfiles y usuarios.

### Persistencia

* PostgreSQL.

### Redis

No se considera necesario inicialmente.

### No debe manejar

* Autenticación.
* Publicaciones.
* Amistades.
* Recomendaciones.
* Lógica perteneciente a otros dominios.

---

# 5. Social Service

### Propósito

Gestionar las relaciones sociales entre usuarios.

### Responsabilidades

* Gestionar amistades.
* Gestionar solicitudes de amistad.
* Gestionar relaciones entre usuarios.

### Datos propios

* Relaciones entre usuarios.

### Comunicación

**Entrada:**

* HTTP/REST.

**Salida:**

* Estado y datos relacionados con las relaciones sociales.
* Eventos sociales mediante RabbitMQ.

### Persistencia

* PostgreSQL.

### No debe manejar

* Perfiles.
* Autenticación.
* Publicaciones.
* Chat.

---

# 6. Post Service

### Propósito

Gestionar las publicaciones y sus interacciones.

### Responsabilidades

* Crear publicaciones.
* Modificar publicaciones.
* Consultar publicaciones.
* Gestionar comentarios.
* Gestionar likes y otras reacciones.

### Datos propios

* Publicaciones.
* Comentarios.
* Reacciones.

### Comunicación

**Entrada:**

* HTTP/REST.

**Salida:**

* Información de publicaciones e interacciones.
* Eventos de interacción mediante RabbitMQ.

### Persistencia

* PostgreSQL.

### Almacenamiento multimedia

Los archivos multimedia asociados a publicaciones son gestionados mediante `media-service`.

### No debe manejar

* Perfiles.
* Multimedia física.
* XP y recompensas.
* Recomendaciones.
* Moderación.

---

# 7. Music Service

### Propósito

Gestionar el catálogo musical de la plataforma.

### Responsabilidades

* Gestionar canciones.
* Gestionar artistas.
* Gestionar álbumes.
* Gestionar géneros.
* Consultar información del catálogo musical.

### Datos propios

* Información del catálogo musical.

### Comunicación

**Entrada:**

* HTTP/REST.

### Persistencia

* PostgreSQL.

### RabbitMQ

Su utilización queda condicionada a que posteriormente aparezcan eventos relacionados con el catálogo musical.

### Multimedia

Los archivos multimedia asociados a contenido musical son gestionados mediante `media-service`.

### No debe manejar

* Publicaciones.
* Perfiles.
* Recomendaciones.
* Machine Learning.

---

# 8. Media Service

### Propósito

Gestionar los archivos multimedia utilizados por la plataforma.

### Responsabilidades

* Subir archivos.
* Consultar archivos.
* Eliminar archivos.
* Gestionar metadatos de archivos.
* Gestionar el ciclo de vida de los archivos multimedia.

### Datos propios

* Metadatos de imágenes.
* Metadatos de audios.
* Metadatos de otros archivos.

### Comunicación

**Entrada:**

* HTTP/REST.

### Persistencia

* PostgreSQL para los metadatos.

### Almacenamiento de archivos

* MinIO/S3 para el almacenamiento físico de imágenes, audios y otros archivos.

### Redis

No se considera necesario inicialmente.

### No debe manejar

* Publicaciones.
* Música como dominio.
* Perfiles.
* Recomendaciones.
* Machine Learning.

---

# 9. Chat Service

### Propósito

Gestionar el chat privado y la comunicación en tiempo real entre usuarios.

### Responsabilidades

* Gestionar conversaciones.
* Gestionar mensajes.
* Gestionar conexiones WebSocket.
* Mantener la comunicación en tiempo real.
* Gestionar información relacionada con usuarios conectados.

### Datos propios

* Conversaciones.
* Historial de mensajes.

### Comunicación

**HTTP/REST:**

Para operaciones tradicionales relacionadas con el chat.

**WebSockets:**

Para comunicación en tiempo real.

La conexión WebSocket seguirá el siguiente flujo:

```text
Flutter
   │
   │ WebSocket
   ▼
API Gateway
   │
   ▼
chat-service
```

### Persistencia

* PostgreSQL para conversaciones e historial de mensajes.

### Redis

Se utilizará para:

* Estado temporal.
* Usuarios online.
* Conexiones activas.

### RabbitMQ

Puede publicar eventos relacionados con mensajes.

### No debe manejar

* Amistades.
* Perfiles.
* Publicaciones.
* Notificaciones.

---

# 10. Gamification Service

### Propósito

Gestionar la progresión y el sistema de recompensas de los usuarios.

### Responsabilidades

* Calcular y gestionar XP.
* Gestionar niveles.
* Gestionar recompensas.
* Gestionar desbloqueos.
* Actualizar el progreso del usuario.
* Publicar eventos relacionados con la progresión.

### Datos propios

* Progreso.
* Niveles.
* Recompensas.

### Comunicación

**Entrada:**

* HTTP/REST cuando corresponda.
* Eventos RabbitMQ como mecanismo principal para recibir acciones que generan progresión.

**Salida:**

* Progreso actualizado.
* Eventos de gamificación, como subida de nivel.

### RabbitMQ

Ejemplo:

```text
post-service
      │
      │ post.liked
      ▼
 RabbitMQ
      │
      ▼
gamification-service
      │
      │ actualiza XP
      ▼
user.level_up
      │
      ▼
 RabbitMQ
      │
      ▼
notification-service
```

### No debe manejar

* Acciones sociales.
* Publicaciones.
* Perfiles.
* Notificaciones.

---

# 11. Recommendation Service

### Propósito

Coordinar y entregar recomendaciones personalizadas.

### Responsabilidades

* Recibir solicitudes de recomendación.
* Preparar los datos necesarios para el procesamiento.
* Comunicarse con `ml-service`.
* Procesar y normalizar los resultados recibidos.
* Entregar recomendaciones a la aplicación.
* Diferenciar tipos de recomendación, principalmente:

  * música;
  * usuarios.

### Datos propios

* Recomendaciones de música.
* Recomendaciones de usuarios.

### Comunicación

**Entrada:**

* HTTP/REST desde `api-gateway`.

**Salida:**

* HTTP/REST hacia `ml-service`.
* Resultados de recomendaciones hacia la aplicación.

Flujo:

```text
Flutter
   │
   │ HTTP/REST
   ▼
API Gateway
   │
   ▼
recommendation-service
   │
   │ HTTP/REST
   ▼
ml-service
   │
   │ Resultado ML
   ▼
recommendation-service
   │
   ▼
Flutter
```

### Dependencias

Para generar recomendaciones puede requerir información proveniente de:

* `user-service`
* `music-service`
* `post-service`
* `social-service`
* `notification-service`
* `auth-service`

### Machine Learning

El algoritmo y procesamiento específico de Machine Learning no pertenecen a este servicio.

### No debe manejar

* Algoritmos de Machine Learning.
* Catálogo musical.
* Perfiles.
* Publicaciones.

---

# 12. Moderation Service

### Propósito

Gestionar los reportes y la moderación del contenido de la plataforma.

### Responsabilidades

* Recibir reportes de usuarios.
* Registrar reportes sobre publicaciones o comentarios.
* Consultar reportes pendientes.
* Gestionar la revisión de contenido reportado.
* Registrar la resolución de un reporte.
* Mantener el estado del proceso de moderación.

### Datos propios

* Reportes.
* Estados de moderación.

### Comunicación

**Entrada:**

* HTTP/REST.

**Salida:**

* Información y estado de los reportes.
* Eventos relacionados con moderación mediante RabbitMQ cuando corresponda.

### Persistencia

* PostgreSQL.

### Dependencias

Puede comunicarse con:

* `post-service` para publicaciones y comentarios.
* `user-service` para información de usuarios.
* `auth-service` para autenticación de moderadores.
* `notification-service` para notificaciones.
* `ml-service` para procesamiento relacionado con moderación cuando corresponda.

### No debe manejar

* Publicaciones.
* Comentarios.
* Perfiles.
* Notificaciones.

---

# 13. Notification Service

### Propósito

Gestionar las notificaciones que reciben los usuarios como consecuencia de eventos ocurridos en la plataforma.

### Responsabilidades

* Recibir eventos de otros microservicios.
* Crear notificaciones.
* Gestionar el estado de las notificaciones.
* Consultar las notificaciones de un usuario.
* Marcar notificaciones como leídas.
* Determinar el contenido de la notificación según el evento recibido.

### Datos propios

* Notificaciones.
* Estado de lectura.

### Comunicación

**Entrada principal:**

* Eventos RabbitMQ.

También puede recibir solicitudes mediante:

* HTTP/REST.

### Persistencia

* PostgreSQL.

### Ejemplos de eventos

```text
post.liked
comment.created
friend.requested
friend.accepted
message.sent
report.created
user.level_up
```

### No debe manejar

La lógica que origina los eventos.

Por ejemplo:

```text
post-service
    │
    │ post.liked
    ▼
RabbitMQ
    │
    ▼
notification-service
```

`notification-service` crea la notificación, pero la responsabilidad de determinar que ocurrió el like pertenece a `post-service`.

---

# 14. ML Service

### Propósito

Ejecutar el procesamiento de Machine Learning de manera independiente al backend principal.

### Responsabilidades

* Recibir datos preparados para generar recomendaciones.
* Procesar los datos mediante modelos de Machine Learning.
* Generar recomendaciones de:

  * música;
  * usuarios.
* Devolver los resultados al `recommendation-service`.
* Encapsular la lógica específica de los modelos de Machine Learning.

### Datos propios

No posee datos maestros propios del sistema.

### Comunicación

**Entrada:**

* HTTP/REST desde `recommendation-service`.

**Salida:**

* Resultados de recomendaciones.

### Flujo

```text
recommendation-service
          │
          │ HTTP/REST
          ▼
      ml-service
          │
          │ Procesamiento ML
          ▼
      Resultado ML
          │
          ▼
recommendation-service
```

### Tecnología

* Python.
* FastAPI.
* Scikit-learn.
* Pandas.
* NumPy.

### No debe manejar

* Usuarios como dominio principal.
* Canciones como dominio principal.
* Publicaciones.
* Autenticación.
* Almacenamiento multimedia.
* Reglas generales de negocio de recomendaciones.

Las reglas generales de negocio y la orquestación de recomendaciones pertenecen a `recommendation-service`.

---

# 15. Resumen de responsabilidades

| Servicio                 | Responsabilidad principal         | Persistencia / almacenamiento |
| ------------------------ | --------------------------------- | ----------------------------- |
| `api-gateway`            | Entrada y enrutamiento            | Ninguna                       |
| `auth-service`           | Autenticación y credenciales      | PostgreSQL                    |
| `user-service`           | Perfiles y preferencias           | PostgreSQL                    |
| `social-service`         | Amistades y relaciones            | PostgreSQL                    |
| `post-service`           | Publicaciones e interacciones     | PostgreSQL                    |
| `music-service`          | Catálogo musical                  | PostgreSQL                    |
| `media-service`          | Archivos multimedia               | PostgreSQL + MinIO/S3         |
| `chat-service`           | Conversaciones y tiempo real      | PostgreSQL + Redis            |
| `gamification-service`   | XP, niveles y recompensas         | PostgreSQL                    |
| `recommendation-service` | Orquestación de recomendaciones   | PostgreSQL                    |
| `moderation-service`     | Reportes y moderación             | PostgreSQL                    |
| `notification-service`   | Notificaciones                    | PostgreSQL                    |
| `ml-service`             | Procesamiento de Machine Learning | Sin persistencia inicial      |

---

# 16. Principio de separación de responsabilidades

Cada servicio debe mantener el control de su propio dominio.

La comunicación entre servicios no debe utilizarse para transferir responsabilidades de un servicio a otro.

Por ejemplo:

```text
post-service
    │
    ├── Gestiona publicaciones
    ├── Gestiona comentarios
    └── Gestiona likes
```

Mientras:

```text
gamification-service
    │
    ├── Calcula XP
    ├── Gestiona niveles
    └── Gestiona recompensas
```

Y:

```text
notification-service
    │
    ├── Recibe eventos
    ├── Crea notificaciones
    └── Gestiona su estado
```

De esta manera, un servicio puede reaccionar a eventos de otro sin asumir la responsabilidad del dominio que originó dicho evento.

# Infraestructura y almacenamiento

Esta sección define el uso de PostgreSQL, Redis, RabbitMQ y almacenamiento multimedia dentro de la arquitectura del sistema.

## 1. PostgreSQL

PostgreSQL será utilizado como sistema principal de persistencia para la información propia de los microservicios.

| Servicio                 | Información almacenada                   |
| ------------------------ | ---------------------------------------- |
| `auth-service`           | Cuentas y credenciales                   |
| `user-service`           | Información de perfiles                  |
| `social-service`         | Relaciones entre usuarios                |
| `post-service`           | Publicaciones, comentarios y reacciones  |
| `music-service`          | Información del catálogo musical         |
| `media-service`          | Metadatos de imágenes, audios y archivos |
| `chat-service`           | Conversaciones e historial de mensajes   |
| `gamification-service`   | Progreso, niveles y recompensas          |
| `recommendation-service` | Recomendaciones de música y usuarios     |
| `moderation-service`     | Reportes y estados de moderación         |
| `notification-service`   | Notificaciones y su estado               |

El `api-gateway` no posee datos propios y, por lo tanto, no utiliza PostgreSQL como almacenamiento de dominio.

## 2. Redis

Redis será utilizado para información temporal, caché y estados que requieran acceso frecuente.

En la arquitectura inicial se contempla principalmente su utilización en `chat-service` para:

* Estado temporal.
* Usuarios conectados.
* Conexiones activas.

`auth-service` puede utilizar Redis posteriormente para la gestión de sesiones si resulta necesario.

El resto de los servicios no requiere Redis inicialmente.

## 3. RabbitMQ

RabbitMQ será utilizado como sistema de mensajería para la comunicación asíncrona mediante eventos entre microservicios.

Entre los eventos definidos inicialmente se encuentran:

| Evento             | Productor              | Consumidor                                     |
| ------------------ | ---------------------- | ---------------------------------------------- |
| `post.liked`       | `post-service`         | `gamification-service`, `notification-service` |
| `comment.created`  | `post-service`         | `gamification-service`, `notification-service` |
| `friend.requested` | `social-service`       | `notification-service`                         |
| `friend.accepted`  | `social-service`       | `notification-service`                         |
| `message.sent`     | `chat-service`         | `notification-service`                         |
| `report.created`   | `moderation-service`   | `notification-service`                         |
| `user.level_up`    | `gamification-service` | `notification-service`                         |

RabbitMQ no será utilizado para las solicitudes síncronas normales, las cuales se realizarán mediante HTTP/REST.

El `recommendation-service` y el `ml-service` utilizarán inicialmente HTTP/REST para su comunicación.

## 4. Almacenamiento multimedia

Los archivos multimedia no serán almacenados directamente en PostgreSQL.

El almacenamiento de imágenes, audios y otros archivos se realizará mediante un sistema de almacenamiento de objetos compatible con S3, utilizando **MinIO/S3**.

La responsabilidad de gestionar estos archivos corresponde a `media-service`.

La separación será:

```text
┌──────────────────────┐
│    media-service     │
│                      │
│ Gestión de archivos  │
│ y metadatos          │
└──────────┬───────────┘
           │
     ┌─────┴─────┐
     │           │
     ▼           ▼
PostgreSQL     MinIO/S3
Metadatos      Archivos
```

### PostgreSQL

Almacena información como:

* Nombre del archivo.
* Tipo.
* Tamaño.
* Metadatos.
* Referencia o ubicación del archivo.

### MinIO/S3

Almacena físicamente:

* Imágenes.
* Audios.
* Otros archivos multimedia.

## 5. Resumen de infraestructura

```text
                    Microservicios Go
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
     PostgreSQL          Redis          RabbitMQ
     Persistencia    Estado temporal      Eventos
          │
          │
          ▼
     media-service
          │
          ▼
       MinIO/S3
       Multimedia
```

La infraestructura se distribuye según las necesidades de cada componente, evitando utilizar una tecnología de almacenamiento o comunicación cuando el servicio no la requiere.

## 6. Criterio de utilización

La arquitectura sigue los siguientes criterios:

* **PostgreSQL:** persistencia de información de dominio.
* **Redis:** información temporal, caché y estados de acceso frecuente.
* **RabbitMQ:** comunicación asíncrona mediante eventos.
* **MinIO/S3:** almacenamiento de archivos multimedia.

Estos componentes forman parte de la infraestructura común del sistema y serán ejecutados mediante Docker Compose en el entorno de desarrollo.



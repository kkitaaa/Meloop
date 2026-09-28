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

Los archivos multimedia no serán almacenados directamente en PostgreSQL/Supabase.

El almacenamiento de imágenes, audios y otros archivos se realizará mediante un sistema de almacenamiento de objetos compatible con S3, utilizando **MinIO** en el entorno local y servicios compatibles con S3 en producción.

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
Metadatos      Archivos (Binarios)
```

### PostgreSQL
Almacena exclusivamente los metadatos:
* Identificador único (`media_id`).
* Clave de objeto en MinIO (`object_key`).
* Tipo MIME (`content_type`).
* Tamaño en bytes.
* Referencia a la entidad asociada (`user_id`, `post_id`, etc.).
* Fecha de subida y metadatos complementarios.

### MinIO/S3
Almacena físicamente los archivos binarios organizados mediante buckets y object keys.

#### Configuración del Bucket y Seguridad
* **Nombre del Bucket:** `meloop-media`.
* **Visibilidad:** Estrictamente **privado** (acceso anónimo deshabilitado).
* **Persistencia:** Volumen persistente dedicado (`minio_data`).
* **Principio de Mínimo Privilegio:** `media-service` utiliza credenciales propias independientes (`MINIO_MEDIA_USER`) con una política IAM (`media-service-policy`) que restringe sus acciones exclusivamente a operaciones sobre `arn:aws:s3:::meloop-media` y `arn:aws:s3:::meloop-media/*`. Ningún microservicio tiene acceso administrativo o de root.
* **Presigned URLs:** El acceso y subida de archivos se realiza mediante URLs prefirmadas de subida (`PUT`) y descarga (`GET`) generadas por `media-service`, garantizando que MinIO no esté expuesto públicamente.

#### Estructura lógica de objetos
Las claves de objetos (*object keys*) generadas por el sistema siguen el estándar:

* **Fotos de perfil:** `profiles/{user_id}/{media_id}`
* **Banners de perfil:** `banners/{user_id}/{media_id}`
* **Multimedia de publicaciones:** `posts/{post_id}/{media_id}`

#### Restricciones del MVP
* **Imágenes:** Formatos soportados `JPEG` y `PNG`, tamaño máximo **5 MB**, resolución hasta **720p**.
* **Audio:** Formatos soportados `MP3` y `WAV`, tamaño máximo **15 MB**.
* **Procesamiento:** No se realiza compresión, transcodificación ni transformación previa de archivos antes de su almacenamiento en esta fase.

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



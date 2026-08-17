# Comunicación HTTP/REST

## 1. Flutter → API Gateway

La aplicación Flutter utiliza HTTP/REST como mecanismo principal
de comunicación con el backend. Todas las solicitudes dirigidas
a los servicios backend pasan por el API Gateway.

## 2. API Gateway → Microservicios

El API Gateway utiliza HTTP/REST para enrutar las solicitudes hacia
el microservicio responsable de cada dominio funcional.

|   Origen    |          Destino       |        Propósito          |
|-------------|------------------------|-------------------------- |
| Flutter     | API Gateway            | Acceso al backend         |
| API Gateway | Auth Service           | Autenticación             |
| API Gateway | User Service           | Usuarios y perfiles       |
| API Gateway | Social Service         | Relaciones sociales       |
| API Gateway | Post Service           | Publicaciones             |
| API Gateway | Music Service          | Información musical       |
| API Gateway | Media Service          | Gestión multimedia        |
| API Gateway | Chat Service           | Operaciones del chat      |
| API Gateway | Gamification Service   | Experiencia y recompensas |
| API Gateway | Recommendation Service | Recomendaciones           |
| API Gateway | Moderation Service     | Moderación                |
| API Gateway | Notification Service   | Notificaciones            |

## 3. Recommendation Service → ML Service

El Recommendation Service utiliza HTTP/REST para solicitar
recomendaciones al ML Service desarrollado en Python.

|   Origen               |    Destino        |        Propósito              |
|------------------------|-------------------|-------------------------------|
| Recommendation Service | ML Service        | Generación de recomendaciones |

# Comunicación asíncrona con RabbitMQ

RabbitMQ se utilizará como mecanismo de comunicación asíncrona entre los microservicios. Su objetivo será permitir que un servicio publique eventos cuando una acción haya ocurrido, mientras otros servicios interesados puedan procesarlos de manera independiente.

La comunicación mediante RabbitMQ se utilizará principalmente para acciones cuyos efectos no requieren una respuesta inmediata al usuario.

## 1. Principio de funcionamiento

La comunicación seguirá el siguiente esquema general:

```text
Microservicio productor
        │
        │ Publica evento
        ▼
     RabbitMQ
        │
        │ Distribuye evento
        ▼
Microservicio consumidor
```

Por ejemplo:

```text
post-service
     │
     │ post.liked
     ▼
 RabbitMQ
     │
     ├──────────────► gamification-service
     │
     └──────────────► notification-service
```

El `post-service` mantiene la responsabilidad de registrar el like. Los servicios consumidores se encargan posteriormente de ejecutar las acciones derivadas del evento.

## 2. Eventos definidos

| Evento             | Productor              | Consumidor             | Propósito                                           |
| ------------------ | ---------------------- | ---------------------- | --------------------------------------------------- |
| `post.liked`       | `post-service`         | `gamification-service` | Otorgar experiencia por una interacción             |
| `post.liked`       | `post-service`         | `notification-service` | Notificar al propietario de la publicación          |
| `comment.created`  | `post-service`         | `gamification-service` | Otorgar experiencia por realizar un comentario      |
| `comment.created`  | `post-service`         | `notification-service` | Notificar al propietario de la publicación          |
| `friend.requested` | `social-service`       | `notification-service` | Notificar una nueva solicitud de amistad            |
| `friend.accepted`  | `social-service`       | `notification-service` | Notificar la aceptación de una solicitud            |
| `message.sent`     | `chat-service`         | `notification-service` | Generar una notificación relacionada con un mensaje |
| `report.created`   | `moderation-service`   | `notification-service` | Notificar la creación de un reporte                 |
| `user.level_up`    | `gamification-service` | `notification-service` | Notificar que un usuario ha subido de nivel         |

## 3. Flujo de eventos

### Interacción con una publicación

```text
Flutter
   │
   │ HTTP/REST
   ▼
API Gateway
   │
   │ HTTP/REST
   ▼
post-service
   │
   │ Registra like
   │
   │ post.liked
   ▼
RabbitMQ
   ├──────────────► gamification-service
   │                    │
   │                    └──► Actualiza XP
   │
   └──────────────► notification-service
                        │
                        └──► Crea notificación
```

### Solicitud de amistad

```text
Flutter
   │
   ▼
API Gateway
   │
   ▼
social-service
   │
   │ friend.requested
   ▼
RabbitMQ
   │
   ▼
notification-service
   │
   └──► Crea notificación
```

### Subida de nivel

```text
gamification-service
        │
        │ user.level_up
        ▼
     RabbitMQ
        │
        ▼
notification-service
        │
        └──► Crea notificación
```

## 4. Separación entre HTTP/REST y RabbitMQ

La arquitectura diferencia las responsabilidades de ambos mecanismos:

**HTTP/REST**

Se utilizará para solicitudes que requieren una respuesta directa, como:

```text
Flutter → API Gateway
API Gateway → Microservicio
Recommendation Service → ML Service
```

**RabbitMQ**

Se utilizará para comunicar eventos derivados de acciones que ya fueron procesadas:

```text
post-service → RabbitMQ → gamification-service
post-service → RabbitMQ → notification-service
social-service → RabbitMQ → notification-service
chat-service → RabbitMQ → notification-service
```

De esta forma, los microservicios mantienen sus responsabilidades separadas y no necesitan realizar llamadas síncronas entre ellos para ejecutar todas las acciones derivadas de una operación.

## 5. Alcance inicial

En esta primera versión, RabbitMQ se limitará a los eventos definidos anteriormente.

El `recommendation-service` y el `ml-service` no utilizarán RabbitMQ inicialmente. Su comunicación se mantendrá mediante HTTP/REST, de acuerdo con la arquitectura definida para el sistema.

La definición de exchanges, routing keys, colas y consumidores será realizada durante la implementación de la infraestructura RabbitMQ.

# Comunicación en tiempo real con WebSockets

WebSockets se utilizará para permitir la comunicación en tiempo real entre la aplicación Flutter y el `chat-service`.

## 1. Arquitectura

El API Gateway será el punto único de entrada para las conexiones WebSocket.

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

El `chat-service` será responsable de gestionar las conversaciones, los mensajes y las conexiones WebSocket.

## 2. Comunicación del chat

Las operaciones tradicionales del chat utilizarán HTTP/REST:

```text
Flutter
   │
   │ HTTP/REST
   ▼
API Gateway
   │
   │ HTTP/REST
   ▼
chat-service
```

La comunicación en tiempo real utilizará WebSockets:

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

Esto permitirá que los mensajes puedan enviarse y recibirse sin necesidad de realizar solicitudes HTTP repetitivas.

## 3. Persistencia y estado

El `chat-service` utilizará:

* **PostgreSQL:** almacenamiento permanente de conversaciones y mensajes.
* **Redis:** manejo de información temporal relacionada con usuarios y conexiones online.

WebSockets se encargará exclusivamente de la comunicación en tiempo real; la persistencia de los mensajes continuará siendo responsabilidad del `chat-service`.

## 4. Alcance

WebSockets se utilizará principalmente para:

* Envío de mensajes en tiempo real.
* Recepción de mensajes en tiempo real.
* Comunicación asociada a las conexiones activas del chat.

Las operaciones que no requieran comunicación en tiempo real continuarán utilizando HTTP/REST.

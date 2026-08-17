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


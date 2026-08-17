# Separación de componentes

La arquitectura del sistema separa sus componentes principales en tres bloques: la aplicación multiplataforma desarrollada con Flutter, los servicios backend desarrollados en Go y el servicio independiente de Machine Learning desarrollado en Python.

Esta separación permite mantener diferenciadas las responsabilidades de presentación, lógica de negocio y procesamiento de Machine Learning.

## 1. Aplicación Flutter

La aplicación cliente será desarrollada utilizando Flutter y Dart.

Su responsabilidad corresponde principalmente a:

* Proporcionar la interfaz de usuario.
* Permitir la interacción del usuario con la plataforma.
* Consumir los servicios proporcionados por el backend.
* Mostrar los resultados obtenidos desde los servicios backend.
* Utilizar HTTP/REST para las operaciones correspondientes.
* Utilizar WebSockets para funcionalidades que requieren comunicación en tiempo real, como el chat.

La aplicación Flutter no será responsable de implementar la lógica de negocio de los microservicios ni los algoritmos de Machine Learning.

## 2. Backend y microservicios Go

El backend estará compuesto por un API Gateway y un conjunto de microservicios desarrollados en Go.

```text
Flutter
   │
   │ HTTP/REST
   │ WebSockets
   ▼
API Gateway
   │
   ├── auth-service
   ├── user-service
   ├── social-service
   ├── post-service
   ├── music-service
   ├── media-service
   ├── chat-service
   ├── gamification-service
   ├── recommendation-service
   ├── moderation-service
   └── notification-service
```

Los microservicios son responsables de la lógica de negocio correspondiente a sus respectivos dominios.

Entre sus responsabilidades se encuentran:

* Autenticación y credenciales.
* Gestión de usuarios y perfiles.
* Relaciones sociales.
* Publicaciones e interacciones.
* Catálogo musical.
* Archivos multimedia.
* Chat.
* Gamificación.
* Recomendaciones.
* Moderación.
* Notificaciones.

Los servicios backend utilizan HTTP/REST para las operaciones síncronas y RabbitMQ para la comunicación mediante eventos asíncronos.

## 3. Servicio de Machine Learning

El `ml-service` será un servicio independiente desarrollado en Python.

Su responsabilidad se limita al procesamiento específico de Machine Learning.

```text
recommendation-service
          │
          │ HTTP/REST
          ▼
      ml-service
          │
          │ Resultado ML
          ▼
recommendation-service
```

El `ml-service` será responsable de:

* Recibir datos preparados para el procesamiento.
* Ejecutar los modelos de Machine Learning.
* Generar recomendaciones de música y usuarios.
* Devolver los resultados al `recommendation-service`.

El `ml-service` no será responsable de:

* Gestionar usuarios.
* Gestionar perfiles.
* Gestionar publicaciones.
* Gestionar autenticación.
* Gestionar el catálogo musical.
* Definir las reglas generales de negocio de las recomendaciones.

La orquestación y las reglas generales de las recomendaciones corresponden al `recommendation-service`.

## 4. Comunicación entre componentes

La separación entre los componentes se mantiene mediante mecanismos de comunicación definidos:

```text
┌─────────────────────────┐
│     Flutter / Dart      │
│   Aplicación cliente    │
└────────────┬────────────┘
             │
       HTTP/REST
       WebSockets
             │
             ▼
┌─────────────────────────┐
│      API Gateway        │
│          Go             │
└────────────┬────────────┘
             │
             ▼
┌─────────────────────────┐
│    Microservicios Go    │
│                         │
│  Lógica de negocio      │
└────────────┬────────────┘
             │
       HTTP/REST
             │
             ▼
┌─────────────────────────┐
│      ML Service         │
│       Python            │
│   Machine Learning      │
└─────────────────────────┘
```

De forma complementaria, los microservicios pueden comunicarse mediante eventos utilizando RabbitMQ:

```text
Microservicio
     │
     │ Evento
     ▼
 RabbitMQ
     │
     ▼
Otro microservicio
```

De esta manera, cada componente mantiene responsabilidades independientes:

| Componente     | Responsabilidad principal             | Tecnología     |
| -------------- | ------------------------------------- | -------------- |
| Aplicación     | Interfaz e interacción con el usuario | Flutter / Dart |
| API Gateway    | Entrada y enrutamiento                | Go             |
| Microservicios | Lógica de negocio                     | Go             |
| ML Service     | Procesamiento de Machine Learning     | Python         |

Esta separación permite que la aplicación cliente, el backend y el procesamiento de Machine Learning evolucionen de forma independiente, manteniendo una comunicación definida entre los distintos componentes.

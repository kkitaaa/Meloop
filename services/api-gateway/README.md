# API Gateway

El API Gateway es el punto de entrada HTTP para la aplicación Flutter. Escucha
en el puerto definido por `PORT` (por defecto, `8080`) y reenvía las
solicitudes al microservicio correspondiente.

## Rutas

| Acceso | Gateway | Destino | Variable de entorno |
| --- | --- | --- | --- |
| Pública | `GET /health` | Respuesta local del gateway | Ninguna |
| Pública | `ANY /test` y `ANY /test/*any` | `test-service`, quitando `/test` | `TEST_SERVICE_URL` |
| Privada | `ANY /users` y `ANY /users/*any` | `user-service`, conservando `/users` | `USER_SERVICE_URL` |
| Pública | `ANY /auth` y `ANY /auth/*any` | `auth-service`, conservando `/auth` | `AUTH_SERVICE_URL` |

Las rutas privadas reciben y reenvían el encabezado `Authorization`; la
validación del token debe realizarse mediante la política de autenticación del
servicio correspondiente.

Las variables de destino aceptan una URL completa. Si no están definidas, el
gateway usa los valores locales `http://localhost:8081`,
`http://localhost:8082` y `http://localhost:8083`, respectivamente.

## Verificación local

Con los servicios levantados, la comunicación con `user-service` puede
verificarse con:

```powershell
Invoke-WebRequest -Uri http://localhost:8080/users -Method Get
```

En Docker Compose, el gateway usa los nombres de servicio de la red interna.
La prueba automatizada del proxy cubre reenvío de método, ruta, query y body,
además del error `502 Bad Gateway` cuando el destino no está disponible.
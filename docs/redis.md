# Redis

Redis es el almacenamiento de datos temporal y de acceso rápido de Meloop. Se utilizará principalmente para caché, estados temporales, usuarios conectados y conexiones activas.

## Configuración local

La instancia local se define en [backend/docker-compose.yml](../backend/docker-compose.yml):

| Parámetro | Valor |
| --- | --- |
| Servicio Docker | `redis` |
| Contenedor | `redis_server` |
| Imagen | `redis:7.2` |
| Host desde la máquina local | `localhost` |
| Puerto | `6379` |
| Volumen | `redis_data` |
| Persistencia | `/data` |
| Autenticación | Contraseña configurada en Docker Compose |

La contraseña actual es un secreto de desarrollo y no debe incluirse en documentación, commits, logs ni mensajes. Para entornos compartidos o producción debe trasladarse a una variable de entorno o a un gestor de secretos.

## Iniciar Redis

Desde la raíz del repositorio:

```powershell
docker compose -f backend/docker-compose.yml up -d redis
```

Comprobar el estado del contenedor:

```powershell
docker compose -f backend/docker-compose.yml ps redis
```

Detener Redis sin eliminar los datos:

```powershell
docker compose -f backend/docker-compose.yml stop redis
```

Detenerlo y eliminar el contenedor, conservando el volumen:

```powershell
docker compose -f backend/docker-compose.yml down
```

Para eliminar también los datos persistidos, usar `docker compose down -v`. Esta operación es destructiva para el volumen local.

## Probar la conexión

Desde el contenedor se puede ejecutar un `PING` usando la contraseña local configurada. Sustituir el marcador por el secreto fuera de la documentación y evitar guardar el comando en el historial del shell:

```powershell
docker exec -it redis_server redis-cli -a '<REDIS_PASSWORD>' ping
```

La respuesta esperada es:

```text
PONG
```

Como alternativa, el backend expone un endpoint de prueba que escribe y lee una clave:

```text
GET http://localhost:3000/redis-test
```

La respuesta esperada es:

```json
"valor"
```

Este endpoint existe únicamente como comprobación inicial. Las operaciones de negocio deben encapsularse en servicios específicos y usar nombres de claves con un espacio de nombres claro.

## Uso desde NestJS

El backend crea el cliente Redis en `backend/src/redis/redis.module.ts` y lo expone con el token `REDIS_CLIENT`. Los servicios que lo necesiten deben inyectar ese token en lugar de crear conexiones nuevas.

Recomendaciones:

- Reutilizar una conexión por proceso.
- Definir tiempos de expiración para datos temporales y caché.
- Usar nombres de clave consistentes, por ejemplo `chat:presence:<userId>`.
- No guardar contraseñas, tokens ni información sensible sin una justificación y protección adecuadas.
- Manejar errores de conexión y reconexión en servicios que dependan de Redis.
- No usar Redis como almacenamiento permanente de información de dominio.

## Responsabilidades previstas

| Caso de uso | Servicio principal |
| --- | --- |
| Usuarios conectados y conexiones activas | `chat-service` |
| Estado temporal de chats | `chat-service` |
| Sesiones, si se habilitan posteriormente | `auth-service` |
| Caché de respuestas frecuentes | Servicio que sea dueño del dato |

La comunicación entre `ml-service` y `recommendation-service` será inicialmente HTTP/REST; Redis no reemplaza esa comunicación.

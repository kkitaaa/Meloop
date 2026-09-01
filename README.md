# Meloop

Proyecto Meloop.

## Estándar de calidad y estilo

Este repositorio define una convención compartida para frontend, backend y servicio de ML con el objetivo de mantener un código consistente entre todos los integrantes del equipo.

### Herramientas base

- ESLint + Prettier para JavaScript/TypeScript.
- Dart formatter + análisis de Flutter para frontend.
- Ruff + Black para Python del servicio ML.

### Comandos principales

```bash
npm run format
npm run lint
```

### Dependencias recomendadas

```bash
python -m pip install ruff black
```

Si se trabaja con Flutter, instalar el SDK de Dart/Flutter en la máquina del desarrollador para habilitar `dart analyze` y `dart format`.

### Reglas básicas

- Usar 2 espacios para JavaScript/TypeScript y 4 espacios para Python.
- Mantener líneas de hasta 100 caracteres.
- Usar comillas simples en TypeScript y JavaScript cuando la configuración lo permita.
- Preferir `const`/`final` para valores inmutables.
- Evitar variables sin uso y logs de depuración en código final.
- Usar nombres de funciones descriptivos y camelCase/ snake_case según el lenguaje.
- Ejecutar el formateo antes de cada pull request.

### Documentación adicional

- [docs/style-guide.md](docs/style-guide.md)
- [docs/logging.md](docs/logging.md)
- [docs/testing.md](docs/testing.md)

## Entorno de desarrollo con Docker

El proyecto utiliza Docker y Docker Compose para ejecutar los microservicios del backend y los componentes de infraestructura necesarios durante el desarrollo.

Supabase se utiliza como servicio externo, por lo que PostgreSQL no se ejecuta como contenedor dentro del entorno local.

### Requisitos

Para levantar el entorno local es necesario contar con:

- Docker Desktop.
- Docker Compose.

Docker Desktop incluye Docker Compose en las instalaciones actuales, por lo que normalmente no es necesario instalarlo por separado.

Se puede comprobar la instalación mediante:

```bash
docker --version
docker compose version
```

### Variables de entorno

Antes de levantar el entorno se debe crear un archivo `.env` en la raíz del proyecto tomando como referencia el archivo `.env.example`.

Las variables utilizadas actualmente son:

```env
REDIS_PASSWORD=

RABBITMQ_USER=
RABBITMQ_PASSWORD=

# MinIO (Almacenamiento de objetos)
MINIO_ROOT_USER=
MINIO_ROOT_PASSWORD=
MINIO_MEDIA_USER=
MINIO_MEDIA_PASSWORD=
MINIO_MEDIA_BUCKET=meloop-media
MINIO_ENDPOINT=http://minio:9000
MINIO_PUBLIC_ENDPOINT=http://localhost:9000

SUPABASE_URL=
SUPABASE_KEY=
```

Las credenciales utilizadas para el desarrollo local deben definirse en el archivo `.env`.

El archivo `.env` no debe subirse al repositorio, ya que puede contener credenciales y otra información sensible. El archivo `.env.example` se mantiene en el repositorio únicamente como referencia para indicar las variables necesarias.

Las credenciales y configuración definitiva de Supabase deben ser proporcionadas por el equipo cuando el servicio correspondiente se encuentre disponible.

### Servicios del entorno

Docker Compose levanta inicialmente los siguientes componentes:

- API Gateway desarrollado en Go.
- Microservicio de prueba desarrollado en Go.
- Redis.
- RabbitMQ.
- MinIO (servidor de almacenamiento de objetos).
- MinIO Init (contenedor de aprovisionamiento automatizado de bucket y políticas).

Los servicios se comunican mediante una red interna de Docker Compose.

Redis se utiliza para caché y almacenamiento temporal, RabbitMQ para la comunicación asíncrona entre microservicios y MinIO para almacenamiento multimedia durante el desarrollo local.

### Levantar el entorno

Desde la raíz del repositorio ejecutar:

```bash
docker compose up --build
```

La opción `--build` permite construir nuevamente las imágenes de los microservicios antes de iniciar los contenedores.

Mientras este comando se encuentre activo, la terminal mostrará los logs generados por los diferentes servicios.

### Verificar los contenedores

Para comprobar el estado de los servicios ejecutar:

```bash
docker compose ps
```

Los contenedores deberían aparecer en estado `Up` o `Running` (el contenedor `meloop-minio-init` saldrá con código `0` tras completar el aprovisionamiento).

### Servicios disponibles

Una vez levantado el entorno se encuentran disponibles los siguientes servicios:

| Servicio | Dirección | Descripción / Credenciales |
| --- | --- | --- |
| API Gateway | `http://localhost:8080` | Punto de entrada del backend |
| Health API Gateway | `http://localhost:8080/health` | Verificación de estado de API Gateway |
| Test Service | `http://localhost:8081` | Microservicio de prueba |
| Health Test Service | `http://localhost:8081/health` | Verificación de estado de Test Service |
| RabbitMQ Management | `http://localhost:15672` | Consola web con `RABBITMQ_USER` / `RABBITMQ_PASSWORD` |
| MinIO API (S3) | `http://localhost:9000` | Endpoint de almacenamiento de objetos S3 |
| MinIO Console | `http://localhost:9001` | Consola web con `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` |
| Redis | `localhost:6379` | Servidor Redis con autenticación por `REDIS_PASSWORD` |

Las interfaces de RabbitMQ y MinIO utilizan las credenciales definidas en el archivo `.env`.

### Verificar comunicación entre microservicios

Los servicios definidos en Docker Compose se encuentran conectados mediante una red interna.

Para comprobar la comunicación entre el API Gateway y el microservicio de prueba se puede ejecutar:

```bash
docker compose exec api-gateway wget -qO- http://test-service:8081/health
```

Si la comunicación funciona correctamente, se debería obtener una respuesta similar a:

```json
{
  "service": "test-service",
  "status": "ok"
}
```

Dentro de la red de Docker Compose los servicios pueden comunicarse utilizando el nombre definido para cada servicio, por ejemplo `test-service`, `redis`, `rabbitmq` y `minio`.

### Detener el entorno

Para detener y eliminar los contenedores creados por Docker Compose ejecutar:

```bash
docker compose down
```

Los volúmenes configurados para Redis, RabbitMQ y MinIO se mantienen, permitiendo conservar los datos locales entre ejecuciones.

### Eliminar los volúmenes

Si además de detener los contenedores se necesita eliminar los datos almacenados localmente, ejecutar:

```bash
docker compose down -v
```

Este comando elimina los volúmenes asociados al entorno, por lo que debe utilizarse únicamente cuando se quiera reiniciar completamente los datos locales.

### Reconstruir el entorno

Después de realizar cambios en los Dockerfile o en los microservicios se puede reconstruir el entorno ejecutando nuevamente:

```bash
docker compose up --build
```

De esta forma todos los integrantes del equipo pueden levantar un entorno de desarrollo uniforme utilizando la misma configuración de Docker Compose.
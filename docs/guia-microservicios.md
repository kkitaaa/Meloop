# Guía para ejecutar y probar los servicios de Meloop

Esta guía describe cómo levantar la base de datos, la infraestructura, las API y
cada módulo Go desde la raíz del repositorio. Los comandos de consola están
escritos para PowerShell en Windows; los comandos de Go, npm, Docker y Supabase
también se pueden adaptar a Bash.

> **Importante:** `docker-compose.yml` no inicia PostgreSQL/Supabase. La base de
> datos local se ejecuta por separado con Supabase CLI. Compose inicia solo los
> servicios que aparecen en su archivo; los demás módulos Go deben iniciarse
> individualmente.

## 1. Requisitos

- Git.
- Go compatible con `go 1.26.6`, declarado en los módulos Go.
- Docker Desktop con Docker Compose.
- Node.js y npm para `backend/`.
- Python 3.11 o superior para `ml-service/`.
- Supabase CLI. Los comandos de esta guía la invocan con `npx supabase`.

Los servicios Go usan módulos independientes. Ejecuta `go run .` y `go test ./...`
desde el directorio del servicio, no desde la raíz. Varios módulos importan
`services/common` mediante una ruta local `replace`, por lo que conserva la
estructura del repositorio.

## 2. Base de datos local (Supabase/PostgreSQL)

Abre Docker Desktop y, desde la raíz del repositorio, inicia Supabase:

```powershell
npx supabase start
npx supabase status
```

La configuración de [`supabase/config.toml`](../supabase/config.toml) usa estos
puertos locales:

| Componente | URL/puerto |
| --- | --- |
| PostgreSQL | `127.0.0.1:15422` |
| Supabase API | `http://127.0.0.1:15421` |
| Supabase Studio | `http://127.0.0.1:15423` |
| Correo local de prueba | `http://127.0.0.1:15424` |

Aplica las migraciones del repositorio a la base local:

```powershell
npx supabase db reset
```

> `db reset` recrea la base **local** y vuelve a aplicar las migraciones. No lo
> ejecutes contra un proyecto remoto ni si necesitas conservar datos locales.
> Consulta `npx supabase status` para obtener el usuario, contraseña y URL exactos
> de conexión generados por tu entorno.

Los módulos ejecutados directamente en Windows deben recibir la URL de
PostgreSQL desde el entorno. Usa el valor que informa `supabase status`:

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
```

Reemplaza el marcador por la contraseña real de tu instancia local. No la
incluyas en documentación compartida ni la subas al repositorio. Algunos
servicios tienen valores por defecto que apuntan a `localhost:5432`; define
`DATABASE_URL` explícitamente para usar la base local de este proyecto, que
escucha en `15422`.

## 3. Redis, RabbitMQ y MinIO

La raíz contiene credenciales vacías en `.env.example`. Crea el archivo local y
completa las variables antes de iniciar Compose:

```powershell
Copy-Item .env.example .env
```

Edita `.env` y define, como mínimo, valores locales para:

```dotenv
REDIS_PASSWORD=<contraseña-local-de-redis>
RABBITMQ_USER=<usuario-local-de-rabbitmq>
RABBITMQ_PASSWORD=<contraseña-local-de-rabbitmq>
MINIO_ROOT_USER=<usuario-local-de-minio>
MINIO_ROOT_PASSWORD=<contraseña-local-de-minio>
MINIO_MEDIA_USER=<usuario-local-del-servicio-media>
MINIO_MEDIA_PASSWORD=<contraseña-local-del-servicio-media>
MINIO_MEDIA_BUCKET=meloop-media
```

Usa credenciales de desarrollo distintas a las de cualquier entorno real. No
guardes secretos en Git. Arranca solo la infraestructura cuando desarrolles los
servicios en procesos locales:

```powershell
docker compose up -d redis rabbitmq minio minio-init
docker compose ps
docker compose logs --tail=100 redis rabbitmq minio minio-init
```

`minio-init` prepara el bucket y las credenciales de media y puede terminar
después de completar esa tarea; no tiene que permanecer ejecutándose.

| Dependencia | Dirección local | Validación |
| --- | --- | --- |
| Redis | `localhost:6379` | `docker compose exec redis redis-cli -a <contraseña> PING` |
| RabbitMQ AMQP | `localhost:5672` | Dashboard en `http://localhost:15672` |
| RabbitMQ Management | `http://localhost:15672` | Inicia sesión con las credenciales de `.env` |
| MinIO S3 | `http://localhost:9000` | Consola en `http://localhost:9001` |

Los servicios iniciados desde Windows usan `localhost` como host de estas
dependencias. Los contenedores de Compose deben comunicarse usando los nombres
de servicio (`redis`, `rabbitmq`, `minio`) y la red de Compose.

## 4. Arranque de todo lo incluido en Docker Compose

Para iniciar los contenedores de aplicación que sí están declarados en el
Compose raíz:

```powershell
docker compose up --build -d
docker compose ps
docker compose logs -f --tail=100
```

Esto incluye el API Gateway, `test-service`, `user-service`, `auth-service`,
`media-service`, `ml-service`, Redis, RabbitMQ y MinIO. **No incluye PostgreSQL**
ni los demás servicios Go de este documento.

Para conectar `user-service` del contenedor a la base Supabase local, define en
`.env` una `DATABASE_URL` accesible desde Docker, por ejemplo usando
`host.docker.internal` y el puerto configurado por Supabase:

```dotenv
DATABASE_URL=postgresql://postgres:<contraseña-local>@host.docker.internal:15422/postgres?sslmode=disable
```

La configuración actual de Compose usa el valor predeterminado `5432` si esa
variable no está definida; ese puerto no coincide con el `15422` de la
configuración Supabase del repositorio.

El servicio `auth-service` también necesita `DATABASE_URL` y `REDIS_URL` cuando
se ejecuta en Docker. El Compose raíz no los define para ese contenedor; agrégalos
en una sobreescritura local de Compose antes de usar los flujos de autenticación.
No asumas que el contenedor está conectado a la base o a Redis solo porque esté
en estado `Up`.

El Gateway en Compose tiene destinos configurados para test, users y auth. Para
que también enrute a servicios desplegados en la misma red Docker, configura sus
destinos con nombres de servicio, por ejemplo `MEDIA_SERVICE_URL` como
`http://media-service:8085`, `SOCIAL_SERVICE_URL` como
`http://social-service:8086` y `RECOMMENDATION_SERVICE_URL` como
`http://recommendation-service:8087`. Esos servicios no están incluidos en el
Compose raíz, así que deben iniciarse aparte o añadirse a una configuración de
Compose adecuada.

Comandos habituales:

```powershell
docker compose ps
docker compose logs -f api-gateway
docker compose stop
docker compose down
```

`down` conserva los volúmenes. Evita `docker compose down -v` si quieres
conservar los datos de Redis, RabbitMQ o MinIO.

## 5. API NestJS (`backend/`)

Esta API es un servicio backend aparte de los módulos Go.

**Dependencias:** Node.js/npm. Redis solo es necesario para las funciones que lo
usen; se puede iniciar con la sección de infraestructura.

```powershell
Set-Location backend
npm install
npm run start:dev
```

La API escucha normalmente en `http://localhost:3000`. En otra terminal, prueba
la ruta documentada en el test e2e:

```powershell
Invoke-RestMethod http://localhost:3000/
```

Respuesta esperada: `Hello World!`.

Pruebas:

```powershell
npm test
npm run test:e2e
npm run test:cov
```

## 6. Servicio ML (`ml-service/`)

**Dependencias:** Python 3.11+, entorno virtual y las librerías de
`ml-service/requirements.txt`.

```powershell
Set-Location ml-service
py -3.11 -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install -r requirements.txt
python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8001
```

Verifica el proceso, readiness y documentación interactiva:

```powershell
Invoke-RestMethod http://localhost:8001/health
Invoke-RestMethod http://localhost:8001/ready
```

- API: `http://localhost:8001`.
- Swagger: `http://localhost:8001/docs`.
- Contrato OpenAPI: `http://localhost:8001/openapi.json`.
- Rutas de recomendaciones: `POST /recommendations/music` y
  `POST /recommendations/friends`.

`/health` verifica que el proceso está vivo. `/ready` debe responder `200` cuando
el servicio está listo; el contenedor de Compose consulta esta ruta en el puerto
`8001`.

Pruebas y dependencias de desarrollo:

```powershell
python -m pip install -r requirements-dev.txt
python -m pytest tests
```

## 7. Microservicios Go

### Convención de ejecución y pruebas

En cada sección, abre una terminal en la raíz del repositorio y ejecuta:

```powershell
Set-Location services\<nombre-del-servicio>
go mod download
go test ./...
go run .
```

`go test ./...` es el comando de pruebas unitarias disponible para cada módulo.
En módulos que todavía no tienen archivos `*_test.go`, Go indicará que no hay
pruebas; el comando sirve para compilar y verificar ese módulo, pero no
representa cobertura de comportamiento. No hay una prueba e2e común que levante
automáticamente todos los servicios.

En las instrucciones individuales, la línea `go test ./...` se ejecuta en el
directorio indicado, antes de `go run .`. Mantén el proceso de la API activo en
su terminal y haz las peticiones desde otra.

### API Gateway — `services/api-gateway`

- **Puerto:** `8080` (`PORT` permite cambiarlo).
- **Dependencias de ejecución:** servicios test, users y auth para sus rutas
  correspondientes; las rutas de media, amigos y recomendaciones requieren
  además sus servicios de destino.
- **Arranque y prueba:**

```powershell
Set-Location services\api-gateway
go mod download
go test ./...
go run .
```

En otra terminal:

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/test/health
```

La primera ruta es una comprobación local del Gateway. La segunda debe reenviar
la petición al `test-service` activo. El Gateway también enruta `/users`,
`/auth`, `/friends`, `/recommendations` y `/media`; no todas esas rutas tienen
un destino levantado en el Compose raíz.

### Auth — `services/auth-service`

- **Puerto:** `8083`.
- **Dependencias:** PostgreSQL local y Redis. `DATABASE_URL` y `REDIS_URL` se
  pueden definir en esta terminal antes de arrancar. SMTP solo se necesita para
  enviar correo real; durante desarrollo se pueden revisar correos de prueba
  con el servicio local de Supabase.
- **Rutas útiles:** `POST /auth/register`, `POST /auth/login`,
  `POST /auth/password-recovery`; rutas protegidas requieren token.

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:REDIS_URL = "redis://:<contraseña-local-de-redis>@127.0.0.1:6379"
Set-Location services\auth-service
go mod download
go test ./...
go run .
```

El servicio no registra una ruta `/health`. Prueba el flujo de registro/inicio
de sesión con una cuenta de desarrollo en vez de tratar cualquier respuesta de
login como un health check. Configuraciones adicionales incluyen
`PASSWORD_MIN_LENGTH`, `SESSION_TTL_HOURS`, `RECOVERY_TOKEN_TTL_MINUTES`,
`RECOVERY_URL_BASE` y `SMTP_*`.

### Users — `services/user-service`

- **Puerto:** `8082` (`PORT`).
- **Dependencias:** PostgreSQL; también llama a `media-service` para verificar
  medios (`MEDIA_SERVICE_URL`, por defecto `http://localhost:8085`).
- **Rutas:** `GET/POST /users`; las rutas `/users/me` y `/v1/users/...` requieren
  autenticación.

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:MEDIA_SERVICE_URL = "http://localhost:8085"
Set-Location services\user-service
go mod download
go test ./...
go run .
```

No tiene una ruta `/health` registrada. Una vez que la base y las migraciones
estén listas, comprueba una ruta pública como `GET http://localhost:8082/users`.

### Test — `services/test-service`

- **Puerto:** `8081`.
- **Dependencias:** ninguna.
- **Rutas de comprobación:** `/health`, `/validation-error`, `/unauthorized`,
  `/forbidden`, `/not-found`, `/internal-error` y `/panic`. Varias rutas devuelven
  deliberadamente errores para probar el formato de respuestas y el middleware.

```powershell
Set-Location services\test-service
go mod download
go test ./...
go run .
```

En otra terminal:

```powershell
Invoke-RestMethod http://localhost:8081/health
Invoke-RestMethod http://localhost:8080/test/health
```

La segunda llamada prueba el paso por el API Gateway si este también está
ejecutándose.

### Media — `services/media-service`

- **Puerto:** `8085` (`PORT`).
- **Dependencia:** MinIO. Variables relevantes:
  `MINIO_ENDPOINT`, `MINIO_PUBLIC_ENDPOINT`, `MINIO_MEDIA_USER`,
  `MINIO_MEDIA_PASSWORD` y `MINIO_MEDIA_BUCKET`.
- **Rutas:** `GET /health`, `POST /media/upload`, `POST /media/verify`.

Inicia MinIO con Compose y, en una terminal local:

```powershell
$env:MINIO_ENDPOINT = "localhost:9000"
$env:MINIO_PUBLIC_ENDPOINT = "http://localhost:9000"
$env:MINIO_MEDIA_USER = "<usuario-media-de-.env>"
$env:MINIO_MEDIA_PASSWORD = "<contraseña-media-de-.env>"
$env:MINIO_MEDIA_BUCKET = "meloop-media"
Set-Location services\media-service
go mod download
go test ./...
go run .
```

Comprueba `http://localhost:8085/health`. El proceso puede iniciar aunque MinIO
no esté disponible, pero las operaciones de almacenamiento requerirán una
conexión y credenciales válidas.

### Post — `services/post-service`

- **Puerto:** `8084`.
- **Dependencias:** PostgreSQL, MinIO y RabbitMQ. Configura `DATABASE_URL`
  (acepta `DB_URL` como alternativa), `MINIO_*` y `RABBITMQ_URL`. El servicio
  puede continuar aunque falle la publicación de un evento a RabbitMQ; registra
  una advertencia.
- **Rutas:** `GET /health`, operaciones de publicaciones, likes y comentarios
  bajo `/api/v1/posts`.

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:RABBITMQ_URL = "amqp://<usuario>:<contraseña>@localhost:5672/"
$env:MINIO_ENDPOINT = "localhost:9000"
Set-Location services\post-service
go mod download
go test ./...
go run .
```

En otra terminal, verifica `http://localhost:8084/health`.

### Social — `services/social-service`

- **Puerto:** `8086` (`SOCIAL_SERVICE_PORT`).
- **Dependencias:** PostgreSQL y RabbitMQ (`DATABASE_URL`, `RABBITMQ_URL`).
- **Rutas:** solicitudes de amistad, amigos, bloqueos y sugerencias bajo
  `/friends` y `/suggestions`. Las rutas están protegidas por autenticación.

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:RABBITMQ_URL = "amqp://<usuario>:<contraseña>@localhost:5672/"
Set-Location services\social-service
go mod download
go test ./...
go run .
```

No registra `/health`; valida sus endpoints proporcionando un token válido.
El API Gateway puede reenviar `/friends` a este módulo si se configura su URL de
destino.

### Recommendation — `services/recommendation-service`

- **Puerto:** `8087` (`RECOMMENDATION_SERVICE_ADDRESS`, por defecto `:8087`).
- **Dependencias:** Redis y el servicio ML para recomendaciones/readiness.
  `DATABASE_URL` y `RABBITMQ_URL` habilitan el acceso a perfiles y el consumidor
  de actividad; sin `DATABASE_URL`, ese consumidor queda desactivado.
- **Rutas:** `GET /health`, `GET /ready`, `GET /recommendations/{userID}`,
  `GET /recommendations/{userID}/music`,
  `GET /recommendations/{userID}/friends`, `POST /interactions`.

```powershell
$env:ML_SERVICE_URL = "http://127.0.0.1:8001"
$env:REDIS_URL = "redis://:<contraseña-local-de-redis>@127.0.0.1:6379"
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:RABBITMQ_URL = "amqp://<usuario>:<contraseña>@localhost:5672/"
Set-Location services\recommendation-service
go mod download
go test ./...
go run .
```

`DATABASE_URL` y `RABBITMQ_URL` son opcionales para iniciar la API, aunque se
necesitan para consumir y persistir eventos. Comprueba `/health` y luego `/ready`;
la segunda requiere que ML esté disponible en el URL configurado.

### Gamification — `services/gamification-service`

- **Puerto:** no verificado. `config.go` declara `8089`, pero `main.go` contiene
  bloques de arranque incompatibles y una versión antigua usa `8085`; no asumas
  que hay un listener activo ni que la ruta indicada abajo está publicada.
- **Dependencias:** PostgreSQL y RabbitMQ (`DATABASE_URL`, `RABBITMQ_URL`).
- **Ruta definida en el router:** `POST /api/v1/gamification/xp`; comprueba el
  montaje y la disponibilidad del endpoint cuando el módulo vuelva a compilar.

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:RABBITMQ_URL = "amqp://<usuario>:<contraseña>@localhost:5672/"
Set-Location services\gamification-service
go mod download
go test ./...
```

**Bloqueo actual:** `go test ./...` falla al compilar porque `UserProgress` está
declarado dos veces en `models/gamification.go`. No se puede iniciar ni probar
la API hasta resolver ese error y verificar el punto de entrada; después, la
configuración actual espera PostgreSQL y RabbitMQ.

### Notification — `services/notification-service`

- **Tipo:** worker/consumidor de mensajes; no abre un puerto HTTP.
- **Dependencias:** PostgreSQL, RabbitMQ y las migraciones del esquema de
  notificaciones.
- **Variables:** `DATABASE_URL`, `RABBITMQ_URL`.

```powershell
$env:DATABASE_URL = "postgresql://postgres:<contraseña-local>@127.0.0.1:15422/postgres?sslmode=disable"
$env:RABBITMQ_URL = "amqp://<usuario>:<contraseña>@localhost:5672/"
Set-Location services\notification-service
go mod download
go test ./...
go run .
```

El worker escucha eventos en el exchange `meloop.events` y escribe
notificaciones en PostgreSQL. Para una prueba funcional, publica un evento
válido (por ejemplo `user.level_up`) y confirma la fila resultante en Supabase
Studio. El README del servicio describe el payload y los criterios de
idempotencia.

### Chat, Moderation y Music

Los directorios `services/chat-service`, `services/moderation-service` y
`services/music-service` contienen actualmente un `main.go` mínimo que registra
que el proceso inició y termina. No levantan servidor HTTP ni tienen pruebas de
servicio configuradas; por tanto, no hay puerto ni endpoint que probar todavía.
Puedes comprobar que cada módulo compila con:

```powershell
Set-Location services\chat-service
go mod download
go test ./...
go run .
```

Repite cambiando `chat-service` por `moderation-service` y `music-service`.
`go run .` finalizará enseguida; esto es el comportamiento actual del código,
no un servicio que quede escuchando.

### Módulo compartido `services/common`

`common` no es un microservicio ejecutable; contiene logging, middleware y
respuestas HTTP reutilizados por otros módulos. Para validar el paquete:

```powershell
Set-Location services\common
go mod download
go test ./...
```

## 8. Kubernetes local con kind

Esta configuración es un entorno local para probar el flujo de recomendaciones:
API Gateway Go (`8080`) → `recommendation-service` (`8087`) → ML (`8001`), con
Redis (`6379`) y un volumen persistente. El Ingress expone el Gateway. No
despliega todos los microservicios de Compose ni crea PostgreSQL: la conexión
`DATABASE_URL` de recomendaciones es opcional y debe apuntar a una base externa
compatible si necesitas perfiles persistidos.

### Requisitos y creación del clúster

Instala Docker Desktop, `kubectl` y `kind`, e inicia Docker Desktop antes de
continuar. Verifica que el daemon esté disponible:

```powershell
docker info
kubectl version --client
kind version
```

Desde la raíz del repositorio, crea el clúster. La configuración mapea los
puertos HTTP/HTTPS del Ingress de kind a `localhost`:

```powershell
kind create cluster --name meloop --config kind\kind-config.yaml
kubectl cluster-info --context kind-meloop
kubectl get nodes
```

Si Docker Desktop no está iniciado, `docker info` o la creación del clúster
fallarán. Espera a que Docker indique que está ejecutándose antes de reintentar.

### Construcción y carga de imágenes

Los Dockerfiles Go necesitan el contexto de la raíz del repositorio porque
importan el módulo local `services/common`. El contexto del servicio ML es su
propio directorio:

```powershell
docker build -t meloop-api-gateway:dev -f services\api-gateway\Dockerfile .
docker build -t meloop-recommendation-service:dev -f services\recommendation-service\Dockerfile .
docker build -t meloop-ml-service:dev .\ml-service
kind load docker-image meloop-api-gateway:dev meloop-recommendation-service:dev meloop-ml-service:dev --name meloop
```

Si cambias el código y reconstruyes una imagen `:dev`, vuelve a ejecutar
`kind load docker-image` y reinicia el Deployment correspondiente para que los
Pods usen la imagen recién cargada:

```powershell
kubectl rollout restart deployment/backend -n meloop
```

Cambia `backend` por `recommendation-service` o `ml-service` según la imagen.

### Ingress y despliegue

Instala el controlador Ingress una vez por clúster y espera a que esté listo:

```powershell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.12.1/deploy/static/provider/kind/deploy.yaml
kubectl wait --namespace ingress-nginx --for=condition=Ready pod --selector=app.kubernetes.io/component=controller --timeout=180s
```

El Namespace debe estar activo antes de crear los recursos que viven en él.
Créalo primero y espera a que Kubernetes lo registre; luego aplica el resto de
la carpeta desde la raíz:

```powershell
kubectl apply -f .\k8s\namespace.yaml
kubectl wait --for=jsonpath='{.status.phase}'=Active namespace/meloop --timeout=60s
kubectl apply -f .\k8s\
kubectl get pods -n meloop -w
```

Cuando los Pods estén `Running` y listos (`READY`), sal del monitoreo con
`Ctrl+C`. Comprueba los Deployments, Services, PVC y el Ingress:

```powershell
kubectl get deployments,services,pvc,ingress -n meloop
kubectl rollout status deployment/backend -n meloop
kubectl rollout status deployment/recommendation-service -n meloop
kubectl rollout status deployment/ml-service -n meloop
```

El Secret `meloop-secrets` del manifiesto se crea vacío; no contiene credenciales.
Para conectar recomendaciones a Postgres/Supabase, crea el Secret con una
`DATABASE_URL` de desarrollo obtenida de tu gestor de secretos. No la guardes en
un archivo versionado ni la compartas:

```powershell
$env:DATABASE_URL = "<URL_DE_CONEXION>"
kubectl create secret generic meloop-secrets -n meloop --from-literal=DATABASE_URL="$env:DATABASE_URL" --dry-run=client -o yaml | kubectl apply -f -
kubectl rollout restart deployment/recommendation-service -n meloop
```

No vuelvas a aplicar `secret.yaml` después de crear esa credencial: representa
un Secret vacío.

### Pruebas

Para probar el backend sin Ingress, deja este comando ejecutándose en una
terminal: >.<

```powershell
kubectl port-forward svc/backend 8080:8080 -n meloop
```

Desde otra terminal, prueba el Gateway y la ruta de recomendaciones:

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/recommendations/1/music
```

La ruta de recomendaciones verifica el camino hacia `recommendation-service` y
ML. Sin `DATABASE_URL` o un perfil guardado, no habrá recomendaciones
personalizadas; eso no significa necesariamente que los Pods estén caídos.

Para probar directamente el servicio ML, deja abierto en una tercera terminal:

```powershell
kubectl port-forward svc/ml-service 8001:8001 -n meloop
```

Y consulta sus verificaciones desde otra terminal:

```powershell
Invoke-RestMethod http://localhost:8001/health
Invoke-RestMethod http://localhost:8001/ready
```

Si el Ingress Controller está listo, puedes probarlo sin editar el archivo
`hosts` de Windows enviando el encabezado `Host`:

```powershell
curl.exe -H "Host: meloop.local" http://127.0.0.1/health
```

Para que `http://meloop.local` funcione sin ese encabezado, agrega
`127.0.0.1 meloop.local` al archivo
`C:\Windows\System32\drivers\etc\hosts` desde un editor ejecutado como
administrador.

### Resiliencia y diagnóstico

Elimina un Pod del backend usando el nombre mostrado por el primer comando y
comprueba que el Deployment crea un reemplazo:

```powershell
kubectl get pods -n meloop -l app.kubernetes.io/name=backend
kubectl delete pod <NOMBRE_DEL_POD> -n meloop
kubectl get pods -n meloop -w
```

Para diagnosticar Pods que no inician o no pasan readiness:

```powershell
kubectl get pods -n meloop
kubectl describe pod <NOMBRE_DEL_POD> -n meloop
kubectl logs deployment/backend -n meloop
kubectl logs deployment/recommendation-service -n meloop
kubectl logs deployment/ml-service -n meloop
kubectl get pvc -n meloop
```

El PVC de Redis debe quedar `Bound`. Para detener el clúster y eliminar los
recursos locales:

```powershell
kind delete cluster --name meloop
```

Eliminar el clúster borra sus Pods y el volumen de Redis asociado.

## 9. Pruebas y comprobaciones rápidas

### Go

Ejecuta las pruebas dentro del módulo que modificaste:

```powershell
Set-Location services\api-gateway
go test ./...
```

Repite para los demás directorios `services/*` que tengan un `go.mod`. No
ejecutes `go test ./...` desde la raíz: cada servicio es su propio módulo.

### Endpoints de verificación

Con los servicios correspondientes en ejecución:

```powershell
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8081/health
Invoke-RestMethod http://localhost:8085/health
Invoke-RestMethod http://localhost:8087/health
Invoke-RestMethod http://localhost:8087/ready
Invoke-RestMethod http://localhost:8001/health
Invoke-RestMethod http://localhost:8001/ready
```

`/ready` de recommendation necesita que el servicio ML esté listo. No todos los
servicios tienen una ruta de health; revisa las notas de cada módulo antes de
interpretar un `404` como fallo de arranque.

### Detener el entorno

Detén cada proceso local con `Ctrl+C`. Después, desde la raíz:

```powershell
docker compose down
npx supabase stop
```

El comando `docker compose down` mantiene los volúmenes. Si ejecutaste solo
infraestructura, también detiene Redis, RabbitMQ y MinIO.

## 10. Puertos y servicios en Compose

| Servicio | Puerto local | ¿Está en Compose raíz? | Tipo/nota |
| --- | ---: | --- | --- |
| API Gateway | `8080` | Sí | API Go |
| Test | `8081` | Sí | API Go para pruebas de respuestas |
| Users | `8082` | Sí | API Go; requiere DB |
| Auth | `8083` | Sí | API Go; requiere DB y Redis |
| Post | `8084` | No | API Go; DB, MinIO y RabbitMQ |
| Media | `8085` | Sí | API Go; almacenamiento MinIO |
| Social | `8086` | No | API Go; DB y RabbitMQ |
| Recommendation | `8087` | No | API Go; Redis y ML |
| Gamification | No verificado | No | Actualmente no compila; ver el bloqueo en su sección |
| Chat, Moderation, Music | — | No | Esqueletos; no escuchan HTTP |
| Notification | — | No | Worker; no escucha HTTP |
| ML | `8001` | Sí | FastAPI |
| Backend NestJS | `3000` | No | API Node.js |
| PostgreSQL local | `15422` | No | Supabase CLI |
| Redis | `6379` | Sí | Cache |
| RabbitMQ | `5672`, `15672` | Sí | AMQP y dashboard |
| MinIO | `9000`, `9001` | Sí | S3 y consola |

El archivo Compose raíz es la referencia para saber qué contenedores están
realmente declarados; algunos documentos antiguos pueden mencionar puertos o
servicios distintos.

# PostgreSQL, Redis y Auth en Kubernetes universitario

Estos manifiestos agregan únicamente PostgreSQL 16, Redis 7.2 y un Job para las seis migraciones. Usan el namespace `student-acofre`; no cambian el Gateway, el Ingress ni los manifiestos de Kubernetes local. No crean otros microservicios.

## Requisitos verificados en el repositorio

- Auth consulta y escribe `USUARIO` y `TOKEN_RECUPERACION`; la migración inicial crea ambas tablas y sus campos usados por Auth.
- Auth emplea `DATABASE_URL` para PostgreSQL y `REDIS_URL` para Redis. Redis almacena las sesiones que se usan para validar el token.
- Las seis migraciones ordenadas por nombre usan SQL compatible con PostgreSQL 16 estándar. Incluyen `JSONB`, bloques `DO`, índices y alteraciones de columnas; no requieren extensiones ni objetos Supabase.
- El Job ejecuta los seis archivos originales, ordenados por nombre, con `psql --set ON_ERROR_STOP=1`; ante el primer error, termina sin aplicar los archivos siguientes. No se modifican las migraciones.

## Credenciales

Desde la raíz del repositorio, después de confirmar el contexto y namespace correctos, ejecuta una sola vez:

```powershell
kubectl config current-context
kubectl get namespace student-acofre
& .\k8s\universidad\configure-secrets.ps1
```

El script genera valores aleatorios URL-safe y los guarda en estos Secrets, sin imprimirlos ni escribirlos en archivos:

- `meloop-postgres-credentials`: `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`.
- `meloop-redis-credentials`: `REDIS_PASSWORD`.
- `meloop-auth-connection`: `DATABASE_URL`, `REDIS_URL`.

Si los tres ya existen, el script no cambia las credenciales. Si solo existe una parte, se detiene para evitar desincronizar passwords. No borres los Secrets ni vuelvas a generar credenciales con datos persistentes sin planificar una rotación.

## Desplegar las bases de datos

```powershell
kubectl apply -f .\k8s\universidad\postgres.yaml
kubectl apply -f .\k8s\universidad\redis.yaml
kubectl wait -n student-acofre --for=condition=Ready pod -l app.kubernetes.io/name=meloop-postgres --timeout=300s
kubectl wait -n student-acofre --for=condition=Ready pod -l app.kubernetes.io/name=meloop-redis --timeout=300s
kubectl get statefulsets,services,pods,pvc -n student-acofre
```

PostgreSQL y Redis usan un PVC cada uno, con `cinder-ceph`: 2 GiB para PostgreSQL y 1 GiB para Redis. Ambos Services son `ClusterIP`. Redis requiere contraseña y habilita AOF en su volumen.

Requests añadidos: PostgreSQL 100m CPU/256Mi, Redis 50m/64Mi. El Job de migraciones agrega temporalmente 25m/64Mi. Con los 50m/128Mi ya existentes, el máximo durante la migración es 225m CPU/512Mi, 5 Pods y 2 PVC por estos recursos; se mantiene bajo los límites indicados de 1 CPU, 1 GiB, 12 Pods y 3 PVC. El almacenamiento añadido suma 3 GiB de los 10 GiB disponibles. Verifica las cuotas efectivas antes de aplicar:

```powershell
kubectl get resourcequota -n student-acofre
kubectl describe resourcequota -n student-acofre
```

## Ejecutar las seis migraciones

Desde la raíz del repositorio:

```powershell
& .\k8s\universidad\run-migrations.ps1
kubectl wait -n student-acofre --for=condition=complete job/meloop-auth-migrations --timeout=600s
kubectl logs -n student-acofre job/meloop-auth-migrations
```

El script valida que haya exactamente seis archivos `.sql`, crea/actualiza el ConfigMap a partir de los archivos originales en `supabase\migrations`, elimina una ejecución previa del Job y crea el Job. Los nombres se ordenan lexicográficamente, que corresponde al orden de las marcas de tiempo de las migraciones. Si `kubectl wait` falla, el Job se detiene en el primer error; examina `kubectl logs -n student-acofre job/meloop-auth-migrations` antes de reintentar.

## Conectar Auth Service

La siguiente operación cambia únicamente las variables de entorno del Deployment ya existente `meloop-auth`; no recrea ni cambia el Gateway o el Ingress:

```powershell
kubectl set env deployment/meloop-auth --namespace student-acofre --from=secret/meloop-auth-connection
kubectl rollout status deployment/meloop-auth --namespace student-acofre --timeout=180s
kubectl logs deployment/meloop-auth --namespace student-acofre --since=10m | Select-String 'database_connected|redis_connected|database_connection_failed|redis_connection_failed'
```

En los logs deben aparecer `database_connected` y `redis_connected`, no sus variantes `*_failed`.

## Comprobar conexiones directamente

```powershell
kubectl exec -n student-acofre statefulset/meloop-postgres -- sh -c 'PGPASSWORD="$POSTGRES_PASSWORD" psql -h 127.0.0.1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT current_database(), version();"'
kubectl exec -n student-acofre statefulset/meloop-redis -- sh -c 'REDISCLI_AUTH="$REDIS_PASSWORD" redis-cli -h 127.0.0.1 ping'
```

PostgreSQL debe devolver `meloop` y la versión 16; Redis, `PONG`.

## Probar registro, login y validación a través del Gateway

El Ingress existente usa el host `meloop-acofre.dev.censei.cl` y enruta `/` al Gateway. No es necesario cambiarlo. Los endpoints de Auth conservan sus rutas: `POST /auth/register`, `POST /auth/login` y `GET /auth/validate`.

```powershell
$gatewayUrl = 'http://meloop-acofre.dev.censei.cl'
$suffix = [Guid]::NewGuid().ToString('N').Substring(0, 10)
$username = "k8s$suffix"
$email = "$username@example.com"
$password = "Meloop-$suffix-A1"

$registerBody = @{
    username = $username
    email    = $email
    password = $password
} | ConvertTo-Json
$registered = Invoke-RestMethod -Uri "$gatewayUrl/auth/register" -Method Post -ContentType 'application/json' -Body $registerBody
if (-not $registered.success) { throw "Registro fallido: $($registered | ConvertTo-Json -Depth 5)" }

$loginBody = @{
    email    = $email
    password = $password
} | ConvertTo-Json
$loggedIn = Invoke-RestMethod -Uri "$gatewayUrl/auth/login" -Method Post -ContentType 'application/json' -Body $loginBody
if (-not $loggedIn.success -or -not $loggedIn.data.session_token) { throw "Login fallido: $($loggedIn | ConvertTo-Json -Depth 5)" }

$validated = Invoke-RestMethod -Uri "$gatewayUrl/auth/validate" -Method Get -Headers @{
    Authorization = "Bearer $($loggedIn.data.session_token)"
}
if (-not $validated.success -or $validated.data.email -ne $email) { throw "Validación fallida: $($validated | ConvertTo-Json -Depth 5)" }

Write-Output "Registro, inicio de sesión y validación por Gateway OK para $email"
```

El nombre y correo de prueba son únicos para cada ejecución. Si el host no resuelve desde tu equipo, configura el DNS/hosts del equipo para apuntar a la dirección del Ingress; no cambies el manifiesto del Ingress para esta prueba.

# Prueba de Carga — Social Service (Operaciones Sociales)

> **Nota de alcance:** Esta prueba de carga evalúa específicamente las operaciones sociales (`GET /friends` y `GET /friends/suggestions`, incluyendo la validación del cálculo de `amigos_en_comun`). No representa la totalidad de las operaciones del sistema Meloop ni la totalidad del requerimiento no funcional RNF-01.

---

## 1. Configuración del Entorno de Prueba

Para garantizar aislamiento estricto y proteger el entorno principal de desarrollo y bases de datos externas, la prueba se ejecutó sobre un stack dedicado definido en `services/social-service/loadtest/docker-compose.yml`, con el nombre de proyecto `meloop-social-loadtest`.

### Componentes y Servicios

| Servicio | Imagen / Contexto | Puertos Internos / Host | Configuración Clave |
|---|---|---|---|
| **PostgreSQL** | `postgres:17-alpine` | `5432` / `127.0.0.1:15432` | DB: `meloop_social_loadtest`, volumen aislado `loadtest-postgres-data`, inicializado con migraciones de `supabase/migrations`. |
| **Redis** | `redis:7.2-alpine` | `6379` (red interna) | Almacén de sesiones para Auth Service. |
| **Auth Service** | Dockerfile de `auth-service` | `8083` / `127.0.0.1:18083` | Conexión directa a `postgres` y `redis` de la red de prueba. Modo `release`. |
| **Social Service** | Dockerfile de `social-service` | `8086` (red interna) | Modo `release`, conectado a PostgreSQL dedicado y Auth Service para validación de sesiones. |
| **API Gateway** | Dockerfile de `api-gateway` | `8080` / `127.0.0.1:18080` | Modo `release`, ruteo autenticado (`SessionAuth`) hacia Social Service y validación con Auth Service. |

### Red y Aislamiento

- **Red:** `meloop-social-loadtest_loadtest` (driver `bridge`).
- **Resolución DNS interna:** Cada contenedor cuenta con un alias de red explícito (`aliases: [postgres]`, `aliases: [redis]`, `aliases: [auth-service]`, `aliases: [social-service]`, `aliases: [api-gateway]`), asegurando conectividad entre microservicios sin colisionar con otros contenedores del host.

---

## 2. Preparación y Siembra de Datos

La preparación se implementó en `services/social-service/loadtest/seed.go`, garantizando idempotencia estricta y verificación de integridad previa a la carga:

1. **Usuarios sembrados:** 500 usuarios (`lt_social_000` a `lt_social_499`), registrados mediante la API de Auth Service (`POST /auth/register`). En reejecuciones, las cuentas existentes actualizan su hash bcrypt en la base de datos de prueba para aceptar la contraseña de la sesión actual de forma inmediata e idempotente.
2. **Amistades aceptadas:** 7.500 relaciones bidireccionales únicas (`estado = 'ACEPTADA'`), estructuradas en topología regular circular donde cada uno de los 500 usuarios posee exactamente **30 amigos directos** (15 vecinos en cada dirección del anillo).
3. **Preferencias musicales:** 5.000 registros insertados en `preferencia_musical` (10 preferencias por usuario: 2 géneros, 3 artistas, 5 canciones) para habilitar el motor de compatibilidad y ranking de sugerencias.
4. **Usuarios virtuales (VUs):** 200 usuarios con sesiones activas autenticadas generadas mediante `POST /auth/login`, con tokens de sesión inyectados en k6 para enviar la cabecera `Authorization: Bearer <token>`.
5. **Candidatos y amigos en común:** 
   - Para cada usuario evaluado existen 469 candidatos elegibles no amigos ($500 - 30 - 1 = 469$).
   - El candidato situado a distancia 16 en el anillo posee exactamente **15 amigos en común** con el usuario, valor validado determinísticamente tanto en preflight como durante el 100% de las iteraciones de la prueba.

---

## 3. Procedimiento Reproducible

La ejecución se automatiza a través de PowerShell y k6:

```powershell
$env:LOADTEST_USER_PASSWORD = ([guid]::NewGuid().ToString('N') + 'Z9!')
try {
    & powershell -NoProfile -ExecutionPolicy Bypass -File .\services\social-service\loadtest\run.ps1
    $result = $LASTEXITCODE
} finally {
    Remove-Item Env:\LOADTEST_USER_PASSWORD -ErrorAction SilentlyContinue
}
if ($null -eq $result) { exit 1 }
exit $result
```

### Fases de Ejecución en `run.ps1`

1. **Levantamiento del entorno:** `docker compose --project-name meloop-social-loadtest --file services/social-service/loadtest/docker-compose.yml up --detach --build --wait`.
2. **Sondeo de salud:** Espera hasta que `http://127.0.0.1:18080/health` responda HTTP 200.
3. **Siembra de datos:** Ejecución de `go run .\loadtest\seed.go` apuntando exclusivamente a `127.0.0.1:15432` y `127.0.0.1:18083`.
4. **Preflight funcional:** Validación funcional antes de someter a carga:
   - `GET /friends`: Verifica éxito HTTP 200 y exactamente 30 amistades coincidentes con el fixture.
   - `GET /friends/suggestions`: Verifica éxito HTTP 200, 469 candidatos únicos y el valor esperado de `amigos_en_comun` (15) para el candidato testigo.
5. **Ejecución de carga con k6:** 200 VUs concurrentes durante 2 minutos ejecutando alternadamente `GET /friends` y `GET /friends/suggestions` a través de `http://127.0.0.1:18080`.

---

## 4. Resultados Reales de la Prueba

### Parámetros de Carga

| Métrica de Ejecución | Valor Real |
|---|---|
| **Usuarios Virtuales (VUs)** | 200 concurrentes constantes |
| **Duración de la prueba** | 2 minutos (2m 02.5s con drenado) |
| **Iteraciones completadas** | 2.831 |
| **Total de peticiones HTTP** | 5.662 (2.831 a `/friends` y 2.831 a `/friends/suggestions`) |
| **Throughput promedio** | 46,21 req/s (23,10 iteraciones/s) |
| **Datos transferidos** | 326 MB recibidos, 974 kB enviados |

### Métricas de Rendimiento y Latencia

| Endpoint / Métrica | Promedio (avg) | Mediana (p50) | p90 | p95 | Mínimo | Máximo | Umbral Objetivo | Estado |
|---|---|---|---|---|---|---|---|---|
| **Global (`http_req_duration`)** | **4.273,88 ms** | 2.245,78 ms | 8.758,69 ms | **9.387,86 ms (9.38s)** | 210,99 ms | 11.307,62 ms | `p(95) < 500ms` | ❌ Incumplido |
| **`GET /friends`** | 948,75 ms | 848,95 ms | 1.473,89 ms | **1.697,69 ms (1.70s)** | 210,99 ms | 2.588,50 ms | `p(95) < 500ms` | ❌ Incumplido |
| **`GET /friends/suggestions`** | 7.598,00 ms | 7.381,44 ms | 9.388,00 ms | **10.032,68 ms (10.03s)** | 1.603,53 ms | 11.307,62 ms | `p(95) < 500ms` | ❌ Incumplido |
| **Tasa de errores (`http_req_failed`)** | **0,00%** | — | — | — | — | — | `< 1%` | ✅ **Cumplido** |

### Verificaciones Funcionales (Checks k6)

| Validación | Total Evaluaciones | Aprobadas | Falladas | Tasa de Éxito |
|---|---|---|---|---|
| `friends returns HTTP 200` | 2.831 | 2.831 | 0 | 100,00% |
| `friends returns a successful JSON envelope` | 2.831 | 2.831 | 0 | 100,00% |
| `friends match the seeded identities and accepted state` | 2.831 | 2.831 | 0 | 100,00% |
| `suggestions returns HTTP 200` | 2.831 | 2.831 | 0 | 100,00% |
| `suggestions returns a successful JSON envelope` | 2.831 | 2.831 | 0 | 100,00% |
| `suggestions contain unique eligible candidates and valid mutual counts` | 2.831 | 2.831 | 0 | 100,00% |
| **Total Checks** | **16.986** | **16.986** | **0** | **100,00%** |
| **Errores funcionales (`functional_error_rate`)** | 5.662 | 0 fallos | 0 | **0,00%** |

---

## 5. Problemas Encontrados y Soluciones Aplicadas

### Problema 1: Fallo de resolución de nombres DNS entre contenedores (Causa del Exit Code 1 original)
- **Diagnóstico:** Al cambiar la red Compose de `internal` a `bridge`, los contenedores no recreados mantuvieron la conexión sin los alias de servicio (`postgres`, `redis`, `auth-service`, `social-service`). Como consecuencia, `social-service` no pudo resolver `postgres:5432` y finalizó con `database_connection_failed`; `auth-service` quedó con el pool de base de datos en `nil` retornando HTTP 500 al sembrador.
- **Solución:** Se definieron alias de red explícitos en `services/social-service/loadtest/docker-compose.yml` (`loadtest.aliases: [service_name]`), permitiendo que Compose aplique la resolución DNS inequívoca en la red bridge.

### Problema 2: Envoltorio de array en PowerShell 5.1 al deserializar sesiones
- **Diagnóstico:** En Windows PowerShell 5.1, la construcción `@(Get-Content -Raw | ConvertFrom-Json)` sobre un arreglo JSON retorna un arreglo de un único elemento que contiene el `Object[]` real. La aserción `$sessions.Count -ne 200` evaluaba `$sessions.Count == 1` y lanzaba una excepción tras completar exitosamente la siembra.
- **Solución:** Se ajustó la asignación en `run.ps1` a `$sessions = Get-Content ... | ConvertFrom-Json` normalizando la colección a `System.Array` sin doble anidación.

### Problema 3: Idempotencia de usuarios con contraseñas dinámicas entre ejecuciones
- **Diagnóstico:** Si la base de datos ya contenía los 500 usuarios creados en una ejecución previa con otra contraseña aleatoria, los intentos posteriores recibían HTTP 409 Conflict en el registro pero fallaban con HTTP 401 Unauthorized al autenticar los 200 VUs.
- **Solución:** En `seed.go`, ante una respuesta HTTP 409 Conflict, se ejecuta una consulta parametrizada que actualiza `contrasena_hash` con el hash bcrypt precomputado de la ejecución actual, permitiendo que la siembra sea 100% reutilizable e idempotente sin necesidad de eliminar volúmenes ni recrear tablas.

---

## 6. Análisis del Rendimiento y Conclusión de Cumplimiento

### Cumplimiento de Umbrales

| Criterio | Objetivo | Resultado Obtenido | Veredicto |
|---|---|---|---|
| **Tasa de fallos HTTP** | `< 1,00%` | **0,00%** (0 fallos en 5.662 peticiones) | **CUMPLE** |
| **Latencia p95** | `< 500 ms` | **9.387,86 ms** (9.38 segundos) | **NO CUMPLE** |

### Evidencia del Cuello de Botella en Latencia

1. **Rendimiento de `GET /friends` (p95 = 1.697 ms):**
   - Aunque la consulta SQL a `amistad` es directa, la alta concurrencia de 200 VUs compartiendo el pool de conexiones de PostgreSQL y la validación por petición a Auth Service a través del Gateway eleva la latencia por encima de los 500 ms requeridos.
2. **Rendimiento de `GET /friends/suggestions` (p95 = 10.032 ms):**
   - El endpoint de sugerencias ejecuta múltiples operaciones complejas por petición:
     - Escaneo completo de candidatos elegibles excluyendo amigos, bloqueos y moderaciones (`GetEligibleCandidates`).
     - Cálculo de amigos en común en tiempo real mediante CTE sobre la tabla `amistad` (`GetMutualFriendsCount`) para 469 candidatos por petición.
     - Extracción en lote de perfiles musicales para 470 identidades desde `preferencia_musical` (`GetBatchMusicalProfiles`).
     - Cálculo en memoria del puntaje de compatibilidad y ranking completo de los 469 candidatos.
   - Sometido a 200 VUs simultáneas sin capa de caché ni paginación en el cálculo de sugerencias, la base de datos y el microservicio sufren saturación de CPU y contención de hilos, generando tiempos de respuesta en el orden de 7 a 11 segundos.

### Veredicto Final

- **Estado de la suite de prueba:** **COMPLETADO**. El entorno aislado, la preparación de datos, la validación de integridad y la ejecución de carga de 2 minutos se completaron con un 100% de éxito funcional (cero errores HTTP y cero discrepancias de negocio).
- **Cumplimiento de objetivos RNF-01 (operaciones sociales):** **INCUMPLIMIENTO PARCIAL**. La disponibilidad y consistencia funcional son impecables (0% errores), pero la latencia p95 bajo 200 VUs concurrentes supera ampliamente el límite de 500 ms, especialmente en el motor de sugerencias de amistad.

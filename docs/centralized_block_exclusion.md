# Centralización de la Exclusión de Usuarios Bloqueados

## 1. Contexto y Objetivo
En la plataforma Meloop, las reglas de bloqueo son de naturaleza **bidireccional**: si un usuario $A$ bloquea a un usuario $B$, o si $B$ bloquea a $A$, ninguno de los dos debe poder ver publicaciones, perfiles o recomendaciones del otro.

Para evitar que múltiples microservicios (`post-service`, `user-service`, `recommendation-service`) dupliquen consultas directas a la tabla `BLOQUEO` o implementen reglas divergentes, `social-service` actúa como la **única fuente de verdad** mediante una función y un endpoint interno reutilizables.

---

## 2. Ubicación y Firmas de Implementación

### 2.1 Capa de Repositorio (`FriendshipRepository`)
- **Archivo:** [`services/social-service/repositories/friendship_repository.go`](file:///c:/Users/alanp/Repositorios/Meloop/services/social-service/repositories/friendship_repository.go)
- **Firma:**
  ```go
  ExcludeBlocked(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
  GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
  ```
- **Consulta SQL optimizada (consulta única bidireccional):**
  ```sql
  SELECT DISTINCT
      CASE 
          WHEN id_usuario_bloqueador::text = $1 THEN id_usuario_bloqueado::text
          ELSE id_usuario_bloqueador::text
      END AS blocked_candidate_id
  FROM bloqueo
  WHERE (id_usuario_bloqueador::text = $1 AND id_usuario_bloqueado::text = ANY($2))
     OR (id_usuario_bloqueado::text = $1 AND id_usuario_bloqueador::text = ANY($2))
  ```

### 2.2 Capa de Servicio (`FriendshipService`)
- **Archivo:** [`services/social-service/services/friendship_service.go`](file:///c:/Users/alanp/Repositorios/Meloop/services/social-service/services/friendship_service.go)
- **Firma conceptual y de servicio:**
  ```go
  ExcludeBlocked(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
  GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
  ```
- **Función pura auxiliar:**
  ```go
  FilterBlockedCandidates(candidateIDs []string, blockedIDs []string) []string
  ```

### 2.3 Cliente HTTP Reutilizable para Consumidores Go
- **Archivo:** [`services/social-service/client/client.go`](file:///c:/Users/alanp/Repositorios/Meloop/services/social-service/client/client.go)
- **Firma:**
  ```go
  type SocialClient interface {
      ExcludeBlocked(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
      GetBlockedCandidateIDs(ctx context.Context, userID string, candidateIDs []string) ([]string, error)
      GetExcludedAndAllowed(ctx context.Context, userID string, candidateIDs []string) (allowed []string, excluded []string, err error)
  }
  ```

---

## 3. Contrato del Endpoint Interno HTTP/REST

Para permitir que otros microservicios consuman esta lógica sin necesidad de importar directamente los paquetes de base de datos de `social-service`:

### Endpoint
- **Método:** `POST`
- **Ruta interna:** `/internal/blocks/exclude` (diseñado para comunicación servicio a servicio, sin necesidad de sesión JWT de usuario final)
- **Ruta alternativa protegida:** `/friends/blocks/exclude` (utiliza el header `X-User-ID` si `user_id` no se incluye en el cuerpo)
- **Headers:** `Content-Type: application/json`

### Solicitud (Request Payload)
```json
{
  "user_id": "usr-12345",
  "candidate_ids": [
    "usr-cand-1",
    "usr-cand-2",
    "usr-cand-3"
  ]
}
```
*Campos soportados:* `user_id` (o `id_usuario`), y `candidate_ids` (o `candidatos`).

### Respuesta Exitosa (`200 OK`)
```json
{
  "success": true,
  "data": {
    "candidate_ids": [
      "usr-cand-1",
      "usr-cand-3"
    ],
    "excluded_ids": [
      "usr-cand-2"
    ]
  },
  "error": null
}
```
- `candidate_ids`: Lista de candidatos conservados (libres de relaciones de bloqueo con `user_id`).
- `excluded_ids`: Lista de candidatos identificados que fueron excluidos debido a un bloqueo mutuo o unilateral.

### Respuestas de Error
- **`400 Bad Request`**: Si el payload JSON es inválido o si `user_id` está ausente/vacío.
  ```json
  {
    "success": false,
    "data": null,
    "error": {
      "code": "BAD_REQUEST",
      "message": "El identificador del usuario de referencia (user_id) es requerido"
    }
  }
  ```
- **`500 Internal Server Error`**: Si ocurre un error inesperado de conectividad con la base de datos.

---

## 4. Semántica del Bloqueo Bidireccional
La lógica aplica estrictamente las siguientes reglas:
1. **Unilateral Saliente ($U \to C$):** Si el usuario de referencia $U$ bloqueó al candidato $C$, $C$ es excluido.
2. **Unilateral Entrante ($C \to U$):** Si el candidato $C$ bloqueó al usuario de referencia $U$, $C$ es excluido.
3. **Bloqueo Mutuo ($U \leftrightarrow C$):** Si ambos se bloquearon mutuamente, $C$ es excluido (sin duplicados en la respuesta).
4. **Sin Bloqueo:** Si no existe registro en `BLOQUEO` entre $U$ y $C$, $C$ se conserva en su posición relativa.
5. **Ausencia de falsos positivos:** Si $C$ o $U$ tienen bloqueos con terceros no involucrados en la consulta, estos no afectan el resultado.
6. **Inmutabilidad:** El slice `candidateIDs` original nunca es modificado ni rebanado in-place; se asigna y retorna una nueva estructura limpia.
7. **Manejo de vacíos:** Si `candidateIDs` está vacío, se retorna inmediatamente un slice vacío `[]` sin realizar llamadas a la base de datos ni provocar errores.

---

## 5. Guía de Integración para los Servicios Consumidores

### 5.1 `post-service`
- **Caso de uso:** Exclusión de publicaciones cuyos autores tienen una relación de bloqueo con el usuario que consulta el feed o realiza búsquedas (`requesterID`).
- **Integración recomendada:**
  1. El controlador o servicio obtiene los posts candidatos desde la base de datos.
  2. Extrae los IDs de autor únicos (`authorIDs`).
  3. Realiza una petición `POST` a `http://social-service:8084/internal/blocks/exclude` enviando `user_id = requesterID` y `candidate_ids = authorIDs`.
  4. Conserva únicamente los posts cuyos autores pertenezcan a `data.candidate_ids` (o descarta aquellos en `data.excluded_ids`).
- **Ejemplo en Go:**
  ```go
  socialURL := os.Getenv("SOCIAL_SERVICE_URL") // Ej: "http://social-service:8084"
  client := client.NewSocialClient(socialURL)
  allowedAuthors, err := client.ExcludeBlocked(ctx, requesterID, authorIDs)
  if err != nil {
      // Manejar error o fallback defensivo
  }
  ```

### 5.2 `user-service`
- **Caso de uso:** Visualización o búsqueda de perfiles públicos. Un usuario bloqueador o bloqueado no debe poder ver perfiles protegidos ni aparecer en búsquedas mutuas.
- **Integración recomendada:**
  1. Al recibir una búsqueda de usuarios (`search`), recopilar los IDs de candidatos devueltos por la consulta base.
  2. Invocar `/internal/blocks/exclude` con el ID del usuario autenticado y los IDs candidatos.
  3. Descartar los usuarios excluidos antes de construir la respuesta de perfiles hacia el cliente.

### 5.3 `recommendation-service`
- **Caso de uso:** Generación de sugerencias de amigos y cálculo de contenido popular (`profile_repository.go`).
- **Integración recomendada:**
  - Actualmente `FilterEligibleFriendCandidates` consulta directamente la tabla `BLOQUEO`.
  - Para centralizar esta regla, `recommendation-service` puede desacoplar el acceso a la tabla `BLOQUEO` invocando el cliente HTTP de `social-service`:
    ```go
    allowedCandidates, err := socialClient.ExcludeBlocked(ctx, userID, candidateUserIDs)
    ```
  - Esto garantiza que cualquier cambio futuro en políticas de bloqueo (como bloqueos temporales, suspensiones u otros estados) se actualice únicamente en `social-service`.

---

## 6. Consideraciones y Limitaciones

1. **Latencia de Red:** La comunicación HTTP entre microservicios agrega latencia de red (usualmente 1-5ms en la red interna de Docker/Kubernetes). Se recomienda configurar timeouts de 3 a 5 segundos y timeouts por contexto (`context.WithTimeout`).
2. **Tamaño del Lote:** Se recomienda enviar lotes de hasta 500 candidatos por invocación. Para conjuntos masivos, procesar en bloques (chunking).
3. **Resiliencia / Fallback:** En caso de caída de `social-service`, los servicios consumidores pueden decidir registrar el fallo con logger estructurado y aplicar una política de degradación segura (por ejemplo, excluir preventivamente si hay sospechas o registrar alerta de infraestructura).

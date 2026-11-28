# Checklist RN-17

| Endpoint | ¿Excluye bloqueados? | ¿Excluye inactivos? | Estado |
|---|---|---|---|
| `GET /api/v1/posts/:id` | Sí, en SQL (404) | No | Implementado; prueba de integración pendiente |
| `GET /api/v1/posts/author/:authorId` | Sí, en SQL antes de paginar | No | Implementado; prueba de integración pendiente |
| `GET /api/v1/posts/search` | Sí, en SQL antes de paginar | No | Implementado; prueba de integración pendiente |
| `GET /v1/feed` (candidatos y detalles) | Sí, en SQL antes del límite/página | Sí | Implementado; prueba de integración pendiente |
| `GET /api/v1/posts/:id/comments` y respuestas | Sí, en SQL | No | Implementado; prueba de integración pendiente |
| `POST /api/v1/posts/:id/comments` | Sí, responde 403 `BLOCKED` | No | Implementado; prueba de integración pendiente |
| `POST /api/v1/posts/:id/comments/:commentId/replies` | Sí, responde 403 `BLOCKED` | No | Implementado; prueba de integración pendiente |
| `POST /posts/:postId/likes` y `POST /v1/posts/:id/like` | Sí, responde 403 `BLOCKED` | No | Implementado; prueba de integración pendiente |
| `POST /v1/me/saved` (`POST /api/v1/posts/:id/save`) | Sí, responde 403 `BLOCKED` | No | Implementado; prueba de integración pendiente |
| `POST /v1/posts/:id/share` (`POST /api/v1/posts/:id/share`) | Sí, responde 403 `BLOCKED` | No | Implementado; prueba de integración pendiente |
| `GET /v1/me/saved` (`GET /api/v1/users/me/saved`) | Sí, en SQL antes de paginar | No | Implementado; prueba de integración pendiente |
| `DELETE /v1/me/saved/:id` | N/A (quitar un guardado) | No | Implementado |

La lista de bloqueados se obtiene mediante `GetBlockedCandidateIDs` de `social-service` (contrato equivalente a `BlockedWith`) y se cachea en Redis durante 60 segundos. La prueba cruzada con dos usuarios requiere PostgreSQL, Redis y `social-service` disponibles.

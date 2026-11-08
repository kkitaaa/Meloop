# Documentación de Meloop

- [Contrato REST OpenAPI del Gateway](./api/openapi.yaml)
- [Estado de la API WebSocket de Chat](./chat-ws.md)
- [Guía del API Gateway y su inventario de prefijos](../services/api-gateway/README.md)

El contrato lista las operaciones que existen en los servicios detrás del
Gateway y señala explícitamente los dominios todavía pendientes.

## Swagger UI local

Con Docker Compose disponible, iniciar el Gateway, sus dependencias configuradas
y la UI de desarrollo con:

```sh
docker compose --profile dev up api-gateway swagger-ui
```

Abrir [http://localhost:8090](http://localhost:8090). La especificación se
monta desde el archivo fuente del Gateway. El perfil `dev` evita publicar la UI
como parte de un despliegue normal. Algunas rutas siguen sin destino en Compose
(Social y Recomendaciones) y no pueden probarse con esa configuración.

## Validar el contrato

Desde la raíz del repositorio:

```sh
make lint-api
```

# Capa de consumo de API

La comunicación con el backend se centraliza en `lib/core/network`.

- `api_config.dart` define la URL del API Gateway. Su valor por defecto es
  `http://localhost:8080`. Para otro entorno se usa, por ejemplo,
  `flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8080`.
- `api_client.dart` expone `get`, `post`, `put` y `delete`. Serializa JSON y
  transforma errores HTTP, de conexión, tiempo de espera y formato en
  `ApiException`.
- `api_response.dart` representa el contrato común del Gateway: `success`,
  `data` y `error`.

Cada funcionalidad conserva sus modelos y servicios bajo `lib/features`. El
ejemplo `features/users` contiene `UserModel` y `UserService`, que consume
`GET /users` y `POST /users` a través del API Gateway. Las pantallas deben usar
los servicios, no el cliente HTTP directamente.

Para añadir una nueva funcionalidad, crear `features/<feature>/models` y
`features/<feature>/services`, inyectar un `ApiClient` en el servicio y convertir
la propiedad `data` de la respuesta al modelo correspondiente.

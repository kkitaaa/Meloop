# Investigación de integración con Spotify Web API

## 1. Objetivo

El objetivo de esta investigación es evaluar la viabilidad técnica de integrar Spotify Web API con Meloop para permitir la búsqueda y referencia de contenido musical dentro de la plataforma.

La investigación contempla:

* Método de autenticación entre el backend de Meloop y Spotify.
* Obtención y manejo de access tokens.
* Restricciones, cuotas y rate limits de Spotify Web API.
* Endpoints necesarios para búsqueda de contenido musical.
* Información que deberá entregar el backend al frontend.
* Uso de identificadores y URI de Spotify para integrar contenido reproducible.
* Desarrollo de una prueba de concepto (PoC) en Go que valide la comunicación con Spotify Web API.

---

## 2. Viabilidad de la integración

La integración de Spotify Web API con Meloop es técnicamente viable.

Spotify proporciona una API REST que permite consultar información de su catálogo musical, incluyendo artistas, canciones y álbumes. Para Meloop, estas funcionalidades permitirían buscar contenido en Spotify desde el backend y entregar al frontend únicamente la información necesaria para representar dicho contenido.

La arquitectura propuesta es:

```text
Frontend Meloop
      |
      v
Backend Meloop
      |
      v
Spotify Web API
```

El frontend no debe comunicarse con Spotify utilizando las credenciales privadas de la aplicación. El `Client Secret` debe permanecer exclusivamente en el backend y almacenarse mediante variables de entorno.

### 2.1. Restricción de Development Mode

Durante la investigación se identificó una restricción relevante de la plataforma de Spotify.

Las aplicaciones nuevas comienzan en **Development Mode**. Según las condiciones actuales de Spotify, la cuenta propietaria de una aplicación en Development Mode debe disponer de una suscripción Spotify Premium para que la aplicación funcione.

Esto fue comprobado durante el desarrollo del PoC: inicialmente Web API no se encontraba disponible para la cuenta utilizada y, después de disponer de Spotify Premium, fue posible habilitar Web API y crear la aplicación necesaria para realizar las pruebas.

Development Mode también posee restricciones adicionales de uso y está orientado principalmente a aplicaciones en construcción, pruebas y acceso limitado.

Esta condición debe considerarse una dependencia externa del proyecto. Si Meloop continuara utilizando Spotify en etapas posteriores, sería necesario revisar nuevamente las políticas, cuotas y modalidades de acceso vigentes antes de una eventual puesta en producción.

---

## 3. Autenticación

Para las operaciones investigadas se utiliza **Client Credentials Flow**.

Este mecanismo está diseñado para comunicación servidor a servidor donde se autentica la aplicación y no un usuario individual.

Es apropiado para el PoC de Meloop porque las operaciones investigadas corresponden a consultas del catálogo de Spotify y no requieren acceder a información privada de una cuenta de usuario.

Las credenciales requeridas son:

```text
SPOTIFY_CLIENT_ID
SPOTIFY_CLIENT_SECRET
```

Estas credenciales deben almacenarse mediante variables de entorno y nunca deben incluirse directamente en el código fuente ni enviarse al frontend.

### 3.1. Flujo de autenticación

El proceso es el siguiente:

1. El backend obtiene `Client ID` y `Client Secret` desde variables de entorno.
2. Las credenciales son utilizadas mediante HTTP Basic Authentication.
3. El backend realiza una petición `POST` al servicio de autenticación de Spotify.
4. La petición utiliza `grant_type=client_credentials`.
5. Spotify valida las credenciales.
6. Spotify devuelve un `access_token`.
7. El backend utiliza dicho token como Bearer Token para realizar solicitudes a Spotify Web API.

El endpoint utilizado para obtener el token es:

```http
POST https://accounts.spotify.com/api/token
```

La petición utiliza:

```text
Content-Type: application/x-www-form-urlencoded
grant_type=client_credentials
```

junto con las credenciales de la aplicación mediante HTTP Basic Authentication.

Ejemplo conceptual de la respuesta:

```json
{
  "access_token": "<token>",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

El campo `expires_in` indica durante cuántos segundos será válido el access token.

### 3.2. Limitaciones de Client Credentials

Client Credentials autentica a la aplicación y no a un usuario.

Por este motivo, este mecanismo no permite acceder a endpoints que requieran información privada o autorización de una cuenta de Spotify.

Si Meloop necesitara posteriormente acceder a información o realizar acciones en representación de un usuario, sería necesario investigar otro mecanismo de OAuth, como Authorization Code o Authorization Code con PKCE.

---

## 4. Manejo y renovación del access token

Los access tokens obtenidos mediante Client Credentials tienen una duración limitada indicada por el campo `expires_in`.

Client Credentials no proporciona un `refresh_token`. Por este motivo, cuando el access token expire o esté próximo a expirar, el backend debe solicitar uno nuevo utilizando nuevamente el `Client ID` y el `Client Secret`.

Para una implementación definitiva en Meloop se recomienda almacenar internamente:

```text
AccessToken
ExpiresAt
```

El backend debería reutilizar el token mientras continúe siendo válido, evitando solicitar un nuevo access token para cada petición realizada a Spotify.

También se recomienda utilizar un pequeño margen antes de la expiración. Por ejemplo, Meloop podría considerar un token inválido 60 segundos antes de `ExpiresAt`.

El flujo propuesto sería:

```text
Petición que requiere Spotify
          |
          v
¿Existe un token válido?
       /       \
     Sí         No
     |           |
     |     Solicitar nuevo token
     |           |
     +-----------+
          |
          v
Realizar petición a Spotify
```

Esta estrategia disminuye solicitudes innecesarias al endpoint de autenticación y reduce la posibilidad de intentar utilizar un token mientras está próximo a expirar.

---

## 5. Rate limits y cuotas

Spotify Web API aplica mecanismos para limitar la cantidad de solicitudes realizadas por las aplicaciones.

Es importante diferenciar dos conceptos:

### 5.1. Rate limit

El rate limit general de Spotify Web API se calcula utilizando una ventana móvil de 30 segundos.

Si una aplicación realiza demasiadas solicitudes dentro de este período, Spotify puede responder:

```http
HTTP 429 Too Many Requests
```

Spotify puede indicar mediante sus respuestas cuándo una aplicación debe reducir la frecuencia de solicitudes.

Meloop no debería asumir una cantidad fija de peticiones por minuto, ya que los límites pueden depender del modo de acceso y de políticas vigentes de Spotify.

### 5.2. Development Mode quota

Las aplicaciones en Development Mode también están sujetas a restricciones de cuota adicionales.

Estas cuotas son diferentes del rate limit general de la Web API. Spotify puede igualmente utilizar una respuesta HTTP `429 Too Many Requests` cuando una cuota ha sido excedida y actualmente puede incluir información que permite distinguir un exceso de cuota de un rate limit.

Por esta razón, una implementación definitiva debe inspeccionar correctamente la respuesta entregada por Spotify antes de decidir cómo realizar un reintento.

### 5.3. Estrategias recomendadas

Para disminuir el consumo de Spotify Web API se recomienda:

* Evitar peticiones duplicadas.
* Implementar caché para búsquedas repetidas.
* Reutilizar los access tokens mientras sean válidos.
* Manejar correctamente respuestas HTTP 429.
* Evitar reintentos inmediatos e ilimitados.
* Implementar backoff o espera cuando corresponda.
* Registrar métricas sobre consumo y errores de la integración.

Meloop ya contempla Redis dentro de su infraestructura, por lo que una implementación futura podría utilizarlo para almacenar temporalmente resultados frecuentes de búsquedas de Spotify.

Un posible flujo sería:

```text
Usuario busca contenido
        |
        v
Backend Meloop
        |
        v
¿Resultado disponible en caché?
      /     \
    Sí       No
    |         |
    |    Spotify Web API
    |         |
    |      Guardar caché
    |         |
    +---------+
        |
        v
Respuesta al frontend
```

---

## 6. Búsqueda de contenido

Para el PoC se investigó el endpoint de búsqueda de Spotify:

```http
GET https://api.spotify.com/v1/search
```

Los principales parámetros utilizados son:

* `q`: término que se desea buscar.
* `type`: tipo de recurso que se desea obtener.
* `limit`: cantidad máxima de resultados solicitados.
* `offset`: permite paginar resultados cuando sea necesario.

Spotify permite buscar distintos tipos de contenido, incluyendo artistas, canciones y álbumes.

Ejemplo conceptual utilizado por el PoC:

```http
GET /v1/search?q=Daft%20Punk&type=artist&limit=1
Authorization: Bearer <access_token>
```

Para Meloop, este endpoint permitiría inicialmente realizar búsquedas de artistas y posteriormente extender la integración a canciones y álbumes.

La documentación actual de Spotify limita el número máximo de resultados obtenidos mediante `limit`, por lo que una implementación que requiera una cantidad mayor deberá utilizar paginación.

---

## 7. Estructura de datos propuesta para el frontend

No se recomienda enviar directamente al frontend toda la respuesta obtenida desde Spotify.

El backend de Meloop debería transformar la respuesta externa a una estructura propia o DTO (Data Transfer Object), exponiendo únicamente la información que necesita la interfaz.

Esto evita acoplar directamente el frontend a la estructura de Spotify Web API.

### 7.1. Artista

Ejemplo de estructura propuesta:

```json
{
  "id": "spotify_artist_id",
  "name": "Daft Punk",
  "spotify_uri": "spotify:artist:spotify_artist_id",
  "spotify_url": "https://open.spotify.com/artist/spotify_artist_id",
  "image_url": "https://i.scdn.co/image/..."
}
```

### 7.2. Canción

Para una futura búsqueda de canciones se propone:

```json
{
  "id": "spotify_track_id",
  "name": "Nombre de la canción",
  "artist": "Nombre del artista",
  "album": "Nombre del álbum",
  "spotify_uri": "spotify:track:spotify_track_id",
  "spotify_url": "https://open.spotify.com/track/spotify_track_id",
  "image_url": "https://i.scdn.co/image/..."
}
```

El uso de DTOs propios permite que Meloop controle el contrato existente entre backend y frontend.

Si Spotify modifica determinados campos de sus respuestas, dichos cambios podrían manejarse dentro de la capa de integración sin necesariamente modificar el frontend.

---

## 8. Spotify ID, URI y contenido reproducible

Spotify identifica sus recursos mediante identificadores únicos.

Durante el PoC se obtuvo, por ejemplo:

```text
Spotify ID: 4tZwfgrHOc3mvqYlEYSvVi
Spotify URI: spotify:artist:4tZwfgrHOc3mvqYlEYSvVi
```

También se obtuvo la URL pública correspondiente:

```text
https://open.spotify.com/artist/4tZwfgrHOc3mvqYlEYSvVi
```

Meloop puede transportar estos identificadores dentro de sus respuestas para permitir que el frontend identifique contenido de Spotify.

### 8.1. Spotify Embeds

Para una interfaz web, Spotify proporciona mecanismos oficiales para incorporar contenido interactivo mediante **Spotify Embeds**.

Los Embeds permiten incorporar contenido de Spotify dentro de una aplicación web sin que Meloop tenga que obtener o distribuir directamente los archivos de audio.

### 8.2. iFrame API

Spotify también proporciona una iFrame API que permite crear e interactuar programáticamente con Embeds.

Esta API puede recibir como referencia tanto un Spotify URI como una URL de Spotify.

Por ejemplo:

```text
spotify:track:<spotify_track_id>
```

Por lo tanto, la arquitectura recomendada sería:

```text
Spotify Web API
      |
      | búsqueda y metadatos
      v
Backend Meloop
      |
      | ID / URI / URL
      v
Frontend Meloop
      |
      v
Spotify Embed / iFrame API
```

Spotify Web API se utilizaría para localizar y obtener información sobre el contenido, mientras que los mecanismos oficiales de Embed se utilizarían para representar contenido reproducible cuando corresponda.

---

## 9. Prueba de concepto en Go

Como parte de la investigación se desarrolló una prueba de concepto independiente en:

```text
backend/poc/Spotify/
```

La estructura utilizada es:

```text
Spotify/
├── .env
├── .env.example
├── go.mod
├── go.sum
└── main.go
```

El archivo `.env` contiene las credenciales reales utilizadas durante el desarrollo y debe permanecer excluido del repositorio Git.

El archivo `.env.example` documenta las variables requeridas sin incluir información sensible:

```env
SPOTIFY_CLIENT_ID=
SPOTIFY_CLIENT_SECRET=
```

### 9.1. Funcionamiento del PoC

La prueba realiza las siguientes operaciones:

1. Lee `SPOTIFY_CLIENT_ID` y `SPOTIFY_CLIENT_SECRET` desde variables de entorno.
2. Construye la autenticación necesaria para Client Credentials.
3. Solicita un access token a Spotify.
4. Utiliza el token para consultar Spotify Web API.
5. Realiza una búsqueda de artista mediante `/v1/search`.
6. Procesa la respuesta JSON.
7. Obtiene el primer resultado.
8. Muestra el nombre, Spotify ID, Spotify URI y Spotify URL.

El flujo completo probado es:

```text
.env
 |
 v
Client ID + Client Secret
 |
 v
POST /api/token
 |
 v
Access Token
 |
 v
GET /v1/search
 |
 v
Respuesta JSON
 |
 v
ID + nombre + URI + URL
```

### 9.2. Ejecución

Para probar la integración se realizó una búsqueda de Daft Punk:

```bash
go run . "Daft Punk"
```

La ejecución produjo exitosamente:

```text
Buscando artista: Daft Punk

Artista encontrado
-------------------
Nombre: Daft Punk
Spotify ID: 4tZwfgrHOc3mvqYlEYSvVi
Spotify URI: spotify:artist:4tZwfgrHOc3mvqYlEYSvVi
Spotify URL: https://open.spotify.com/artist/4tZwfgrHOc3mvqYlEYSvVi
```

El resultado demuestra experimentalmente que es posible:

* Autenticar el backend mediante Client Credentials.
* Obtener un access token válido.
* Comunicarse con Spotify Web API.
* Ejecutar una búsqueda sobre el catálogo.
* Procesar la respuesta JSON.
* Recuperar el identificador de Spotify correspondiente al contenido encontrado.

Por lo tanto, el PoC cumple con el objetivo de validar técnicamente la comunicación básica entre Meloop y Spotify Web API.

---

## 10. Propuesta para una implementación futura

El código desarrollado corresponde exclusivamente a una prueba de concepto y no debería considerarse directamente una implementación productiva.

Si la integración es incorporada definitivamente a Meloop, se recomienda separar responsabilidades.

Una posible estructura sería:

```text
spotify-service/
├── config/
│   └── config.go
├── clients/
│   └── spotify_client.go
├── controllers/
│   └── spotify_controller.go
├── models/
│   └── spotify.go
├── services/
│   └── spotify_service.go
└── main.go
```

Las responsabilidades podrían dividirse de la siguiente manera:

### Spotify Client

Responsable de:

* Comunicación HTTP con Spotify.
* Obtención de access tokens.
* Envío de Bearer Tokens.
* Manejo de códigos HTTP.
* Manejo de rate limits y cuotas.

### Spotify Service

Responsable de:

* Lógica de búsqueda.
* Transformación de respuestas de Spotify.
* Construcción de DTOs.
* Integración con caché.

### Controller

Responsable de exponer endpoints propios de Meloop, por ejemplo:

```http
GET /spotify/search?q=Daft%20Punk&type=artist
```

El frontend consumiría el endpoint de Meloop en lugar de comunicarse directamente con Spotify Web API.

---

## 11. Consideraciones de seguridad

La integración debe cumplir como mínimo las siguientes medidas:

* Nunca almacenar `Client Secret` directamente en el código.
* Mantener las credenciales en variables de entorno.
* Excluir archivos `.env` del repositorio Git.
* Proporcionar `.env.example` sin credenciales reales.
* No enviar `Client Secret` al frontend.
* No registrar credenciales en logs.
* No registrar access tokens completos en logs.
* Manejar correctamente errores HTTP provenientes de Spotify.
* Implementar timeouts para las solicitudes externas.
* Implementar una estrategia controlada de reintentos.
* Evitar reintentos ilimitados ante HTTP 429.
* Reutilizar access tokens mientras sean válidos.
* Revisar periódicamente cambios en las políticas de Spotify Developer.

---

## 12. Limitaciones identificadas

La investigación permitió identificar las siguientes limitaciones:

1. Client Credentials no permite acceder a recursos privados de usuarios.
2. Los access tokens tienen una duración limitada.
3. La aplicación debe manejar rate limits y cuotas.
4. Development Mode posee restricciones adicionales.
5. Actualmente, el propietario de una aplicación en Development Mode debe disponer de Spotify Premium.
6. Las condiciones de Spotify Developer pueden cambiar y deben revisarse antes de una eventual puesta en producción.
7. Una aplicación destinada a una audiencia amplia requeriría evaluar las condiciones de acceso y cuota correspondientes de Spotify.
8. La reproducción debe realizarse utilizando mecanismos permitidos por Spotify, como sus Embeds, en lugar de intentar distribuir directamente contenido de audio.

---

## 13. Conclusión

La investigación y la prueba de concepto demuestran que la integración entre Meloop y Spotify Web API es técnicamente viable para operaciones de consulta del catálogo musical.

Client Credentials permite al backend autenticarse contra Spotify sin requerir autorización individual de un usuario para las operaciones públicas investigadas.

La prueba desarrollada en Go confirmó exitosamente la obtención de un access token y la búsqueda de un artista mediante Spotify Web API, obteniendo como resultado su nombre, Spotify ID, Spotify URI y URL pública.

Para una implementación definitiva se recomienda encapsular la comunicación con Spotify dentro del backend de Meloop e incorporar manejo automático de expiración de tokens, caché, tratamiento de rate limits y cuotas, timeouts, manejo de errores y DTOs propios para la comunicación con el frontend.

La utilización de Spotify ID y Spotify URI permite además conectar los resultados obtenidos mediante Web API con mecanismos oficiales como Spotify Embeds y la iFrame API.

Sin embargo, la integración presenta dependencias externas importantes, especialmente las restricciones actuales de Development Mode y el requisito de Spotify Premium para la cuenta propietaria de la aplicación.

Por este motivo, la integración puede considerarse técnicamente viable para el desarrollo y las pruebas de Meloop, pero las políticas y modalidades de acceso de Spotify deberán volver a evaluarse antes de una eventual implementación productiva.

---

## 14. Referencias

La investigación se realizó utilizando principalmente la documentación oficial de Spotify for Developers.

* Spotify for Developers — Web API
  https://developer.spotify.com/documentation/web-api

* Spotify for Developers — Authorization
  https://developer.spotify.com/documentation/web-api/concepts/authorization

* Spotify for Developers — Client Credentials Flow
  https://developer.spotify.com/documentation/web-api/tutorials/client-credentials-flow

* Spotify for Developers — Search for Item
  https://developer.spotify.com/documentation/web-api/reference/search

* Spotify for Developers — Rate Limits
  https://developer.spotify.com/documentation/web-api/concepts/rate-limits

* Spotify for Developers — Quota Modes
  https://developer.spotify.com/documentation/web-api/concepts/quota-modes

* Spotify for Developers — February 2026 Web API Development Mode Migration Guide
  https://developer.spotify.com/documentation/web-api/tutorials/february-2026-migration-guide

* Spotify for Developers — Web API quota updates for Development Mode (23 de julio de 2026)
  https://developer.spotify.com/blog/2026-07-23-web-api-quota-updates

* Spotify for Developers — Embeds
  https://developer.spotify.com/documentation/embeds

* Spotify for Developers — Using the iFrame API
  https://developer.spotify.com/documentation/embeds/tutorials/using-the-iframe-api

* Spotify for Developers — iFrame API Reference
  https://developer.spotify.com/documentation/embeds/references/iframe-api

* Spotify for Developers — oEmbed API
  https://developer.spotify.com/documentation/embeds/reference/oembed

**Fecha de consulta:** 31 de agosto de 2026.

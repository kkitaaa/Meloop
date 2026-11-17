# Meloop ML Service

## Propósito

`ml-service` es una API independiente en Python que recibe perfiles y candidatos
preparados por otro componente y devuelve puntuaciones de similitud y
recomendaciones explicables. Proporciona:

- Compatibilidad entre perfiles de usuarios para ordenar candidatos a amistad.
- Similitud de contenido entre el perfil musical de un usuario y un catálogo
  enviado en la petición.
- Un endpoint de demostración (`/predict`) que genera sugerencias de ejemplo
  usando el ID y el número de interacciones.

El servicio **no** administra cuentas, autenticación, amistades, publicaciones ni
el catálogo musical. Tampoco consulta una base de datos, no obtiene candidatos
por su cuenta, no decide reglas de negocio ni aplica filtros de privacidad o
bloqueos. El backend debe validar permisos, construir el perfil y proporcionar
los candidatos permitidos. El servicio no entrena un modelo durante el
arranque: procesa las entradas recibidas y aplica un modelo de similitud.

> **Alcance del modelo:** el código actual no define los pesos de una
> especificación denominada “ML-12”. Las características de género, artista y
> canción tienen el mismo peso. No se debe interpretar el resultado como si ya
> aplicara ponderaciones ML-12.

## Cómo correrlo

### Requisitos

- Python 3.11 o posterior.
- pip.
- Opcional: Docker y Docker Compose para ejecutarlo en contenedor.

### Entorno virtual

Desde la raíz del repositorio, crea y activa un entorno virtual. En PowerShell:

```powershell
cd ml-service
python -m venv .venv
.\.venv\Scripts\Activate.ps1
python -m pip install --upgrade pip
pip install -r requirements-dev.txt
```

En Linux o macOS:

```bash
cd ml-service
python3 -m venv .venv
source .venv/bin/activate
python -m pip install --upgrade pip
pip install -r requirements-dev.txt
```

`requirements-dev.txt` instala las dependencias de ejecución, pytest, HTTPX y
Ruff. Para ejecutar solamente la API, también se puede instalar
`requirements.txt`.

### Iniciar la API localmente

Desde el directorio `ml-service`, con el entorno virtual activado:

```bash
uvicorn app.main:app --reload --host 127.0.0.1 --port 8001
```

La API queda disponible en `http://127.0.0.1:8001`; la documentación
interactiva de FastAPI está en `http://127.0.0.1:8001/docs`. Para comprobar que
el servicio está listo:

```bash
curl http://127.0.0.1:8001/ready
```

En PowerShell se puede usar `Invoke-RestMethod`:

```powershell
Invoke-RestMethod http://127.0.0.1:8001/ready
```

### Docker Compose

Desde la raíz del repositorio:

```bash
docker compose up --build -d ml-service
docker compose ps ml-service
```

La API se publica en `http://localhost:8001`. Para detener el contenedor:

```bash
docker compose stop ml-service
```

No se necesita levantar el resto de Meloop para probar la API de ML. Compose
incluye un healthcheck contra `/ready`.

## Variables de entorno

| Variable | Obligatoria | Valor predeterminado | Descripción |
|---|---:|---|---|
| `LOG_LEVEL` | No | `INFO` | Nivel del log estructurado: `DEBUG`, `INFO`, `WARNING`, `ERROR` o `CRITICAL`. |

El servicio no requiere credenciales ni acceso a una base de datos para iniciar.
Las preferencias y candidatos se reciben en cada petición. En la configuración
actual de Docker Compose no se sobrescribe `LOG_LEVEL`.

## Endpoints

Todos los endpoints aceptan y devuelven JSON cuando corresponde. Los ejemplos y
respuestas siguientes se obtuvieron ejecutando la aplicación actual con
FastAPI TestClient.

### `GET /`

Información breve del servicio y ruta de la documentación interactiva.

```http
GET /
```

Respuesta `200`:

```json
{
  "message": "Bienvenido al servicio de ML de Meloop",
  "docs": "/docs"
}
```

### `GET /health`

Comprueba que el proceso responde. No garantiza por sí solo que el modelo esté
listo para procesar solicitudes.

```http
GET /health
```

Respuesta `200`:

```json
{
  "status": "ok",
  "service": "ml-service"
}
```

### `GET /ready`

Comprueba que el modelo y el pipeline de preferencias están inicializados.
Devuelve `200` cuando está listo y `503` si no lo está.

```http
GET /ready
```

Respuesta `200`:

```json
{
  "status": "ready",
  "service": "ml-service",
  "model_loaded": true,
  "data_pipeline_loaded": true,
  "model_version": "1.0.0"
}
```

### `POST /predict`

Endpoint de demostración y compatibilidad con payloads básicos. No compara los
perfiles musicales ni consulta un catálogo: con este código genera IDs y
puntuaciones de ejemplo en función del `user_id` y la cantidad de
interacciones. Para resultados reales de compatibilidad musical usa
`/recommendations/music`; para compatibilidad entre usuarios usa
`/recommendations/friends`.

Petición:

```json
{
  "user_id": 42,
  "limit": 2,
  "interactions": [
    {
      "type": "like",
      "target_id": 101
    }
  ]
}
```

Respuesta `200`:

```json
{
  "user_id": 42,
  "recommendations": [
    {
      "item_id": 53,
      "score": 0.91,
      "reason": "Sugerencia adaptada a la actividad reciente del usuario"
    },
    {
      "item_id": 54,
      "score": 0.81,
      "reason": "Sugerencia adaptada a la actividad reciente del usuario"
    }
  ],
  "model": "profile_compatibility",
  "model_version": "1.0.0",
  "interaction_count": 1
}
```

### `POST /recommendations/friends`

Compara el perfil musical/social del usuario con los perfiles candidatos
enviados. Excluye al propio usuario de los candidatos y ordena por puntuación
descendente; en caso de empate, usa el ID ascendente. `limit` está entre 1 y 50
y tiene valor predeterminado 10.

Petición:

```json
{
  "user_id": 42,
  "limit": 2,
  "profile": {
    "genres": ["rock"],
    "artists": ["Arctic Monkeys"],
    "songs": []
  },
  "candidate_profiles": [
    {
      "user_id": 7,
      "profile": {
        "genres": ["rock"],
        "artists": ["Radiohead"],
        "songs": []
      }
    },
    {
      "user_id": 8,
      "profile": {
        "genres": ["jazz"],
        "artists": ["Miles Davis"],
        "songs": []
      }
    }
  ]
}
```

Respuesta `200`:

```json
{
  "user_id": 42,
  "recommendations": [
    {
      "item_id": 7,
      "score": 0.5,
      "reason": "Ambos escuchan a Rock"
    },
    {
      "item_id": 8,
      "score": 0.0,
      "reason": "No se encontraron preferencias musicales o sociales en común"
    }
  ],
  "model": "profile_compatibility",
  "model_version": "1.0.0",
  "interaction_count": 0
}
```

La puntuación está entre 0 y 1. Si el usuario no tiene preferencias musicales,
la respuesta es `200` con `recommendations: []`.

### `POST /recommendations/music`

Ordena canciones o artistas del catálogo enviado según su similitud de
contenido con el perfil musical del usuario. El catálogo es obligatorio y debe
contener al menos un elemento. `limit` está entre 1 y 50, con valor
predeterminado 10.

Petición:

```json
{
  "user_id": 42,
  "limit": 2,
  "profile": {
    "genres": ["rock"],
    "artists": ["Arctic Monkeys"],
    "songs": []
  },
  "catalog": [
    {
      "item_id": 101,
      "profile": {
        "genres": ["rock"],
        "artists": ["Arctic Monkeys"],
        "songs": []
      }
    },
    {
      "item_id": 102,
      "profile": {
        "genres": ["jazz"],
        "artists": ["Miles Davis"],
        "songs": []
      }
    }
  ]
}
```

Respuesta `200`:

```json
{
  "user_id": 42,
  "recommendations": [
    {
      "item_id": 101,
      "score": 1.0,
      "reason": "Coincide con tus preferencias: Arctic Monkeys, Rock"
    },
    {
      "item_id": 102,
      "score": 0.0,
      "reason": "No se encontraron preferencias musicales en común"
    }
  ],
  "model": "music_content_similarity",
  "model_version": "1.0.0",
  "interaction_count": 0
}
```

### Validación de solicitudes

Los IDs de usuario, candidatos y catálogo deben ser enteros positivos. Los
tipos de interacción admitidos son `like`, `friend_added`, `comment` y
`post_interaction`; cada interacción requiere un `target_id` positivo. Los
campos de perfil (`genres`, `artists`, `songs`) son listas de texto opcionales.
Una petición que no cumple el esquema, incluyendo un catálogo vacío para música,
recibe `422 Unprocessable Entity` con los detalles de validación.

### Esquema OpenAPI

FastAPI publica el esquema de ejecución en `/openapi.json` y la interfaz
interactiva en `/docs`. La copia versionada del esquema exportado está en
[`docs/ml-service-openapi.json`](../docs/ml-service-openapi.json).

Para regenerarla desde el directorio `ml-service`, usando las dependencias
instaladas:

```bash
python -c "import json; from pathlib import Path; from app.main import app; Path('../docs/ml-service-openapi.json').write_text(json.dumps(app.openapi(), ensure_ascii=False, indent=2) + '\n', encoding='utf-8')"
```

Regenera el archivo si cambian rutas, modelos de entrada/salida o metadatos de
FastAPI.

## Cómo funciona el modelo

### Preparación de perfiles

`UserPreferencesPipeline` limpia los valores de género, artista y canción:
normaliza Unicode, elimina espacios sobrantes, convierte texto a minúsculas y
quita duplicados. Después representa cada preferencia como una característica
binaria, por ejemplo `genre:rock` o `artist:arctic monkeys`. Las características
se alinean entre el perfil y los candidatos para que las posiciones
comparadas representen el mismo dato.

### Compatibilidad de perfiles y contenido

El modelo compara los vectores binarios mediante **similitud coseno**. En
términos simples, mide qué proporción de las preferencias representadas apunta
en la misma dirección entre dos perfiles. Un resultado cercano a `1` indica
muchas preferencias en común; `0` significa que no comparten ninguna de las
características representadas. Para amistades la puntuación se expresa como
porcentaje en la lógica interna y se devuelve normalizada de `0` a `1`. Para
música se ordena el catálogo y se devuelve la puntuación coseno, también de `0`
a `1`.

La respuesta incluye una explicación basada en las características coincidentes
detectadas. Si no hay coincidencias, devuelve una razón explícita de que no se
encontraron preferencias en común.

### Pesos y alcance de ML-12

En la implementación actual, cada característica de género, artista y canción
se codifica como `0` o `1` y recibe el mismo peso. No hay una configuración ni
una constante que defina ponderaciones de ML-12 en el servicio, así que las
puntuaciones actuales **no aplican pesos diferenciados** por tipo de dato. Si
ML-12 define pesos de negocio, deben especificarse e implementarse antes de
presentar esta salida como una recomendación ponderada.

Las interacciones se cuentan y se incluyen en la respuesta; en el endpoint
musical no alteran todavía el cálculo de similitud. El score mide semejanza de
características, no probabilidad de que una persona escuche o prefiera un
elemento.

## Cómo ejecutar las pruebas

Desde la raíz del repositorio, con dependencias instaladas:

```bash
make test-ml
```

En una instalación de Windows con MinGW, el equivalente es:

```powershell
mingw32-make test-ml
```

O desde el directorio `ml-service`:

```bash
python -m pytest
```

Las pruebas cubren el procesamiento de preferencias, los cálculos del modelo,
los endpoints con `TestClient`, los esquemas de entrada y los rechazos `422`,
además de la demostración reproducible con perfiles sintéticos. Para ejecutar
formato y lint de forma manual:

```bash
ruff format --check .
ruff check .
```

## Problemas frecuentes

| Problema | Causa probable | Solución |
|---|---|---|
| `ModuleNotFoundError` al iniciar | Dependencias instaladas fuera del entorno virtual o entorno no activado. | Activa `.venv` y ejecuta `pip install -r requirements-dev.txt` desde `ml-service`. |
| `uvicorn` no se reconoce | No se instaló `requirements.txt` o el entorno virtual no está activo. | Activa el entorno y reinstala las dependencias; inicia con `python -m uvicorn app.main:app --reload --host 127.0.0.1 --port 8001`. |
| El puerto 8001 está ocupado | Otro proceso ya usa el puerto. | Detén ese proceso o elige otro puerto en el comando de Uvicorn y ajusta la URL usada para probarlo. |
| La API devuelve `422` | El JSON no cumple el esquema, un ID no es positivo, el tipo de interacción no está permitido o el catálogo musical está vacío. | Revisa la respuesta `detail`, el esquema en `/docs` y los ejemplos de esta guía. |
| `/health` responde pero el healthcheck falla | `/health` solo comprueba que el proceso vive; el healthcheck de Compose verifica `/ready`. | Consulta `GET /ready` y revisa los logs del contenedor con `docker compose logs ml-service`. |
| `docker compose up` no encuentra archivos o servicio | El comando se ejecutó fuera de la raíz del repositorio o Docker Desktop no está iniciado. | Cambia a la raíz de Meloop e inicia Docker Desktop antes de levantar el servicio. |
| Cambió la API pero el JSON documentado no | El archivo OpenAPI versionado no se regeneró. | Ejecuta el comando de exportación de OpenAPI descrito arriba y revisa el diff. |
| No aparecen candidatos con perfil vacío | No hay características que comparar; el servicio devuelve una lista vacía. | Envía preferencias válidas en el perfil y, para amigos, una lista de candidatos. |

## Arquitectura

El diagrama de llamadas está en
[`docs/ml-service-architecture.md`](../docs/ml-service-architecture.md). El
cliente habitual del servicio es `recommendation-service`, que prepara los
perfiles y candidatos, llama por HTTP y decide cómo aplicar las reglas de
negocio del producto.

# ML Service

Servicio independiente de machine learning para Meloop.

## Requisitos

- Python 3.11+
- Entorno virtual

## Crear entorno virtual

```bash
cd ml-service
python -m venv .venv
source .venv/bin/activate  # Linux/macOS
.venv\Scripts\activate     # Windows PowerShell
```

## Instalar dependencias

```bash
pip install -r requirements.txt
```

## Ejecutar la API

```bash
uvicorn app.main:app --reload --host 0.0.0.0 --port 8001
```

## Endpoints base

- GET `/` - información general
- GET `/health` - estado del servicio
- POST `/predict` - endpoint de prueba para recibir payloads del backend
- POST `/recommendations/friends` - devuelve candidatos ordenados por compatibilidad musical

### Recomendaciones de amigos

`POST /recommendations/friends` recibe el `user_id`, el perfil musical del usuario y los
perfiles candidatos. La respuesta incluye como máximo `limit` candidatos, su puntuación de
compatibilidad y una razón legible basada en las preferencias compartidas. FastAPI publica
el contrato interactivo en `/docs` y el esquema OpenAPI en `/openapi.json`.

Ejemplo de petición:

```json
{
	"user_id": 42,
	"limit": 5,
	"profile": {
		"genres": ["rock"],
		"artists": ["Arctic Monkeys"],
		"songs": []
	},
	"candidate_profiles": [
		{
			"user_id": 7,
			"profile": {"genres": ["rock"], "artists": ["Radiohead"], "songs": []}
		}
	]
}
```

Si el usuario no tiene géneros, artistas, canciones o preferencias, el servicio responde
`200` con `recommendations: []` y no inventa candidatos. Los IDs deben ser positivos y
`limit` está restringido al intervalo 1-50.

## Estructura

```text
ml-service/
├── app/
│   ├── models/
│   ├── services/
│   ├── schemas/
│   ├── __init__.py
│   └── main.py
├── requirements.txt
├── README.md
└── .venv/
```

## Preparación para recomendaciones

La estructura está pensada para ir agregando:

- modelos de recomendación en `app/models/`
- lógica de procesamiento en `app/services/`
- validaciones de entrada/salida en `app/schemas/`
- comunicación con el backend mediante JSON en `/predict` u otros endpoints futuros

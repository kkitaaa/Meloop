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

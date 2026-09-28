# 5. ML Service (Python)

En este documento se explica cómo configurar e instalar el servicio de Machine Learning en Python.

**Tiempo estimado:** 15-20 minutos

---

## ML Service Overview

**Tecnología:** Python 3.11 + FastAPI  
**Puerto:** 8000  
**Dependencias principales:**
- FastAPI 0.141.1
- Uvicorn 0.52.3
- Pandas 3.0.5
- NumPy 2.3.5
- scikit-learn 1.9.0

---

## 🐍 Instalación del Entorno Python

### Paso 1: Verificar Python 3.11

```bash
python --version
# Output: Python 3.11.x

# En Linux, puede ser python3.11
python3.11 --version
```

### Paso 2: Crear entorno virtual

Es **altamente recomendado** crear un entorno virtual para aislar las dependencias:

```bash
cd meloop/ml-service

# Windows
python -m venv venv
.\venv\Scripts\Activate.ps1

# macOS/Linux
python3.11 -m venv venv
source venv/bin/activate
```

**Verificar activación:**
```bash
# El prompt debería mostrar: (venv) C:\...\ml-service>
# o
# (venv) user@machine ml-service %
```

### Paso 3: Actualizar pip

```bash
python -m pip install --upgrade pip

# Verificar
pip --version
```

---

## 📦 Instalar Dependencias

### Opción A: Dependencias de Producción

```bash
pip install -r requirements.txt
```

**Archivo requirements.txt:**
```
fastapi==0.141.1
uvicorn==0.52.3
pandas==3.0.5
numpy==2.3.5
scikit-learn==1.9.0
```

### Opción B: Dependencias de Desarrollo

Si necesitas también las herramientas de testing y linting:

```bash
pip install -r requirements.txt
pip install -r requirements-dev.txt
```

**Archivo requirements-dev.txt:**
```
pytest==8.4.2
httpx==0.28.1
ruff==0.13.2
black==24.x.x
```

### Verificar instalación

```bash
pip list

# Salida esperada:
# fastapi                       0.141.1
# uvicorn                       0.52.3
# pandas                        3.0.5
# numpy                         2.3.5
# scikit-learn                  1.9.0
```

---

## 🚀 Ejecutar el Servicio

### Opción A: Desarrollo (con auto-reload)

```bash
# Asegúrate que el entorno virtual está activado
# (venv) debería aparecer en el prompt

# Ejecutar con auto-reload
python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Salida esperada:
# INFO:     Uvicorn running on http://0.0.0.0:8000
# INFO:     Application startup complete
```

### Opción B: Producción (sin auto-reload)

```bash
python -m uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
```

### Opción C: Usar Docker (Recomendado)

```bash
cd meloop

# Compilar imagen (si no existe)
docker-compose build ml-service

# Ejecutar
docker-compose up ml-service

# Ver logs
docker-compose logs -f ml-service
```

---

## Estructura de Carpetas - ML Service

```
ml-service/
├── app/
│   ├── __init__.py
│   ├── main.py              # Entrada principal (FastAPI)
│   ├── logging_config.py    # Configuración de logs
│   ├── models/              # Modelos de ML
│   │   ├── __init__.py
│   │   └── [...modelos].py
│   ├── schemas/             # Esquemas Pydantic (request/response)
│   │   ├── __init__.py
│   │   └── [...esquemas].py
│   └── services/            # Lógica de negocio
│       ├── __init__.py
│       └── [...servicios].py
├── tests/
│   ├── __init__.py
│   ├── test_api.py         # Tests de API
│   └── [...tests].py
├── requirements.txt
├── requirements-dev.txt
├── pyproject.toml
├── pytest.ini
└── README.md
```

---

## Archivo main.py Básico

```python
# app/main.py
import logging
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from app.logging_config import setup_logging

# Configurar logging
setup_logging()
logger = logging.getLogger(__name__)

app = FastAPI(
    title=\"Meloop ML Service\",
    description=\"Servicio de Machine Learning para Meloop\",
    version=\"1.0.0\",
    docs_url=\"/docs\",
    redoc_url=\"/redoc\"
)

@app.get(\"/health\")
async def health_check():
    \"\"\"Health check endpoint\"\"\"
    return JSONResponse(
        status_code=200,
        content={\"status\": \"ok\", \"service\": \"ml-service\"}
    )

@app.get(\"/\")
async def root():
    \"\"\"Root endpoint\"\"\"
    return {
        \"message\": \"Meloop ML Service\",
        \"docs\": \"/docs\",
        \"health\": \"/health\"
    }

if __name__ == \"__main__\":
    import uvicorn
    uvicorn.run(
        app,
        host=\"0.0.0.0\",
        port=8000,
        log_level=\"info\"
    )
```

---

## 🧪 Testing

### Ejecutar tests

```bash
# Activar entorno virtual primero
# (venv) ...

# Ejecutar todos los tests
pytest

# Tests con cobertura
pytest --cov=app

# Ver logs durante tests
pytest -v -s

# Tests de un archivo específico
pytest tests/test_api.py
```

### Ejemplo de test básico

```python
# tests/test_api.py
import pytest
from httpx import AsyncClient
from app.main import app

@pytest.mark.asyncio
async def test_health_check():
    \"\"\"Test del endpoint /health\"\"\"
    async with AsyncClient(app=app, base_url=\"http://test\") as client:
        response = await client.get(\"/health\")
        assert response.status_code == 200
        assert response.json()[\"status\"] == \"ok\"

@pytest.mark.asyncio
async def test_root():
    \"\"\"Test del endpoint root\"\"\"
    async with AsyncClient(app=app, base_url=\"http://test\") as client:
        response = await client.get(\"/\")
        assert response.status_code == 200
        assert \"message\" in response.json()
```

---

## 🔍 Validar Instalación

### Paso 1: Health check

```bash
# El servicio debe estar ejecutándose
curl http://localhost:8000/health

# Salida esperada:
# {\"status\":\"ok\",\"service\":\"ml-service\"}
```

### Paso 2: Documentación interactiva

Abre en tu navegador: **http://localhost:8000/docs**

Verás una interfaz Swagger donde puedes:
- Ver todos los endpoints
- Probarlos interactivamente
- Ver esquemas de solicitud/respuesta

### Paso 3: Test desde Python

```python
import requests

response = requests.get('http://localhost:8000/health')
print(response.json())
# Output: {'status': 'ok', 'service': 'ml-service'}
```

---

## Linting y Formato

### Verificar con Ruff

```bash
# Checar código
ruff check app/

# Reparar automáticamente
ruff check --fix app/
```

### Formatear con Black

```bash
# Verificar formato
black --check app/

# Aplicar formato
black app/
```

### En una línea (verificación completa)

```bash
ruff check app/ && black --check app/ && pytest
```

---

## Integración con otros Servicios

### Llamar a otros servicios desde ML

```python
# app/services/api_client.py
import aiohttp
import logging

logger = logging.getLogger(__name__)

class APIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url
    
    async def get_user_data(self, user_id: str):
        \"\"\"Obtener datos del usuario desde User Service\"\"\"
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(
                    f\"{self.base_url}/users/{user_id}\"
                ) as resp:
                    if resp.status == 200:
                        return await resp.json()
                    else:
                        logger.error(f\"Error: {resp.status}\")
                        return None
        except Exception as e:
            logger.error(f\"Connection error: {e}\")
            return None

# Uso en endpoint
from fastapi import FastAPI

api_client = APIClient(\"http://localhost:8082\")

@app.get(\"/predict/{user_id}\")
async def predict(user_id: str):
    user_data = await api_client.get_user_data(user_id)
    if not user_data:
        return {\"error\": \"Cannot fetch user data\"}
    # Hacer predicción...
    return {\"prediction\": \"...\"}
```

---

## Problemas Comunes

### "No module named 'app'"

```bash
# Verifica que estás en el directorio ml-service
cd meloop/ml-service

# Verifica que app/ existe
ls app/

# Verifica que el entorno virtual está activado
# (venv) debería aparecer en el prompt
```

### Port 8000 already in use

```bash
# Windows
netstat -ano | findstr :8000
taskkill /PID <PID> /F

# macOS/Linux
lsof -i :8000
kill -9 <PID>
```

### "python: No module named uvicorn"

```bash
# Verifica que el entorno virtual está activado
# (venv) ...

# Reinstala dependencias
pip install -r requirements.txt
```

### Entorno virtual no se activa

```bash
# Windows - PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Luego intenta de nuevo
.\venv\Scripts\Activate.ps1

# Si aún falla, usa CMD
cmd
venv\Scripts\activate.bat
```

---

## 🚀 Próximo Paso

Ahora que tienes el ML Service ejecutándose:

👉 **Ve a [6. Flutter App →](06-flutter-setup.md)**

---

## Checklist de ML Service

```
□ Python 3.11 instalado
□ Entorno virtual creado y activado
□ Dependencias instaladas: pip install -r requirements.txt
□ Estructura de carpetas validada
□ main.py existe y es válido
□ Servicio ejecutándose en :8000
□ Health check responde correctamente
□ Documentación Swagger accesible en /docs
□ Tests pasan: pytest
□ Linting pasa: ruff check
□ Código formateado: black
```

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0

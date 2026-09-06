# 7. Inicio Rápido

Resumen ejecutivo de comandos para poner el proyecto en marcha rápidamente.

**Tiempo total:** ~15 minutos (si todo ya está instalado)

---

## El Comando Mágico (Todo en 5 comandos)

Si ya tienes todas las herramientas instaladas:

```bash
# 1. Clonar repositorio
git clone https://github.com/tuorganizacion/meloop.git
cd meloop

# 2. Configurar entorno
cp .env.example .env
# EDITA .env con tus credenciales de Supabase

# 3. Levantar servicios Docker
docker-compose up -d

# 4. Ejecutar backend (en terminal nueva)
cd backend
npm install
npm run start:dev

# 5. Ejecutar Flutter (en terminal nueva)
cd app/flutter_app
flutter pub get
flutter run -d chrome
```

---

## Checklist Rápido de Instalación

**Tiempo:** 5 minutos

```
□ Git instalado
□ Go 1.26.6 instalado
□ Python 3.11 instalado
□ Node.js 18+ LTS instalado
□ Flutter 3.13.0+ instalado
□ Docker Desktop ejecutándose
```

---

## 🚀 Paso a Paso - 10 Minutos

### 1. Clonar y Configurar (1 min)

```bash
git clone https://github.com/tuorganizacion/meloop.git
cd meloop

# Copiar configuración
cp .env.example .env

# EDITAR .env con:
# - SUPABASE_URL
# - SUPABASE_KEY
# - Credenciales de MinIO
```

### 2. Levantar Infraestructura (3 min)

```bash
# Terminal 1: Levantar Docker
docker-compose up -d

# Validar
docker-compose ps
```

### 3. Backend NestJS (2 min)

```bash
# Terminal 2: Backend
cd backend
npm install
npm run start:dev

# Esperar a ver:
# [Nest] xxxxx - 09/01/2024, xx:xx:xx AM LOG [NestApplication] Listening on port 3000
```

### 4. Frontend Flutter (2 min)

```bash
# Terminal 3: Flutter
cd app/flutter_app
flutter pub get
flutter run -d chrome

# Esperar a que se abra navegador en http://localhost:5000
```

### 5. ML Service (opcional) (2 min)

```bash
# Terminal 4: ML Service
cd ml-service
python -m venv venv

# Windows
.\venv\Scripts\Activate.ps1

# macOS/Linux
source venv/bin/activate

pip install -r requirements.txt
python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

---

## Validación Rápida

Una vez todo está ejecutándose:

```bash
# 1. API Gateway
curl http://localhost:8080/health
# Esperado: {"status":"ok"}

# 2. Backend NestJS
curl http://localhost:3000/health
# Esperado: {"status":"ok"}

# 3. ML Service
curl http://localhost:8000/health
# Esperado: {"status":"ok","service":"ml-service"}

# 4. Redis
docker exec -it meloop-redis-1 redis-cli -a <PASSWORD> PING
# Esperado: PONG

# 5. RabbitMQ Dashboard
# Abre en navegador: http://localhost:15672
# Usuario: guest, Contraseña: guest

# 6. MinIO Console
# Abre en navegador: http://localhost:9001
# Usuario: minioadmin, Contraseña: minioadmin

# 7. Flutter App
# Ya debería estar abierto en http://localhost:5000
```

---

## 📱 Accesos Rápidos

| Servicio | URL | Credenciales |
|----------|-----|--------------|
| **API Gateway** | http://localhost:8080 | - |
| **Backend NestJS** | http://localhost:3000 | - |
| **ML Service** | http://localhost:8000 | - |
| **Swagger ML** | http://localhost:8000/docs | - |
| **Flutter App** | http://localhost:5000 | - |
| **RabbitMQ** | http://localhost:15672 | guest / guest |
| **MinIO** | http://localhost:9001 | minioadmin / minioadmin |

---

## 🛑 Detener Servicios

```bash
# Presionar Ctrl+C en cada terminal

# O, si ejecutaste con -d (detached):
docker-compose stop

# Detener todo
docker-compose down
```

---

## 🔄 Reiniciar Rápido

```bash
# Detener y limpiar
docker-compose down

# Reiniciar
docker-compose up -d

# En terminales existentes, presiona Ctrl+C y re-ejecuta:
cd backend && npm run start:dev
cd app/flutter_app && flutter run -d chrome
cd ml-service && python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

---

## Comandos Útiles por Sección

### Docker

```bash
docker-compose ps                    # Ver estado de servicios
docker-compose logs -f               # Ver logs en tiempo real
docker-compose logs -f redis         # Logs de servicio específico
docker-compose restart               # Reiniciar servicios
docker-compose down                  # Detener y eliminar contenedores
docker volume ls                     # Ver volúmenes
```

### Backend NestJS

```bash
npm install                          # Instalar dependencias
npm run start:dev                    # Ejecutar en desarrollo
npm run build                        # Compilar
npm run test                         # Tests
npm run lint                         # Linting
npm run format                       # Formater código
```

### ML Service (Python)

```bash
python -m venv venv                  # Crear entorno virtual
source venv/bin/activate             # Activar (Linux)
.\venv\Scripts\Activate.ps1          # Activar (Windows)
pip install -r requirements.txt      # Instalar dependencias
python -m uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
pytest                               # Tests
ruff check app/                      # Linting
black app/                           # Formater
```

### Flutter

```bash
flutter pub get                      # Obtener dependencias
flutter run -d chrome                # Ejecutar en web
flutter run -d windows               # Ejecutar en Windows
flutter run -d linux                 # Ejecutar en Linux
flutter test                         # Tests
flutter analyze                      # Análisis de código
flutter build web                    # Compilar para web
```

### Git

```bash
git status                           # Ver estado
git add .                            # Agregar cambios
git commit -m \"mensaje\"             # Commitear
git push                             # Subir cambios
git pull                             # Descargar cambios
```

---

## 🆘 Si Algo Falla

1. **Consulta [Troubleshooting](08-troubleshooting.md)**
2. **Verifica los logs:**
   ```bash
   docker-compose logs <servicio>
   ```
3. **Intenta limpiar y reiniciar:**
   ```bash
   docker-compose down -v
   docker-compose up -d
   ```

---

## Próximos Pasos Después de Setup

1. ✅ **Verificar que todo funciona** → Sigue los pasos 1-5 anterior
2. 📖 **Leer documentación de arquitectura** → `docs/architecture.md`
3. 🐛 **Familiarizarse con los logs** → Sigue `docs/logging.md`
4. 🧪 **Ejecutar tests** → `npm run test`, `flutter test`, `pytest`
5. 📝 **Revisar guía de estilo** → `docs/style-guide.md`
6. 🚀 **Empezar a desarrollar** → ¡Diviértete!

---

## 💡 Tips Útiles

### Desarrollo más rápido con hot reload

- **Flutter:** Presiona `r` en terminal para hot reload
- **NestJS:** npm run start:dev automáticamente reinicia en cambios
- **Python:** --reload en uvicorn automáticamente reinicia

### Mantener múltiples terminales organizadas

```bash
# En Windows (PowerShell)
# Abre nuevas tabs con Ctrl+Shift+T

# En macOS/Linux
# Usa tmux o screen para múltiples sesiones
tmux new-session -s meloop
tmux new-window -t meloop
tmux select-window -t meloop:0
```

### Monitorear uso de recursos

```bash
docker stats

# Output:
# CONTAINER ID  NAME           CPU %   MEM USAGE
# abc123        meloop-redis   0.1%    45MiB
```

---

## 🚀 Siguiente Lectura

- **Problemas frecuentes:** [Troubleshooting →](08-troubleshooting.md)
- **Arquitectura detallada:** [docs/architecture.md](../architecture.md)
- **Guía de estilos:** [docs/style-guide.md](../style-guide.md)

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0

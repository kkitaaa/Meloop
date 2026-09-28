# 8. Troubleshooting - Solución de Problemas Frecuentes

Guía completa de solución de problemas comunes durante el setup y ejecución.

---

## 🔍 Diagnóstico Inicial

Antes de resolver un problema, ejecuta este checklist:

```bash
# 1. Ver estado de servicios Docker
docker-compose ps

# 2. Ver logs
docker-compose logs

# 3. Verificar puerto disponible
netstat -ano | findstr :8080      # Windows
lsof -i :8080                      # macOS/Linux

# 4. Verificar conectividad
ping localhost
curl http://localhost:8080/health
```

---

## Índice de Problemas

- [Git & Repositorio](#git--repositorio)
- [Herramientas](#herramientas)
- [Docker](#docker)
- [Backend (NestJS)](#backend-nestjs)
- [Servicios Go](#servicios-go)
- [ML Service (Python)](#ml-service-python)
- [Flutter](#flutter)
- [Supabase](#supabase)
- [Conectividad](#conectividad)
- [Rendimiento](#rendimiento)

---

## Git & Repositorio

### ❌ "Permission denied (publickey)"

**Síntoma:**
```
git@github.com: Permission denied (publickey).
fatal: Could not read from remote repository.
```

**Solución:**

```bash
# 1. Generar clave SSH
ssh-keygen -t ed25519 -C \"tu.email@ejemplo.com\"

# 2. Agregar a ssh-agent
eval $(ssh-agent -s)
ssh-add ~/.ssh/id_ed25519

# 3. Copiar clave pública a GitHub
cat ~/.ssh/id_ed25519.pub
# Copia la salida a GitHub Settings > SSH Keys

# 4. Verificar conexión
ssh -T git@github.com
# Debería responder: Hi username! You've successfully authenticated...

# 5. Reintentar clone
git clone git@github.com:tuorganizacion/meloop.git
```

### ❌ "Repository not found"

**Solución:**
```bash
# Verifica la URL del repositorio
git remote -v

# Actualiza si es necesario
git remote set-url origin https://github.com/tuorganizacion/meloop.git

# O usando SSH
git remote set-url origin git@github.com:tuorganizacion/meloop.git
```

---

## Herramientas

### ❌ Comando no encontrado (Go, Python, Node.js, etc.)

**Windows - PowerShell:**
```powershell
# Verifica que está en PATH
$env:Path -split ';' | grep go

# Si no aparece, reinstala la herramienta
# Después, reinicia PowerShell

# Si aún falla, agrega manualmente a PATH:
[Environment]::SetEnvironmentVariable(
    \"Path\",
    $env:Path + \";C:\\Program Files\\Go\\bin\",
    [System.EnvironmentVariableTarget]::User
)

# Reinicia PowerShell
```

**macOS/Linux:**
```bash
# Verifica que está en PATH
echo $PATH

# Agrega a ~/.bashrc o ~/.zshrc
echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verifica
go version
```

### ❌ Versión incorrecta

**Solución:**
```bash
# Verifica versión instalada
go version
python --version
node --version
flutter --version

# Desinstala versión antigua
# Reinstala versión correcta
# Reinicia terminal
```

---

## Docker

### ❌ Docker no inicia en Windows

**Síntoma:**
```
Docker daemon is not running
```

**Solución:**

1. **Reinicia Docker Desktop:**
   - Cierra Docker Desktop completamente
   - Abre nuevamente
   - Espera a que esté listo (icono estable)

2. **Verifica WSL 2:**
   ```powershell
   wsl --list --verbose
   # Debería mostrar un distro con version 2
   
   # Si necesitas instalar WSL 2:
   wsl --install
   ```

3. **Reinicia tu computadora:**
   - Es efectivo en 80% de los casos

4. **Limpia caché de Docker:**
   ```powershell
   docker system prune -a
   docker volume prune
   ```

### ❌ "Docker daemon is not running" en Linux

**Solución:**
```bash
# Inicia el servicio Docker
sudo systemctl start docker

# Verifica estado
sudo systemctl status docker

# Autoriza tu usuario (sin sudo)
sudo usermod -aG docker $USER
newgrp docker

# Reinicia sesión
logout
login
```

### ❌ Puerto ya en uso

**Síntoma:**
```
Address already in use
Bind: permission denied
```

**Windows:**
```powershell
# Encuentra proceso en el puerto
netstat -ano | findstr :8080

# Resultado: TCP 0.0.0.0:8080 0.0.0.0:0 LISTENING 12345

# Mata el proceso
taskkill /PID 12345 /F

# O cambia el puerto en docker-compose.yml
# Reemplaza: 8080:8080
# Por: 8081:8080
```

**macOS/Linux:**
```bash
# Encuentra proceso
lsof -i :8080

# Resultado: COMMAND PID  USER  FD TYPE DEVICE SIZE NODE NAME
#           chrome  123 user   12u IPv4 ...

# Mata el proceso
kill -9 123

# O
sudo lsof -i :8080 | grep LISTEN | awk '{print \$2}' | xargs kill -9
```

### ❌ "No such file or directory" en docker-compose.yml

**Solución:**
```bash
# Verifica que estás en la carpeta correcta
cd meloop

# Verifica que docker-compose.yml existe
ls docker-compose.yml

# Valida el archivo
docker-compose config

# Si hay errores de YAML, abre y verifica la indentación
cat docker-compose.yml | head -20
```

### ❌ Contenedor se detiene inmediatamente

**Solución:**
```bash
# Ver logs detallados
docker-compose logs redis
docker-compose logs rabbitmq

# Si falta archivo, verifica volumes
docker volume ls
docker volume inspect meloop_redis_data

# Si falta, recrea
docker-compose down -v
docker-compose up -d
```

### ❌ Los volúmenes no persisten (Linux)

**Solución:**
```bash
# Verifica permisos
ls -la /var/lib/docker/volumes/

# Ajusta permisos
sudo chown -R $USER:$USER /var/lib/docker/volumes/meloop_*

# O
sudo chmod 755 /var/lib/docker/volumes/meloop_*
```

---

## Backend (NestJS)

### ❌ \"Cannot find module '@nestjs/common'\"

**Solución:**
```bash
cd backend

# Limpia instalación
rm -r node_modules package-lock.json

# Reinstala
npm install

# Verifica
npm list @nestjs/core
```

### ❌ Puerto 3000 ya en uso

```bash
# Windows
netstat -ano | findstr :3000
taskkill /PID <PID> /F

### Linux
lsof -i :3000
kill -9 <PID>

# O cambia el puerto en main.ts
# De: listen(3000)
# A: listen(3001)
```

### ❌ NestJS no inicia con \"Cannot find module\"

**Solución:**
```bash
cd backend

# Opción 1: Limpiar y reinstalar
npm cache clean --force
rm -r node_modules package-lock.json
npm install

# Opción 2: Compilar antes de ejecutar
npm run build
npm run start

# Opción 3: Verificar configuración
cat tsconfig.json | grep -A 5 paths
```

### ❌ Hot reload no funciona

```bash
# Verifica que ejecutas con start:dev (no start)
npm run start:dev

# Si aún no funciona, fuerza recarga manual
# Presiona Ctrl+C y reinicia
npm run start:dev
```

### ❌ Conexión rechazada a Redis/RabbitMQ

```bash
# Verifica que servicios Docker están corriendo
docker-compose ps

# Si no están, inicia
docker-compose up -d

# Verifica logs
docker-compose logs redis
docker-compose logs rabbitmq

# Verifica .env tiene credenciales correctas
cat .env | grep REDIS
cat .env | grep RABBITMQ
```

---

## Servicios Go

### ❌ \"go: no go files in\"

**Solución:**
```bash
# Verifica estructura
ls services/api-gateway/
# Debería incluir: main.go, go.mod, go.sum

# Si falta go.mod
cd services/api-gateway
go mod init github.com/meloop/api-gateway
```

### ❌ Build error: \"undefined reference\"

```bash
# Actualiza dependencias
go get -u ./...
go mod tidy

# Reconstruye
go build -o app
```

### ❌ \"connection refused\" a otros servicios

```bash
# Verifica que todos los servicios están ejecutándose
# En Docker:
docker-compose ps

# Localmente:
# Asegúrate que ejecutas en terminales separadas:
# Terminal 1: cd services/api-gateway && go run main.go
# Terminal 2: cd services/auth-service && go run main.go
# Terminal 3: cd services/user-service && go run main.go
```

### ❌ Main.go no compila

```bash
# Verifica sintaxis
go vet ./...

# Verifica imports
grep -r \"package main\" .

# Si hay múltiples main, mantén solo uno por servicio
```

---

## ML Service (Python)

### ❌ \"No module named 'app'\"

**Solución:**
```bash
cd ml-service

# Verifica que app/ existe
ls app/

# Verifica que estás en el directorio correcto
pwd
# Output: /path/to/meloop/ml-service

# Verifica que el entorno virtual está activado
# (venv) debería aparecer en el prompt

# Si no, activa:
.\venv\\Scripts\\Activate.ps1    # Windows
source venv/bin/activate         # macOS/Linux
```

### ❌ \"python: No module named uvicorn\"

**Solución:**
```bash
# Verifica que el entorno virtual está activado
which python
# Debería mostrar: /path/to/venv/bin/python

# Reinstala dependencias
pip install -r requirements.txt

# Verifica instalación
pip list | grep uvicorn
```

### ❌ Entorno virtual no se activa (PowerShell)

**Síntoma:**
```
File cannot be loaded because running scripts is disabled
```

**Solución:**
```powershell
# Cambiar política de ejecución
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Intentar activar nuevamente
.\\venv\\Scripts\\Activate.ps1

# Si aún falla, usa cmd.exe
cmd
venv\\Scripts\\activate.bat
```

### ❌ Módulo numpy/pandas: \"ImportError\"

**Solución:**
```bash
# Actualiza pip
pip install --upgrade pip

# Reinstala paquetes
pip install --force-reinstall -r requirements.txt

# Verifica versiones correctas
pip show numpy pandas scikit-learn
```

### ❌ Puerto 8000 ya en uso

```bash
# Windows
netstat -ano | findstr :8000
taskkill /PID <PID> /F

### Linux
lsof -i :8000
kill -9 <PID>

# O ejecuta en puerto diferente
python -m uvicorn app.main:app --port 8001
```

---

## Flutter

### ❌ \"flutter: command not found\"

**Solución:**
```bash
# Verifica instalación
ls $HOME/flutter/bin/flutter
# O en Windows
dir \"C:\\src\\flutter\\bin\\flutter.exe\"

# Agrega al PATH
echo 'export PATH=\"\$PATH:\$HOME/flutter/bin\"' >> ~/.bashrc
source ~/.bashrc

# O en Windows, usa Variables de Entorno
setx PATH \"%PATH%;C:\\src\\flutter\\bin\"
```

### ❌ \"Unable to locate a browser\"

**Solución:**
```bash
# Instala Chrome
# Windows: https://www.google.com/chrome/
# macOS: brew install google-chrome
# Linux: sudo apt install google-chrome-stable

# Especifica navegador en Flutter
flutter run -d chrome --web-renderer=html
```

### ❌ \"No devices available\"

**Solución:**
```bash
# Lista dispositivos disponibles
flutter devices

# Si no hay, habilita plataforma web
flutter config --enable-web

# Prueba en web
flutter run -d chrome

# O crea un simulador (Android/iOS)
flutter emulators
flutter emulators --launch <emulator-id>
```

### ❌ \"Build failed\" en Flutter

**Solución:**
```bash
cd app/flutter_app

# Limpia compilación anterior
flutter clean

# Obtiene dependencias
flutter pub get

# Ejecuta análisis
flutter analyze

# Intenta compilar nuevamente
flutter run
```

### ❌ Hot reload no funciona

```bash
# Presiona Ctrl+C para salir

# Limpia e intenta nuevamente
flutter clean
flutter pub get
flutter run

# Si persiste, reinicia el debugger
# En el terminal de Flutter, presiona q para salir
# Luego ejecuta flutter run nuevamente
```

### ❌ \"AndroidManifest.xml not found\"

```bash
# Para Android, asegúrate de tener SDK instalado
flutter doctor

# Si AndroidSDK falta, instálalo:
flutter config --android-sdk-path=<path-to-sdk>

# O compila solo para web:
flutter run -d chrome
```

---

## Supabase

### ❌ \"Invalid API Key\"

**Síntoma:**
```
Supabase Error: Unauthorized
```

**Solución:**
```bash
# Verifica credenciales en .env
cat .env | grep SUPABASE

# Debe mostrar:
# SUPABASE_URL=https://xxxx.supabase.co
# SUPABASE_KEY=eyJ...

# Verifica que es la clave ANON, no service_role
# En Supabase Dashboard: Settings > API > anon public key

# Copia nuevamente credenciales
# (puede haber espacios o caracteres ocultos)

# Actualiza .env
nano .env  # o tu editor preferido
```

### ❌ \"Connection refused\" a Supabase

**Síntoma:**
```
Error connecting to Supabase
```

**Solución:**
```bash
# Verifica que el proyecto existe en Supabase
# Dashboard: https://app.supabase.com

# Verifica que la URL es accesible
curl https://tu-proyecto.supabase.co/rest/v1/

# Verifica que el proyecto está \"Active\"
# Si está \"Paused\", reinicia desde Dashboard

# Prueba con curl
curl -H \"apikey: tu_key_aqui\" https://tu-proyecto.supabase.co/rest/v1/
```

### ❌ \"Row-level security violates policy\"

**Síntoma:**
```
Error: RLS policy violation
```

**Solución:**
```sql
-- En SQL Editor de Supabase

-- Verifica que RLS está habilitado
SELECT tablename FROM pg_tables WHERE schemaname = 'public';

-- Ver políticas actuales
SELECT * FROM pg_policies;

-- Crear política básica de lectura
CREATE POLICY \"Allow public read\"
  ON public.users
  FOR SELECT
  USING (true);

-- Crear política de escritura para usuarios autenticados
CREATE POLICY \"Allow authenticated write\"
  ON public.users
  FOR INSERT
  WITH CHECK (auth.uid() = id);
```

---

## Conectividad

### ❌ Servicios no pueden conectarse entre sí

**Síntoma:**
```
Connection refused
Failed to connect
```

**Solución:**

1. **Verifica que todos están ejecutándose:**
   ```bash
   docker-compose ps
   curl http://localhost:8080/health
   curl http://localhost:3000/health
   ```

2. **Verifica URLs correctas:**
   - En Docker: usa nombre del servicio (ej: redis:6379)
   - Localmente: usa localhost (ej: localhost:6379)
   - Desde host a Docker: usa docker.for.mac.localhost o host.docker.internal

3. **Verifica puertos:**
   ```bash
   docker-compose ps
   # Verifica PORTS
   ```

4. **Prueba conectividad:**
   ```bash
   telnet localhost 8080
   telnet localhost 6379
   ```

---

## Rendimiento

### ❌ Servicios muy lentos

**Solución:**
```bash
# Verifica recursos disponibles
docker stats

# Limpia imágenes no usadas
docker image prune -a

# Limpia volúmenes no usados
docker volume prune

# Limpia caché de build
docker builder prune
```

### ❌ Memoria agotada

```bash
# Aumenta memoria disponible para Docker
# Settings > Resources > Memory Limit

# O limita servicios específicos
# En docker-compose.yml:
services:
  redis:
    mem_limit: 256m
  rabbitmq:
    mem_limit: 512m
```

### ❌ CPU al 100%

```bash
# Identifica proceso problemático
docker stats

# Ver proceso específico
docker top <container-name>

# Reinicia el servicio
docker-compose restart <service>
```

---

## 🆘 Cuando Nada Funciona

### \"Reinicio Nuclear\" (Última opción)

```bash
# ¡CUIDADO! Esto elimina TODO

# 1. Detén servicios
docker-compose down -v

# 2. Limpia todo Docker
docker system prune -a
docker volume prune

# 3. Limpia proyecto
cd meloop
rm -r backend/node_modules ml-service/venv app/flutter_app/build

# 4. Recopia .env
cp .env.example .env

# 5. Comienza de nuevo
# Sigue: 07-quick-start.md
```

### Recoletar Logs para Soporte

```bash
# Guarda todos los logs
docker-compose logs > logs.txt

# Agrega info del sistema
docker info >> logs.txt
docker ps -a >> logs.txt
docker volume ls >> logs.txt

# Comparte logs.txt con el equipo de soporte
```

---

## Pedir Ayuda

Si nada funciona:

1. **Ejecuta diagnóstico completo:**
   ```bash
   # Recolecta información
   docker-compose ps
   docker-compose logs
   docker info
   flutter doctor
   go version
   python --version
   node --version
   ```

2. **Crea reporte:**
   - Versiones de herramientas
   - SO (Windows/macOS/Linux)
   - Pasos que ejecutaste
   - Mensaje de error exacto
   - Logs relevantes

3. **Contacta:**
   - Team lead de desarrollo
   - GitHub Issues
   - Slack/Discord del equipo

---

## Recursos Adicionales

- **Docker:** https://docs.docker.com/
- **NestJS:** https://docs.nestjs.com/
- **Flutter:** https://flutter.dev/docs
- **Go:** https://golang.org/doc/
- **Python/FastAPI:** https://fastapi.tiangolo.com/
- **Supabase:** https://supabase.com/docs

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0

---

**¿No encontraste tu problema?** Crea un issue en GitHub o contacta al equipo de desarrollo.

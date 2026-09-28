# 2. Configuración del Entorno

En este documento se explica cómo configurar las variables de entorno, credenciales y servicios externos como Supabase.

**Tiempo estimado:** 15-20 minutos

---

## 📝 Archivo .env

El archivo `.env` contiene todas las variables de entorno que necesitan los servicios. Se encuentra en la raíz del proyecto.

### Paso 1: Crear archivo .env

```bash
cd meloop

# En Windows (PowerShell)
Copy-Item .env.example .env

### Linux
```bash
cp .env.example .env
```
```

### Paso 2: Configurar variables de .env

Abre `meloop/.env` en tu editor de texto favorito y configura los siguientes valores:

```env
# =============================================================================
# REDIS - Cache y sesiones
# =============================================================================
REDIS_PASSWORD=tu_contrasena_redis_aqui

# =============================================================================
# RABBITMQ - Message Broker
# =============================================================================
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# =============================================================================
# MINIO - S3-Compatible Storage (para imágenes, audios, etc)
# =============================================================================
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

# Credenciales para la aplicación (diferente a root)
MINIO_MEDIA_USER=mediauser
MINIO_MEDIA_PASSWORD=cambiar_esto_en_produccion
MINIO_MEDIA_BUCKET=meloop-media

# Endpoints (http://minio:9000 es interno en Docker)
MINIO_ENDPOINT=http://minio:9000
MINIO_PUBLIC_ENDPOINT=http://localhost:9000

# =============================================================================
# SUPABASE - Auth y Base de Datos
# =============================================================================
SUPABASE_URL=https://tuproyecto.supabase.co
SUPABASE_KEY=tu_anon_key_aqui

# =============================================================================
# API GATEWAY
# =============================================================================
API_GATEWAY_PORT=8080
API_GATEWAY_HOST=0.0.0.0

# =============================================================================
# SERVICIOS GO
# =============================================================================
# Auth Service
AUTH_SERVICE_PORT=8083

# User Service
USER_SERVICE_PORT=8082

# Chat Service
CHAT_SERVICE_PORT=8084

# Media Service
MEDIA_SERVICE_PORT=8085

# =============================================================================
# BACKEND NESTJS
# =============================================================================
NESTJS_PORT=3000
NODE_ENV=development

# =============================================================================
# ML SERVICE (PYTHON)
# =============================================================================
ML_SERVICE_PORT=8000
ML_SERVICE_HOST=0.0.0.0

# =============================================================================
# LOGGING
# =============================================================================
LOG_LEVEL=debug

# =============================================================================
# ENTORNO
# =============================================================================
ENVIRONMENT=development
```

---

## 🔐 Configuración de Supabase

Supabase es un servicio de backend como servicio que proporciona autenticación, base de datos PostgreSQL y almacenamiento en tiempo real.

### Paso 1: Crear cuenta en Supabase

1. Accede a [supabase.com](https://supabase.com)
2. Haz clic en "Start your project"
3. Inicia sesión con GitHub o crea una cuenta

### Paso 2: Crear un nuevo proyecto

1. Haz clic en "New project"
2. Completa los datos:
   - **Database password:** Elige una contraseña segura (cópiala, la necesitarás)
   - **Region:** Elige la región más cercana a tus usuarios
   - **Pricing plan:** Elige "Free" para desarrollo

3. Espera a que se provisione (2-3 minutos)

### Paso 3: Obtener credenciales

Una vez creado el proyecto:

1. En el menú izquierdo, ve a **Settings** → **API**
2. Encontrarás:
   - **Project URL** (en la sección "URL")
   - **anon public** (en "API Keys")

### Paso 4: Configurar .env con credenciales de Supabase

```env
SUPABASE_URL=https://tuproyecto.supabase.co
SUPABASE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

> **⚠️ IMPORTANTE:** La `SUPABASE_KEY` que necesitas es la **clave anónima (`anon`)**, NO la clave de servicio.

### Paso 5: Configurar tablas y políticas de seguridad

Supabase proporciona un editor SQL integrado. Deberás crear las tablas necesarias para tu aplicación.

Accede a **SQL Editor** en el dashboard de Supabase y crea las tablas según tu esquema. Ejemplo básico:

```sql
-- Tabla de usuarios (esta es generada automáticamente por Supabase Auth)
-- Solo necesitas crear tablas adicionales específicas de tu app

-- Tabla de perfiles de usuario
CREATE TABLE public.user_profiles (
  id uuid NOT NULL PRIMARY KEY,
  username text NOT NULL UNIQUE,
  full_name text,
  avatar_url text,
  bio text,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now()
);

-- Tabla de posts
CREATE TABLE public.posts (
  id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES public.user_profiles(id) ON DELETE CASCADE,
  title text NOT NULL,
  content text NOT NULL,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now()
);

-- Habilitar RLS (Row Level Security)
ALTER TABLE public.user_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.posts ENABLE ROW LEVEL SECURITY;

-- Políticas de seguridad (ejemplo)
CREATE POLICY "Users can view any profile"
  ON public.user_profiles
  FOR SELECT
  USING (true);

CREATE POLICY "Users can update their own profile"
  ON public.user_profiles
  FOR UPDATE
  USING (auth.uid() = id);
```

### Paso 6: Verificar conexión

Verifica que Supabase está configurado correctamente ejecutando este comando en Node.js:

```bash
cd backend

# Instala el cliente de Supabase si aún no lo tienes
npm install @supabase/supabase-js

# En Node REPL, prueba la conexión
node -e "
const { createClient } = require('@supabase/supabase-js');
const client = createClient(
  process.env.SUPABASE_URL,
  process.env.SUPABASE_KEY
);
client.auth.getSession().then(console.log).catch(console.error);
"
```

---

## 🎛️ Variables de Entorno por Servicio

### Redis

```env
REDIS_PASSWORD=tu_contrasena_segura
```

**Notas:**
- Se accede internamente como `redis://default:PASSWORD@redis:6379/0`
- Puertos: `6379` (interno), `6379` (externo)

### RabbitMQ

```env
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest
```

**Notas:**
- Por defecto usa credenciales `guest`/`guest`
- Puertos: `5672` (AMQP), `15672` (Management)
- Acceso al dashboard: http://localhost:15672

### MinIO (S3-Compatible Storage)

```env
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

MINIO_MEDIA_USER=mediauser
MINIO_MEDIA_PASSWORD=cambiar_esto
MINIO_MEDIA_BUCKET=meloop-media

MINIO_ENDPOINT=http://minio:9000
MINIO_PUBLIC_ENDPOINT=http://localhost:9000
```

**Notas:**
- Usa credenciales root para gestión (minioadmin)
- Usa credenciales de aplicación (mediauser) en código
- Puertos: `9000` (API), `9001` (Console)
- Console web: http://localhost:9001

---

## 🔑 Valores Recomendados para Desarrollo

Para desarrollo local, puedes usar valores simples y fáciles de recordar:

```env
# Redis
REDIS_PASSWORD=devpass123

# RabbitMQ (usa valores por defecto)
RABBITMQ_USER=guest
RABBITMQ_PASSWORD=guest

# MinIO
MINIO_ROOT_USER=dev
MINIO_ROOT_PASSWORD=devpass123
MINIO_MEDIA_USER=appuser
MINIO_MEDIA_PASSWORD=apppass123
MINIO_MEDIA_BUCKET=meloop-dev

# Supabase (ver sección anterior)
SUPABASE_URL=https://tu-proyecto.supabase.co
SUPABASE_KEY=eyJ...

# Servicios
API_GATEWAY_PORT=8080
NESTJS_PORT=3000
ML_SERVICE_PORT=8000
LOG_LEVEL=debug
ENVIRONMENT=development
```

---

## Validación de Configuración

### Paso 1: Verificar archivo .env

```bash
# Windows PowerShell
cat meloop\.env

# Linux
cat meloop/.env

# Deberías ver todas las variables configuradas
```

### Paso 2: Cargar variables en tu sesión actual

```bash
# En PowerShell (Windows)
Get-Content .env | ForEach-Object {
  $name, $value = $_ -split '=', 2
  if ($name -and -not $name.StartsWith('#')) {
    [Environment]::SetEnvironmentVariable($name, $value)
  }
}

### Linux
```bash
export $(cat .env | grep -v '^#' | xargs)
```
```

### Paso 3: Verificar variables cargadas

```bash
# Windows PowerShell
Get-ChildItem env: | Where-Object { $_.Name -like "REDIS_*" -or $_.Name -like "SUPABASE_*" }

# Linux
env | grep -E "REDIS_|SUPABASE_|MINIO_"
```

---

## 🚀 Próximo Paso

Ahora que tienes el entorno configurado:

👉 **Ve a [3. Docker & Servicios →](03-docker-setup.md)**

---

## Problemas Comunes

### Supabase URL o Key incorrectos
- Verifica que copiaste exactamente desde el dashboard
- Asegúrate de usar la clave **`anon`**, no la de servicio
- Debería ser similar a: `https://xxxxxxxxxxxx.supabase.co`

### Variables de entorno no se cargan
- En Windows, reinicia PowerShell
- Usa la ruta completa: `$env:REDIS_PASSWORD` en PowerShell
- En Bash, verifica con: `echo $VARIABLE_NAME`

### Conexión a Supabase rechazada
- Verifica que la URL es correcta (sin espacios extra)
- Confirma que el proyecto está activo en Supabase
- Revisa la región de Supabase (debe ser consistente)

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0

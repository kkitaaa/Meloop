# 3. Docker & Servicios

En este documento se explica cómo levantar y gestionar los servicios de Docker Compose (Redis, RabbitMQ, MinIO).

**Tiempo estimado:** 20-30 minutos

---

## Servicios en Docker Compose

El proyecto Meloop usa Docker Compose para orquestar los siguientes servicios:

| Servicio | Puerto Externo | Puerto Interno | Función |
|----------|---|---|---|
| **API Gateway** | 8080 | 8080 | Enrutador principal (Go) |
| **Test Service** | 8081 | 8081 | Servicio de prueba (Go) |
| **User Service** | 8082 | 8082 | Gestión de usuarios (Go) |
| **Auth Service** | 8083 | 8083 | Autenticación (Go) |
| **Redis** | 6379 | 6379 | Cache en memoria |
| **RabbitMQ** | 5672, 15672 | 5672, 15672 | Message broker + Dashboard |
| **MinIO** | 9000, 9001 | 9000, 9001 | S3-compatible storage + Console |

---

## 🚀 Iniciar Docker Compose

### Paso 1: Verificar Docker está ejecutándose

```bash
docker ps
```

Si Docker no está ejecutándose, abre Docker Desktop en Windows o inicia el servicio en Linux.

### Paso 2: Crear la red Docker (si no existe)

```bash
docker network create meloop-network
```

### Paso 3: Levantar los servicios

```bash
cd meloop

# Levantar todos los servicios en background
docker-compose up -d

# Ver el progreso (opcional)
docker-compose up
# Presiona Ctrl+C para volver al terminal manteniendo servicios ejecutándose
```

**Salida esperada:**
```
Creating meloop-redis-1 ... done
Creating meloop-rabbitmq-1 ... done
Creating meloop-minio-1 ... done
Creating meloop-minio-init-1 ... done
Creating meloop-api-gateway-1 ... done
```

### Paso 4: Verificar servicios activos

```bash
docker-compose ps
```

**Salida esperada:**
```
NAME                      STATUS              PORTS
meloop-api-gateway-1      Up 2 minutes        0.0.0.0:8080->8080/tcp
meloop-redis-1            Up 2 minutes        0.0.0.0:6379->6379/tcp
meloop-rabbitmq-1         Up 2 minutes        0.0.0.0:5672->5672/tcp, 0.0.0.0:15672->15672/tcp
meloop-minio-1            Up 2 minutes        0.0.0.0:9000->9000/tcp, 0.0.0.0:9001->9001/tcp
```

---

## Validar Servicios

### Redis

```bash
# Conectar a Redis y probar comando
docker exec -it meloop-redis-1 redis-cli -a <REDIS_PASSWORD>

# Una vez dentro de Redis
127.0.0.1:6379> PING
PONG
127.0.0.1:6379> SET test "Hello Redis"
OK
127.0.0.1:6379> GET test
"Hello Redis"
127.0.0.1:6379> exit
```

**Validación remota:**
```bash
# Desde tu máquina
telnet localhost 6379

# Si ves la conexión abierta, Redis está ejecutándose
```

### RabbitMQ

**Dashboard web:**
- URL: http://localhost:15672
- Usuario: `guest`
- Contraseña: `guest`

**Verificar desde terminal:**
```bash
docker exec -it meloop-rabbitmq-1 rabbitmq-diagnostics status
```

### MinIO

**Console web:**
- URL: http://localhost:9001
- Usuario: `minioadmin`
- Contraseña: `minioadmin`

**Crear bucket usando CLI:**
```bash
# Instalar MC (MinIO Client)
# Windows
choco install minio-client

# macOS
brew install minio-mc

# Linux
curl https://dl.min.io/client/mc/release/linux-amd64/mc --create-dirs -o ./mc
chmod +x ./mc

# Configurar conexión
mc alias set minio http://localhost:9000 minioadmin minioadmin

# Ver buckets existentes
mc ls minio/

# Crear bucket si no existe
mc mb minio/meloop-media --ignore-existing
```

---

## 🐳 Comandos Útiles de Docker Compose

### Ver logs

```bash
# Todos los servicios
docker-compose logs

# Servicio específico
docker-compose logs redis
docker-compose logs rabbitmq
docker-compose logs minio

# Seguir logs en tiempo real (último 100 líneas)
docker-compose logs -f --tail=100

# Ver logs de un servicio específico
docker-compose logs -f api-gateway
```

### Detener servicios

```bash
# Detener manteniendo datos
docker-compose stop

# Detener servicio específico
docker-compose stop redis

# Esperar a que se detengan
docker-compose stop --timeout=30
```

### Reiniciar servicios

```bash
# Reiniciar todos
docker-compose restart

# Reiniciar servicio específico
docker-compose restart redis
```

### Detener y eliminar contenedores (pero mantener volúmenes)

```bash
docker-compose down

# Eliminar también volúmenes (CUIDADO: pierdes datos)
docker-compose down -v
```

### Reconstruir imágenes

```bash
# Reconstruir servicio específico
docker-compose build api-gateway

# Reconstruir todos
docker-compose build

# Reconstruir y reiniciar
docker-compose up -d --build
```

---

## 📁 Volúmenes Persistentes

Los servicios usan volúmenes para persistir datos:

```yaml
volumes:
  redis_data:    # Datos de Redis
  rabbitmq_data: # Datos de RabbitMQ
  minio_data:    # Datos de MinIO
```

Para ver los volúmenes en tu sistema:

```bash
docker volume ls

# Ver detalles de un volumen
docker volume inspect meloop_redis_data

# Limpiar volúmenes no usados (CUIDADO)
docker volume prune
```

---

## Conectarse a Servicios desde tu Aplicación

### Desde dentro de Docker (ej: otros contenedores)

```
redis://default:PASSWORD@redis:6379/0
amqp://guest:guest@rabbitmq:5672//
http://minio:9000
```

### Desde tu máquina local

```
redis://default:PASSWORD@localhost:6379/0
amqp://guest:guest@localhost:5672//
http://localhost:9000
```

---

## 🚀 Levantar Servicios Backend Adicionales

Además de Docker Compose, también necesitarás ejecutar:

### Backend NestJS

```bash
cd backend

# Instalar dependencias
npm install

# Ejecutar en desarrollo
npm run start:dev

# Debería mostrar:
# [Nest] 12345 - 09/01/2024, 10:30:00 AM     LOG [NestFactory] Starting Nest application...
# [Nest] 12345 - 09/01/2024, 10:30:01 AM     LOG [InstanceLoader] AppModule dependencies initialized...
```

**Validar NestJS:**
```bash
curl http://localhost:3000/health
# Output: {"status":"ok"}
```

### Servicios Go (en paralelo)

```bash
cd services/api-gateway
go run main.go

# En otra terminal
cd services/auth-service
go run main.go
```

---

## Resolución de Problemas

### Puerto ya en uso

```bash
# Windows PowerShell - encontrar proceso en puerto 8080
Get-Process | Where-Object { $_.Handles.Count -gt 0 } | Where-Object { netstat -ano | Select-String ":8080" }

# Método más simple
netstat -ano | findstr :8080

# macOS/Linux
lsof -i :8080
sudo kill -9 <PID>
```

### Docker daemon no responde

```bash
# Reiniciar Docker
# Reinicia Docker Desktop
# Linux
sudo systemctl restart docker
```

### Contenedor se detiene inmediatamente

```bash
# Ver logs detallados
docker-compose logs api-gateway

# Verificar que el archivo docker-compose.yml es válido
docker-compose config
```

### Problema de permisos en volúmenes (Linux)

```bash
# Ajustar permisos
sudo chown -R $USER:$USER /var/lib/docker/volumes/meloop_*
```

### MinIO bucket no se crea automáticamente

```bash
# Crear manualmente
docker exec -it meloop-minio-1 mc mb /data/meloop-media
```

---

## Monitoreo de Recursos

### Ver consumo de recursos

```bash
docker stats

# Output:
# CONTAINER ID   NAME             CPU %  MEM USAGE / LIMIT   MEM %
# abc123         meloop-redis-1   0.2%   45.2MiB / 1GiB      4.41%
# def456         meloop-minio-1   1.5%   128MiB / 1GiB       12.5%
```

---

## 🚀 Próximo Paso

Una vez que Docker Compose está ejecutándose correctamente:

👉 **Ve a [4. Backend (Go + NestJS) →](04-backend-setup.md)**

---

## Checklist de Docker

```
□ Docker Desktop instalado y ejecutándose
□ Docker Compose versión 2.20+
□ Red meloop-network creada
□ Comando: docker-compose up -d ejecutado exitosamente
□ docker-compose ps muestra todos los servicios
□ Redis responde a redis-cli PING
□ RabbitMQ dashboard accesible en :15672
□ MinIO console accesible en :9001
□ Volúmenes persisten datos correctamente
□ Logs no muestran errores críticos
```

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0

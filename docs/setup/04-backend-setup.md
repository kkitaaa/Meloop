# 4. Backend (Go + NestJS)

En este documento se explica cómo configurar e instalar los servicios backend en Go y NestJS.

**Tiempo estimado:** 20-30 minutos

---

## Resumen de Servicios Backend

| Servicio | Lenguaje | Puerto | Descripción |
|----------|----------|--------|-------------|
| **API Gateway** | Go | 8080 | Enrutador central |
| **Auth Service** | Go | 8083 | Autenticación y autorización |
| **User Service** | Go | 8082 | Gestión de usuarios |
| **Chat Service** | Go | 8084 | Mensajería |
| **Media Service** | Go | 8085 | Gestión de archivos |
| **Backend NestJS** | TypeScript/Node | 3000 | Servicio monolítico |

---

## Servicios Go

### Paso 1: Verificar Go está instalado

```bash
go version
# Output: go version go1.26.6 windows/amd64
```

### Paso 2: Descargar dependencias globales

```bash
# En la raíz del proyecto
go mod download

# Inicializar workspace
go work use ./services/common
go work use ./services/api-gateway
go work use ./services/auth-service
go work use ./services/user-service
go work use ./services/chat-service
go work use ./services/media-service
```

### Paso 3: Compilar servicios Go

Antes de ejecutar, necesitas compilar los binarios. Hay dos opciones:

#### Opción A: Compilación directa (Desarrollo)

```bash
# Compilar API Gateway
cd services/api-gateway
go build -o api-gateway.exe  # Windows
go build -o api-gateway      # Linux

# Compilar Auth Service
cd ../auth-service
go build -o auth-service.exe
go build -o auth-service

# Compilar User Service
cd ../user-service
go build -o user-service.exe
go build -o user-service
```

#### Opción B: Usar Docker (Recomendado)

Los servicios Go ya están configurados en `docker-compose.yml`. Si levantaste Docker Compose en el paso anterior, ya están compilados y ejecutándose.

```bash
docker-compose ps | grep -E "api-gateway|auth-service|user-service"
```

### Paso 4: Ejecutar servicios Go directamente (sin Docker)

Si prefieres ejecutarlos sin Docker:

```bash
# Terminal 1: API Gateway
cd meloop/services/api-gateway
go run main.go
# Output: [GIN-debug] Listening and serving HTTP on :8080

# Terminal 2: Auth Service
cd meloop/services/auth-service
go run main.go
# Output: [GIN-debug] Listening and serving HTTP on :8083

# Terminal 3: User Service
cd meloop/services/user-service
go run main.go
# Output: [GIN-debug] Listening and serving HTTP on :8082
```

### Paso 5: Validar servicios Go

```bash
# Validar API Gateway
curl http://localhost:8080/health
# Output: {"status":"ok"}

# Validar Auth Service
curl http://localhost:8083/health
# Output: {"status":"ok"}

# Validar User Service
curl http://localhost:8082/health
# Output: {"status":"ok"}
```

---

## 🏗️ Backend NestJS

### Paso 1: Instalar dependencias

```bash
cd meloop/backend

# Instalar usando npm
npm install

# O usando yarn
yarn install
```

**Salida esperada:**
```
added XXX packages in X.XXs
```

### Paso 2: Verificar instalación

```bash
npm list @nestjs/core

# Output:
# backend@0.0.1 /path/to/meloop/backend
# └── @nestjs/core@11.0.1
```

### Paso 3: Ejecutar en modo desarrollo

```bash
npm run start:dev
```

**Salida esperada:**
```
[Nest] 12345 - 09/01/2024, 10:30:00 AM     LOG [NestFactory] Starting Nest application...
[Nest] 12345 - 09/01/2024, 10:30:01 AM     LOG [NestLoader] AppModule dependencies initialized
[Nest] 12345 - 09/01/2024, 10:30:02 AM     LOG [NestApplication] Nest application successfully started
[Nest] 12345 - 09/01/2024, 10:30:02 AM     LOG [NestApplication] Listening on port 3000
```

### Paso 4: Validar NestJS

```bash
curl http://localhost:3000/health
# Output: {"status":"ok"}
```

### Alternativa: Ejecutar en modo producción

```bash
# Compilar
npm run build

# Ejecutar versión compilada
npm run start:prod
```

---

## Configuración de Conexiones entre Servicios

Los servicios se comunican entre sí. Aquí está cómo configurarlos:

### Desde NestJS a Servicios Go

```typescript
// backend/src/services/api.service.ts
import axios from 'axios';

@Injectable()
export class ApiService {
  private readonly logger = new Logger(ApiService.name);

  constructor() {}

  // Llamar a Auth Service
  async verifyToken(token: string) {
    try {
      const response = await axios.get('http://localhost:8083/verify', {
        headers: { Authorization: `Bearer ${token}` }
      });
      return response.data;
    } catch (error) {
      this.logger.error('Auth service error:', error.message);
      throw error;
    }
  }

  // Llamar a User Service
  async getUser(userId: string) {
    try {
      const response = await axios.get(`http://localhost:8082/users/${userId}`);
      return response.data;
    } catch (error) {
      this.logger.error('User service error:', error.message);
      throw error;
    }
  }
}
```

### Desde Servicios Go a otros Servicios

```go
// services/api-gateway/controllers/handler.go
package controllers

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
)

func GetUserHandler(c *gin.Context) {
    userID := c.Param("id")
    
    // Llamar a User Service
    resp, err := http.Get(fmt.Sprintf("http://localhost:8082/users/%s", userID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User service error"})
        return
    }
    defer resp.Body.Close()
    
    // Procesar respuesta...
    c.JSON(http.StatusOK, gin.H{"user_id": userID})
}
```

---

## 🧪 Tests

### Backend NestJS

```bash
cd backend

# Ejecutar tests unitarios
npm run test

# Ejecutar tests e2e
npm run test:e2e

# Ejecutar tests con cobertura
npm run test:cov
```

### Servicios Go

```bash
cd services/api-gateway

# Ejecutar tests
go test ./...

# Ejecutar tests con cobertura
go test -cover ./...
```

---

## Linting y Formatos

### Backend NestJS

```bash
cd backend

# Linting
npm run lint

# Formatear código
npm run format
```

### Servicios Go

```bash
# Formatear código Go
go fmt ./...

# Lint (requiere golangci-lint)
golangci-lint run ./...
```

---

## Estructura de Directorios - Go

```
services/
├── api-gateway/          # Puerta de entrada principal
│   ├── main.go
│   ├── config/           # Configuración
│   ├── controllers/       # Handlers HTTP
│   ├── models/           # Structs de datos
│   ├── repositories/     # Acceso a datos
│   ├── routes/           # Rutas
│   └── services/         # Lógica de negocio
├── auth-service/         # Autenticación
├── user-service/         # Gestión de usuarios
├── common/               # Código compartido
│   ├── httpresponse/
│   └── logging/
└── [otros servicios]/
```

---

## Estructura de Directorios - NestJS

```
backend/
├── src/
│   ├── app.controller.ts
│   ├── app.module.ts
│   ├── app.service.ts
│   ├── main.ts
│   ├── redis/            # Configuración Redis
│   └── [módulos]/
├── test/                 # Tests e2e
├── package.json
├── tsconfig.json
└── nest-cli.json
```

---

## Conexión a Servicios Externos

### Conexión a Redis

```typescript
// backend/src/redis/redis.module.ts
import { Module } from '@nestjs/common';
import { createClient } from 'redis';

@Module({
  providers: [
    {
      provide: 'REDIS_CLIENT',
      useValue: createClient({
        socket: {
          host: 'localhost',  // o 'redis' en Docker
          port: 6379
        },
        password: process.env.REDIS_PASSWORD
      })
    }
  ],
  exports: ['REDIS_CLIENT']
})
export class RedisModule {}
```

### Conexión a RabbitMQ

```typescript
// backend/src/rabbitmq/rabbitmq.module.ts
import { Module } from '@nestjs/common';
import * as amqp from 'amqplib';

@Module({
  providers: [
    {
      provide: 'RABBITMQ_CONNECTION',
      useFactory: async () => {
        return await amqp.connect(
          `amqp://${process.env.RABBITMQ_USER}:${process.env.RABBITMQ_PASSWORD}@localhost:5672`
        );
      }
    }
  ],
  exports: ['RABBITMQ_CONNECTION']
})
export class RabbitMQModule {}
```

---

## Problemas Comunes

### Puerto ya en uso

```bash
# Windows: Encontrar proceso
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux
lsof -i :8080
kill -9 <PID>
```

### "Cannot find module" en Go

```bash
# Limpiar cache de módulos
go clean -modcache

# Descargar módulos nuevamente
go mod download
go mod tidy
```

### NestJS no inicia con "Cannot find module"

```bash
# Limpiar node_modules
rm -r node_modules package-lock.json
npm install
```

### Servicios no pueden conectarse entre sí

- Verifica que todos están ejecutándose: `docker-compose ps`
- Verifica puertos correctos en tu código
- Usa `localhost` si ejecutas localmente, o nombre del servicio si está en Docker

### Error de conexión a Redis/RabbitMQ

```bash
# Verificar que servicios de Docker están ejecutándose
docker-compose ps redis rabbitmq

# Verificar logs
docker-compose logs redis
docker-compose logs rabbitmq

# Verificar .env tiene credenciales correctas
cat .env | grep REDIS_PASSWORD
cat .env | grep RABBITMQ_
```

---

## 🚀 Próximo Paso

Ahora que tienes el backend ejecutándose:

👉 **Ve a [5. ML Service (Python) →](05-ml-service-setup.md)**

---

## Checklist de Backend

```
□ Go 1.26.6 instalado
□ Node.js 18+ LTS instalado
□ Dependencias descargadas: go mod download
□ Dependencias NestJS: npm install (en backend/)
□ API Gateway compilado y ejecutándose
□ Auth Service compilado y ejecutándose
□ User Service compilado y ejecutándose
□ Backend NestJS ejecutándose en :3000
□ Todos los /health endpoints responden correctamente
□ Tests pasan: npm run test
□ Linting pasa: npm run lint
□ Código formateado: npm run format
```

---

**Última actualización:** Septiembre 2024  
**Versión:** 1.0

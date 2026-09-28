# Guía de Configuración del Proyecto Meloop

Bienvenido a la guía completa de configuración para el proyecto **Meloop**. Esta documentación te ayudará a configurar tu entorno de desarrollo desde cero.

## 🗺️ Índice de Documentación

Esta guía está dividida en los siguientes pasos, en orden recomendado:

| # | Documento | Descripción | Tiempo |
|---|-----------|-------------|--------|
| 1 | [Requisitos Previos](01-prerequisites.md) | Verificación e instalación de todas las herramientas necesarias | 30-45 min |
| 2 | [Configuración del Entorno](02-environment-setup.md) | Configuración de variables de entorno, Supabase y credenciales | 15-20 min |
| 3 | [Docker & Servicios](03-docker-setup.md) | Levantamiento de Docker Compose y validación de servicios | 20-30 min |
| 4 | [Backend (Go + NestJS)](04-backend-setup.md) | Instalación y ejecución de servicios Go y backend NestJS | 20-30 min |
| 5 | [ML Service (Python)](05-ml-service-setup.md) | Configuración e instalación del servicio de Machine Learning | 15-20 min |
| 6 | [Flutter App](06-flutter-setup.md) | Configuración y ejecución de la aplicación móvil | 20-30 min |
| 7 | [Inicio Rápido](07-quick-start.md) | Resumen rápido de comandos para ejecutar todo | 5 min |
| 8 | [Troubleshooting](08-troubleshooting.md) | Solución de problemas frecuentes | - |

---

## ✅ Requisitos del Sistema

| Herramienta | Versión Recomendada | Versión Mínima |
|-------------|-------------------|----------------|
| **Go** | 1.26.6 | 1.21 |
| **Python** | 3.11 | 3.10 |
| **Node.js** | 18+ (LTS) | 16 |
| **Dart/Flutter** | 3.13.0+ | 3.13.0 |
| **Docker** | Latest | 20.10+ |
| **Docker Compose** | 2.20+ | 2.10+ |
| **Git** | Latest | 2.30+ |

**SO Soportados:** Windows 11+, Linux (Ubuntu 22.04+)

---

## Primeros Pasos (5 minutos)

Si solo quieres **empezar rápidamente**, ejecuta:

```bash
# 1. Clona el repositorio
git clone https://github.com/tuorganizacion/meloop.git
cd meloop

# 2. Copia el archivo de configuración
cp .env.example .env

# 3. Abre la guía de requisitos previos
# Sigue: 01-prerequisites.md
```

---

## Checklist de Configuración

Usa este checklist para validar tu setup:

```
HERRAMIENTAS
□ Git instalado y configurado
□ Go 1.26.6 instalado
□ Python 3.11 instalado
□ Node.js 18+ instalado
□ Flutter SDK 3.13.0+ instalado
□ Docker instalado
□ Docker Compose instalado

CONFIGURACIÓN
□ Repositorio clonado
□ Archivo .env configurado
□ Credenciales de Supabase agregadas
□ MinIO buckets configurados

SERVICIOS
□ Docker Compose levantado (redis, rabbitmq, minio)
□ API Gateway ejecutando en :8080
□ Servicios Go compilados
□ Backend NestJS ejecutando en :3000
□ ML Service ejecutando en :8000
□ Flutter compilado para tu plataforma

VALIDACIÓN
□ API Gateway responde a http://localhost:8080/health
□ Redis conectado
□ RabbitMQ accesible en http://localhost:15672
□ MinIO accesible en http://localhost:9001
□ ML Service responde a http://localhost:8000/health
□ Flutter app conectada al API Gateway
```

---

## Arquitectura General

```
┌─────────────────────────────────────────────────────────────┐
│                    MELOOP ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────┐         ┌──────────────────────┐      │
│  │  Flutter App    │ HTTP    │   API Gateway        │      │
│  │  (Mobile)       │────────▶│   (Go - :8080)       │      │
│  └─────────────────┘         └──────────────────────┘      │
│                                         │                   │
│                ┌────────────────────────┼───────────┐       │
│                │                        │           │       │
│        ┌───────▼────────┐     ┌────────▼──────┐   │       │
│        │ Go Services    │     │ NestJS Backend│   │       │
│        │ - Auth         │     │ (:3000)       │   │       │
│        │ - User         │     │               │   │       │
│        │ - Chat         │     └────────────────┘   │       │
│        │ - Media        │                          │       │
│        │ - Post, Music..│                          │       │
│        └────────────────┘                          │       │
│                │                                   │       │
│        ┌───────┴──────────────────────────────────▼──┐    │
│        │         Infrastructure Services             │    │
│        ├────────────────────────────────────────────┤    │
│        │ • Redis (Caching)        - :6379           │    │
│        │ • RabbitMQ (Messaging)   - :5672, :15672   │    │
│        │ • MinIO (Storage)        - :9000, :9001    │    │
│        │ • Supabase (Auth/DB)     - External        │    │
│        │ • ML Service (Python)    - :8000           │    │
│        └───────────────────────────────────────────┘    │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## Soporte

Si encuentras problemas:

1. **Consulta la sección de [Troubleshooting](08-troubleshooting.md)** para soluciones comunes
2. **Verifica tu instalación** con los comandos de validación en cada guía
3. **Revisa los logs**: Cada sección incluye cómo revisar logs de errores
4. **Contacta al equipo de desarrollo** si el problema persiste

---

## 🚀 Próximo Paso

👉 **Comienza con [1. Requisitos Previos →](01-prerequisites.md)**

---

**Última actualización:** Septiembre 2024  
**Versión de documentación:** 1.0

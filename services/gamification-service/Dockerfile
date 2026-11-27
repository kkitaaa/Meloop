FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar los módulos necesarios
COPY services/common/go.mod services/common/go.sum ./services/common/
COPY services/gamification-service/go.mod services/gamification-service/go.sum ./services/gamification-service/

# Descargar dependencias
WORKDIR /app/services/gamification-service
RUN go mod download

# Copiar el código fuente
WORKDIR /app
COPY services/common ./services/common
COPY services/gamification-service ./services/gamification-service

# Compilar
WORKDIR /app/services/gamification-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/gamification-service .

# Imagen final
FROM alpine:latest

RUN addgroup -S -g 10001 app && adduser -S -u 10001 -G app app

WORKDIR /app

COPY --from=builder --chown=10001:10001 /out/gamification-service ./gamification-service

USER 10001:10001

EXPOSE 8089

ENTRYPOINT ["./gamification-service"]
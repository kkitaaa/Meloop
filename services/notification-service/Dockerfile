FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY services/notification-service/go.mod services/notification-service/go.sum ./services/notification-service/
COPY services/common/go.mod services/common/go.sum ./services/common/

WORKDIR /app/services/notification-service

RUN go mod download

# Copiar código fuente
WORKDIR /app
COPY services/common ./services/common
COPY services/notification-service ./services/notification-service

# Compilar
WORKDIR /app/services/notification-service

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /notification-service .

# Imagen final
FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app

COPY --from=builder --chown=app:app /notification-service ./notification-service

USER app

CMD ["./notification-service"]
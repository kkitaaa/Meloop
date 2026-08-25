FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos primero para cachear dependencias
COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/auth-service/go.mod services/auth-service/go.sum* ./services/auth-service/

# Descargar dependencias
WORKDIR /app/services/auth-service
RUN go mod download

# Copiar el código fuente
WORKDIR /app
COPY services/common ./services/common
COPY services/auth-service ./services/auth-service

# Compilar el binario
WORKDIR /app/services/auth-service
RUN go mod tidy
RUN go build -o auth-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/auth-service/auth-service .

EXPOSE 8083

CMD ["./auth-service"]

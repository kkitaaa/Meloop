FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos primero para cachear dependencias
COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/test-service/go.mod services/test-service/go.sum* ./services/test-service/

# Copiar el código fuente
COPY services/common ./services/common
COPY services/test-service ./services/test-service

# Compilar el binario
WORKDIR /app/services/test-service
RUN go build -o test-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/test-service/test-service .

EXPOSE 8081

CMD ["./test-service"]
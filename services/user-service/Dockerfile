FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos primero para cachear dependencias
COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/user-service/go.mod services/user-service/go.sum* ./services/user-service/

# Descargar dependencias
WORKDIR /app/services/user-service
RUN go mod download

# Copiar el código fuente
WORKDIR /app
COPY services/common ./services/common
COPY services/user-service ./services/user-service

# Compilar el binario
WORKDIR /app/services/user-service
RUN go mod tidy
RUN go build -o user-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/user-service/user-service .

EXPOSE 8082

CMD ["./user-service"]

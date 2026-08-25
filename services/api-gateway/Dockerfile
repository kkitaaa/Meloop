FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos primero para cachear dependencias
COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/api-gateway/go.mod services/api-gateway/go.sum* ./services/api-gateway/

# Descargar dependencias externos
WORKDIR /app/services/api-gateway
RUN go mod download

# Copiar el código fuente
WORKDIR /app
COPY services/common ./services/common
COPY services/api-gateway ./services/api-gateway

# Compilar el binario
WORKDIR /app/services/api-gateway
RUN go mod tidy
RUN go build -o api-gateway .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/api-gateway/api-gateway .

EXPOSE 8080

ENV PORT=8080

CMD ["./api-gateway"]
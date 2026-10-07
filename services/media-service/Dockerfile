FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos primero para cachear dependencias
COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/media-service/go.mod services/media-service/go.sum* ./services/media-service/

# Descargar dependencias
WORKDIR /app/services/media-service
RUN go mod download

# Copiar el código fuente
WORKDIR /app
COPY services/common ./services/common
COPY services/media-service ./services/media-service

# Compilar el binario
WORKDIR /app/services/media-service
RUN go mod tidy
RUN go build -o media-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/media-service/media-service .

EXPOSE 8085

CMD ["./media-service"]

FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copiar archivos de módulos primero para cachear dependencias
COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY infrastructure/go.mod infrastructure/go.sum* ./infrastructure/
COPY services/post-service/go.mod services/post-service/go.sum* ./services/post-service/

# Descargar dependencias
WORKDIR /app/services/post-service
RUN go mod download

# Copiar el código fuente necesario
WORKDIR /app
COPY services/common ./services/common
COPY infrastructure ./infrastructure
COPY services/post-service ./services/post-service

# Compilar binario estático para Linux
WORKDIR /app/services/post-service
RUN CGO_ENABLED=0 GOOS=linux go build -o post-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/services/post-service/post-service .

EXPOSE 8084

CMD ["./post-service"]
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY services/recommendation-service/go.mod services/recommendation-service/go.sum ./services/recommendation-service/
WORKDIR /app/services/recommendation-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/recommendation-service ./services/recommendation-service

WORKDIR /app/services/recommendation-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /recommendation-service .

FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /recommendation-service ./recommendation-service

EXPOSE 8087

USER app

CMD ["./recommendation-service"]

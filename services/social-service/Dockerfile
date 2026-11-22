FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/social-service/go.mod services/social-service/go.sum ./services/social-service/

WORKDIR /app/services/social-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/social-service ./services/social-service

WORKDIR /app/services/social-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /social-service .

FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /social-service ./social-service

EXPOSE 8086

USER app

CMD ["./social-service"]

FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum ./services/common/
COPY services/moderation-service/go.mod services/moderation-service/go.sum ./services/moderation-service/

WORKDIR /app/services/moderation-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/moderation-service ./services/moderation-service

WORKDIR /app/services/moderation-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /moderation-service .

FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /moderation-service ./moderation-service

EXPOSE 8089
USER app

CMD ["./moderation-service"]

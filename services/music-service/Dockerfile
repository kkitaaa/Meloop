FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum ./services/common/
COPY services/music-service/go.mod services/music-service/go.sum ./services/music-service/

WORKDIR /app/services/music-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/music-service ./services/music-service

WORKDIR /app/services/music-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /music-service .

FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /music-service ./music-service

EXPOSE 8091
ENV PORT=8091

USER app

CMD ["./music-service"]

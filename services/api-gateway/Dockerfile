FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum ./services/common/
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/

WORKDIR /app/services/api-gateway
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/api-gateway ./services/api-gateway

WORKDIR /app/services/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /api-gateway .

FROM alpine:latest

RUN apk add --no-cache ca-certificates \
    && addgroup -S -g 10001 app \
    && adduser -S -u 10001 -G app app

WORKDIR /app
COPY --from=builder --chown=app:app /api-gateway ./api-gateway

EXPOSE 8080

ENV PORT=8080

USER app

CMD ["./api-gateway"]
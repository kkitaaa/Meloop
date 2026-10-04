FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/notification-service/go.mod services/notification-service/go.sum* ./services/notification-service/

WORKDIR /app/services/notification-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/notification-service ./services/notification-service

WORKDIR /app/services/notification-service
RUN go mod tidy
RUN go build -o notification-service .

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/services/notification-service/notification-service .

EXPOSE 8089

CMD ["./notification-service"]
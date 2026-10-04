FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/gamification-service/go.mod services/gamification-service/go.sum* ./services/gamification-service/

WORKDIR /app/services/gamification-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/gamification-service ./services/gamification-service

WORKDIR /app/services/gamification-service
RUN go mod tidy
RUN go build -o gamification-service .

FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/services/gamification-service/gamification-service .

EXPOSE 8088

CMD ["./gamification-service"]
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum* ./services/common/
COPY services/social-service/go.mod services/social-service/go.sum* ./services/social-service/

WORKDIR /app/services/social-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/social-service ./services/social-service

WORKDIR /app/services/social-service
RUN go mod tidy
RUN go build -o social-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/services/social-service/social-service .

EXPOSE 8086

CMD ["./social-service"]

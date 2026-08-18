FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go build -o test-service .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/test-service .

EXPOSE 8081

CMD ["./test-service"]
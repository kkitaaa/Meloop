FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY services/common/go.mod services/common/go.sum ./services/common/
COPY services/social-service/go.mod services/social-service/go.sum ./services/social-service/

WORKDIR /app/services/social-service
RUN go mod download

WORKDIR /app
COPY services/common ./services/common
COPY services/social-service ./services/social-service

WORKDIR /app/services/social-service
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /social-service .

FROM scratch

COPY --from=builder /social-service /social-service

USER 65532:65532
EXPOSE 8086

ENTRYPOINT ["/social-service"]

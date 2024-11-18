FROM golang:1.22-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o /app/main ./cmd/main.go

# Финальный минималистичный образ
FROM alpine:3.15

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/.env .
COPY --from=builder /app/prometheus.yaml /prometheus.yaml
COPY --from=builder /app/config/local.yaml ./config/local.yaml

CMD ["/app/main"]

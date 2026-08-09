FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o garuda-api ./cmd/garuda-api

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl postgresql-client
WORKDIR /root/
COPY --from=builder /app/garuda-api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/internal/api/openapi.yaml ./internal/api/openapi.yaml
COPY build/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]
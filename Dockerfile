# Stage 1: Build static Go binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Cache module layers
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .

# Compile fully static, stripped binary with release metadata
ARG VERSION=v1.0.0
ARG COMMIT=release
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -extldflags '-static'" \
    -o /out/garuda-api \
    ./cmd/garuda-api

# Stage 2: Minimal non-root scratch runtime
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/garuda-api /garuda-api
COPY migrations /migrations

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/garuda-api"]

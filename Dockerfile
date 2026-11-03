# Stage 1: Build all 4 static zero-CGO Garuda binaries
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

# Cache module layers
COPY go.mod go.sum ./
RUN go mod download

# Copy source tree
COPY . .

# Compile all 4 fully static, stripped binaries with version metadata
ARG VERSION=v1.9.11
ARG COMMIT=release
RUN mkdir -p /out && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/garuda ./cmd/garuda && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/garuda-api ./cmd/garuda-api && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/garuda-mcp ./cmd/garuda-mcp && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" -o /out/garuda-talon ./cmd/garuda-talon

# Stage 2: Minimal non-root runtime containing CLI, API, MCP Server & Talon Harness
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /out/garuda /garuda
COPY --from=builder /out/garuda-api /garuda-api
COPY --from=builder /out/garuda-mcp /garuda-mcp
COPY --from=builder /out/garuda-talon /garuda-talon
COPY --from=builder /out/garuda /usr/local/bin/garuda
COPY --from=builder /out/garuda-api /usr/local/bin/garuda-api
COPY --from=builder /out/garuda-mcp /usr/local/bin/garuda-mcp
COPY --from=builder /out/garuda-talon /usr/local/bin/garuda-talon
COPY capabilities.json /capabilities.json
COPY version.txt /version.txt
COPY migrations /migrations
COPY policies /policies

USER 65532:65532
ENV GARUDA_BIND=0.0.0.0:8080
EXPOSE 8080

ENTRYPOINT ["/garuda-api"]

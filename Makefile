
# -----------------------------------------------------------------------------
# Build & Benchmark Targets (Garuda v1.5 + Garuda Talon)
# -----------------------------------------------------------------------------
.PHONY: build test benchmark benchmark-strict benchmark-record talon-init talon-guard

build:
	mkdir -p bin
	go build -o bin/garuda ./cmd/garuda
	go build -o bin/garuda-api ./cmd/garuda-api
	go build -o bin/garuda-mcp ./cmd/garuda-mcp
	go build -o bin/garuda-talon ./cmd/garuda-talon

test:
	go test -v -race ./internal/...

talon-init: build
	./bin/garuda talon init .

talon-guard: build
	./bin/garuda talon guard . --fail-on-block

benchmark:
	go run scripts/run_benchmarks.go

benchmark-strict:
	go run scripts/run_benchmarks.go --strict

benchmark-record:
	go run scripts/run_benchmarks.go --record-baseline

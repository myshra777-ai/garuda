# -----------------------------------------------------------------------------
# Build, Release Matrix & Benchmark Targets (Garuda v1.9.1 + Garuda Talon)
# -----------------------------------------------------------------------------
.PHONY: build release-matrix test benchmark benchmark-strict benchmark-record talon-init talon-guard

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/ ./cmd/garuda ./cmd/garuda-api ./cmd/garuda-mcp ./cmd/garuda-talon

release-matrix:
	chmod +x scripts/build_release_matrix.sh
	./scripts/build_release_matrix.sh dist

test:
	go test -v -race ./internal/... ./cmd/...

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


# -----------------------------------------------------------------------------
# Benchmark & Ground-Truth Regression Targets
# -----------------------------------------------------------------------------
.PHONY: benchmark benchmark-strict benchmark-record

benchmark:
	go run scripts/run_benchmarks.go

benchmark-strict:
	go run scripts/run_benchmarks.go --strict

benchmark-record:
	go run scripts/run_benchmarks.go --record-baseline

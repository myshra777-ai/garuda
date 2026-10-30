#!/usr/bin/env bash
# scripts/benchmark_v2.sh
# Usage: bash scripts/benchmark_v2.sh <trial_count>

TRIALS=${1:-10}
OUT_DIR="benchmark_v2_results"
mkdir -p "$OUT_DIR"

MODEL="gpt-5.6-terra" # Pin the exact model
QUESTION="List the number of function entities reachable from SQLModel within three hops, and give the exact call chains for main and test_deprecated_from_orm_inheritance. Be concise."

GARUDA_PROMPT="$QUESTION
Before answering, you MUST invoke the Garuda MCP tool 'garuda.blast_radius' for SQLModel at depth 3. Use its returned path evidence as the source of truth."

echo "Starting $TRIALS trials per arm..."

for i in $(seq 1 $TRIALS); do
    echo "--- Trial $i ---"
    
    # 1. Control Arm (Raw)
    codex exec --ephemeral --ignore-rules --json --color never \
        --model "$MODEL" \
        --output-last-message "$OUT_DIR/raw_${i}.md" \
        "$QUESTION" > "$OUT_DIR/raw_${i}.transcript.log" 2>&1
        
    # 2. Treatment Arm (Garuda)
    codex exec --ephemeral --ignore-rules --json --color never \
        --model "$MODEL" \
        --output-last-message "$OUT_DIR/garuda_${i}.md" \
        "$GARUDA_PROMPT" > "$OUT_DIR/garuda_${i}.transcript.log" 2>&1
done

echo "Trials complete. Extracting telemetry..."
#!/usr/bin/env bash
set -e

TOKEN=$(curl -s http://localhost:8080/debug/token | jq -r '.token // .access_token // empty')
[ -z "$TOKEN" ] && TOKEN=$(curl -s http://localhost:8080/debug/token | tr -d '"\r\n ')

CONCURRENCY=20
TOTAL_REQUESTS=500
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Executing $TOTAL_REQUESTS requests with concurrency $CONCURRENCY..."
START_TIME=$(date +%s%N)

seq 1 $TOTAL_REQUESTS | xargs -P $CONCURRENCY -I {} bash -c "
  DEC_ID=\$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid)
  IDEM_KEY=\"bench-\$DEC_ID\"
  
  CODE=\$(curl -s -o /dev/null -w '%{http_code}' -X POST http://localhost:8080/api/v1/decisions \
    -H 'Authorization: Bearer $TOKEN' \
    -H 'Content-Type: application/json' \
    -d '{
      \"decision_id\": \"'\$DEC_ID'\",
      \"title\": \"Benchmark Run Decision {}\",
      \"statement\": \"Load testing Merkle hash tree commit rate\",
      \"owner\": \"bench-worker\",
      \"scope\": {\"domain\": \"bench\", \"system\": \"garuda-api\"},
      \"confidence\": 1.0,
      \"idempotency_key\": \"'\$IDEM_KEY'\"
    }')
  echo \$CODE >> '$TMP_DIR/results.log'
"

END_TIME=$(date +%s%N)
DURATION_SEC=$(awk "BEGIN {print ($END_TIME - $START_TIME) / 1000000000}")
RPS=$(awk "BEGIN {print $TOTAL_REQUESTS / $DURATION_SEC}")

echo "----------------------------------------"
echo "Duration:    ${DURATION_SEC}s"
echo "Throughput:  ${RPS} req/sec"
echo "HTTP Status Counts:"
sort "$TMP_DIR/results.log" | uniq -c
echo "----------------------------------------"

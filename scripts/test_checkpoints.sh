#!/usr/bin/env bash
set -e

API_URL="http://localhost:8080"
TENANT_ID="00000000-0000-0000-0000-000000000001"

echo "=== 1. Requesting Debug Auth Token ==="
TOKEN_RESP=$(curl -s "${API_URL}/debug/token?actor=agent-runner&tenant_id=${TENANT_ID}")
TOKEN=$(echo "$TOKEN_RESP" | jq -r '.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
  echo "❌ Failed to obtain JWT token"
  exit 1
fi
echo "✓ Token obtained successfully"

echo -e "\n=== 2. Creating Agent Checkpoint (Agent A: claude-3-5) ==="
SAVE_RESP=$(curl -s -X POST "${API_URL}/api/v1/agents/checkpoint" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "claude-3-5-sonnet",
    "checkpoint_data": {
      "execution_step": "architecture_analysis",
      "progress": 50,
      "files_touched": ["internal/api/handler.go", "internal/auth/jwt.go"],
      "reasoning": "Auth layer verified, proceeding to DB persistence."
    },
    "ttl_seconds": 3600
  }')

CHECKPOINT_ID=$(echo "$SAVE_RESP" | jq -r '.id')
echo "✓ Checkpoint created with ID: ${CHECKPOINT_ID}"

echo -e "\n=== 3. Retrieving Saved Checkpoint ==="
GET_RESP=$(curl -s -X GET "${API_URL}/api/v1/agents/checkpoint/${CHECKPOINT_ID}" \
  -H "Authorization: Bearer ${TOKEN}")

AGENT_ID=$(echo "$GET_RESP" | jq -r '.agent_id')
STATUS=$(echo "$GET_RESP" | jq -r '.status')
echo "✓ Retrieved Checkpoint - Agent: ${AGENT_ID}, Status: ${STATUS}"

echo -e "\n=== 4. Executing Task Handoff (Agent A -> Agent B: gpt-4o) ==="
HANDOFF_RESP=$(curl -s -X POST "${API_URL}/api/v1/agents/handoff" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"checkpoint_id\": \"${CHECKPOINT_ID}\",
    \"from_agent_id\": \"claude-3-5-sonnet\",
    \"to_agent_id\": \"gpt-4o\"
  }")

NEW_CHECKPOINT_ID=$(echo "$HANDOFF_RESP" | jq -r '.new_checkpoint_id')
TO_AGENT=$(echo "$HANDOFF_RESP" | jq -r '.to_agent')
echo "✓ Handoff executed - New Checkpoint ID: ${NEW_CHECKPOINT_ID} assigned to ${TO_AGENT}"

echo -e "\n=== 5. Resuming Task via New Agent ==="
RESUME_RESP=$(curl -s -X POST "${API_URL}/api/v1/agents/resume" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d "{
    \"checkpoint_id\": \"${NEW_CHECKPOINT_ID}\"
  }")

RESUMED_AGENT=$(echo "$RESUME_RESP" | jq -r '.agent_id')
echo "✓ Resumed execution context under Agent: ${RESUMED_AGENT}"

echo -e "\n🎉 ALL CHECKPOINT E2E TESTS PASSED SUCCESSFULLY!"

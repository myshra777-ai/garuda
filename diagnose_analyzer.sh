#!/bin/bash
# save as diagnose_analyzer.sh
# chmod +x diagnose_analyzer.sh
# ./diagnose_analyzer.sh

OUTPUT="analyzer_diagnosis.txt"
echo "=== GARUDA ANALYZER DIAGNOSIS ===" > "$OUTPUT"
echo "Date: $(date)" >> "$OUTPUT"
echo "" >> "$OUTPUT"

echo "1. Go version" >> "$OUTPUT"
go version >> "$OUTPUT" 2>&1
echo "" >> "$OUTPUT"

echo "2. Garuda version / commit" >> "$OUTPUT"
git rev-parse HEAD 2>/dev/null >> "$OUTPUT"
echo "" >> "$OUTPUT"

echo "3. Analyzer source (ast_extractor.go)" >> "$OUTPUT"
cat internal/analyzer/ast_extractor.go >> "$OUTPUT"
echo "" >> "$OUTPUT"

echo "4. Sample analysis of a single file (to verify AST traversal)" >> "$OUTPUT"
# Create a small test file to isolate the issue
mkdir -p /tmp/garuda_test
cat > /tmp/garuda_test/main.go << 'EOF'
package main

import (
    "fmt"
    "net/http"
)

func main() {
    fmt.Println("hello")
    http.Get("https://example.com")
}
EOF

# Run the analyzer on this tiny codebase
go run cmd/garuda/main.go analyze /tmp/garuda_test -o /tmp/test.json
echo "Analyzer output:" >> "$OUTPUT"
cat /tmp/test.json >> "$OUTPUT"
echo "" >> "$OUTPUT"

echo "5. Count of relationships in test" >> "$OUTPUT"
grep -c '"type":"IMPORTS"' /tmp/test.json >> "$OUTPUT"
grep -c '"type":"CALLS"' /tmp/test.json >> "$OUTPUT"
echo "" >> "$OUTPUT"

echo "6. All relationship types present in test" >> "$OUTPUT"
grep -E '"type":[[:space:]]*"[A-Z_]+"' /tmp/test.json | head -20 >> "$OUTPUT"
echo "" >> "$OUTPUT"

echo "7. PostgreSQL schema (tables, columns)" >> "$OUTPUT"
psql -U test -h localhost -p 5433 -d garuda_test -c "\dt" >> "$OUTPUT" 2>&1
psql -U test -h localhost -p 5433 -d garuda_test -c "\d entities" >> "$OUTPUT" 2>&1
psql -U test -h localhost -p 5433 -d garuda_test -c "\d claims" >> "$OUTPUT" 2>&1
echo "" >> "$OUTPUT"

echo "8. Current database counts" >> "$OUTPUT"
psql -U test -h localhost -p 5433 -d garuda_test -c "SELECT COUNT(*) FROM entities;" >> "$OUTPUT" 2>&1
psql -U test -h localhost -p 5433 -d garuda_test -c "SELECT COUNT(*) FROM claims;" >> "$OUTPUT" 2>&1
echo "" >> "$OUTPUT"

echo "Diagnosis written to $OUTPUT"
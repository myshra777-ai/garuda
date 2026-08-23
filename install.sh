#!/usr/bin/env bash
# Copyright 2026 Rohit Mishra
# SPDX-License-Identifier: Apache-2.0
set -e

echo "🦅 Installing Garuda Epistemic Intelligence..."

# 1. Build or download binary
mkdir -p bin
go build -o bin/garuda ./cmd/garuda

# 2. Run instant 1-click initialization
./bin/garuda init

echo ""
echo "✨ Installation complete! You can now ask Claude Desktop or Cursor:"
echo "   'Show me all active contradictions in this codebase using Garuda MCP'"

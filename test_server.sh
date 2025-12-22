#!/bin/bash

# Start server with minimal config
export DATABASE_URL="postgres://test:test@localhost:5432/test"
export OPENAI_API_KEY="test-key"
export LLM_PROVIDER="openai"
export PORT=8081

# Start server in background
go run cmd/server/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Test health endpoint
RESPONSE=$(curl -s http://localhost:8081/health)
EXIT_CODE=$?

# Kill server
kill $SERVER_PID 2>/dev/null

# Check response
if [ $EXIT_CODE -eq 0 ] && echo "$RESPONSE" | grep -q '"status":"ok"'; then
    echo "✅ Health check test passed"
    echo "Response: $RESPONSE"
    exit 0
else
    echo "❌ Health check test failed"
    echo "Response: $RESPONSE"
    exit 1
fi

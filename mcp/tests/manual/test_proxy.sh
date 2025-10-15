#!/bin/bash
# Manual test script for SlimMCP proxy
# This script tests both User Story 1 (automatic filtering) and User Story 2 (explicit field selection)

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
PROXY_URL="http://localhost:8080"
MOCK_SERVER_URL="http://localhost:3000"

echo -e "${BLUE}=== SlimMCP Proxy Test Suite ===${NC}\n"

# Function to make a request and display results
test_request() {
    local test_name="$1"
    local request_file="$2"
    local expected="$3"

    echo -e "${BLUE}Test: ${test_name}${NC}"
    echo "Request:"
    cat "$request_file" | jq '.' 2>/dev/null || cat "$request_file"
    echo ""

    echo "Response:"
    response=$(curl -s -X POST "${PROXY_URL}/mock" \
        -H "Content-Type: application/json" \
        -d @"$request_file")

    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    echo ""

    # Check for expected content
    if [ -n "$expected" ]; then
        if echo "$response" | grep -q "$expected"; then
            echo -e "${GREEN}✓ PASS${NC}\n"
        else
            echo -e "${RED}✗ FAIL: Expected '$expected' not found${NC}\n"
        fi
    else
        echo -e "${GREEN}✓ Completed${NC}\n"
    fi

    sleep 1
}

# Check if proxy is running
echo "Checking if proxy is running..."
if ! curl -s "${PROXY_URL}/health" > /dev/null 2>&1; then
    echo -e "${RED}Error: Proxy is not running on ${PROXY_URL}${NC}"
    echo "Please start the proxy with: ./bin/slim -config examples/dev.yaml"
    exit 1
fi
echo -e "${GREEN}✓ Proxy is running${NC}\n"

# Check if mock server is running
echo "Checking if mock server is running..."
if ! curl -s "${MOCK_SERVER_URL}" > /dev/null 2>&1; then
    echo -e "${RED}Error: Mock server is not running on ${MOCK_SERVER_URL}${NC}"
    echo "Please start the mock server in another terminal"
    exit 1
fi
echo -e "${GREEN}✓ Mock server is running${NC}\n"

# Create test requests directory
TEST_DIR="tests/manual/requests"
mkdir -p "$TEST_DIR"

# Test 1: Health Check
echo -e "${BLUE}=== Test 1: Health Check ===${NC}"
curl -s "${PROXY_URL}/health" | jq '.'
echo -e "${GREEN}✓ PASS${NC}\n"

# Test 2: User Story 1 - Automatic Filtering (Minimal Strategy)
cat > "${TEST_DIR}/us1_minimal.json" << 'EOF'
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_database",
    "arguments": {
      "database_id": "db123456"
    }
  },
  "id": 1
}
EOF

test_request "User Story 1: Automatic Filtering (No _meta)" \
    "${TEST_DIR}/us1_minimal.json" \
    "\"strategy\":\"smart\""

# Test 3: User Story 2 - Explicit Include
cat > "${TEST_DIR}/us2_include.json" << 'EOF'
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_database",
    "arguments": {
      "database_id": "db123456"
    }
  },
  "_meta": {
    "fieldSelection": {
      "include": ["id", "title", "url"],
      "strategy": "explicit"
    }
  },
  "id": 2
}
EOF

test_request "User Story 2: Explicit Include Fields" \
    "${TEST_DIR}/us2_include.json" \
    "\"strategy\":\"explicit\""

# Test 4: User Story 2 - Explicit Exclude
cat > "${TEST_DIR}/us2_exclude.json" << 'EOF'
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_database",
    "arguments": {
      "database_id": "db123456"
    }
  },
  "_meta": {
    "fieldSelection": {
      "exclude": ["icon", "cover", "created_by"],
      "strategy": "explicit"
    }
  },
  "id": 3
}
EOF

test_request "User Story 2: Explicit Exclude Fields" \
    "${TEST_DIR}/us2_exclude.json" \
    "\"strategy\":\"explicit\""

# Test 5: User Story 2 - Nested Field Selection
cat > "${TEST_DIR}/us2_nested.json" << 'EOF'
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_database",
    "arguments": {
      "database_id": "db123456"
    }
  },
  "_meta": {
    "fieldSelection": {
      "include": ["results", "results.id", "results.properties.Status", "results.properties.Priority"],
      "strategy": "explicit"
    }
  },
  "id": 4
}
EOF

test_request "User Story 2: Nested Field Selection (Dot Notation)" \
    "${TEST_DIR}/us2_nested.json" \
    "properties"

# Test 6: Invalid Request - Both Include and Exclude
cat > "${TEST_DIR}/invalid_both.json" << 'EOF'
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "query_database",
    "arguments": {
      "database_id": "db123456"
    }
  },
  "_meta": {
    "fieldSelection": {
      "include": ["title"],
      "exclude": ["icon"],
      "strategy": "explicit"
    }
  },
  "id": 5
}
EOF

echo -e "${BLUE}Test: Invalid Request (Both Include and Exclude)${NC}"
echo "Request:"
cat "${TEST_DIR}/invalid_both.json" | jq '.'
echo ""
echo "Response:"
response=$(curl -s -X POST "${PROXY_URL}/mock" \
    -H "Content-Type: application/json" \
    -d @"${TEST_DIR}/invalid_both.json")
echo "$response" | jq '.'
echo ""

if echo "$response" | grep -q "error"; then
    echo -e "${GREEN}✓ PASS: Error returned as expected${NC}\n"
else
    echo -e "${RED}✗ FAIL: Should have returned an error${NC}\n"
fi

# Test 7: Check Filtering Metadata
echo -e "${BLUE}=== Test 7: Verify Filtering Metadata ===${NC}"
response=$(curl -s -X POST "${PROXY_URL}/mock" \
    -H "Content-Type: application/json" \
    -d @"${TEST_DIR}/us1_minimal.json")

echo "Checking for filtering metadata in response..."
echo "$response" | jq '._meta.fieldSelection'
echo ""

if echo "$response" | jq '._meta.fieldSelection.savings' | grep -qE '^[0-9]+'; then
    savings=$(echo "$response" | jq -r '._meta.fieldSelection.savings')
    echo -e "${GREEN}✓ PASS: Token savings reported: ${savings}%${NC}\n"
else
    echo -e "${RED}✗ FAIL: No savings metadata found${NC}\n"
fi

# Summary
echo -e "${BLUE}=== Test Summary ===${NC}"
echo "All manual tests completed!"
echo ""
echo "To verify the results:"
echo "1. Check that responses contain only requested fields"
echo "2. Check that _meta.fieldSelection contains filtering statistics"
echo "3. Check that token savings are calculated"
echo ""
echo -e "${GREEN}Testing complete!${NC}"

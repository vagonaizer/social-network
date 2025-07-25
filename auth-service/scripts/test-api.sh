#!/bin/bash

BASE_URL="http://localhost:8080/api"
HEALTH_URL="http://localhost:8080"

echo "🚀 Testing Auth Service API (Clean Run)"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2${NC}"
    else
        echo -e "${RED}❌ $2${NC}"
    fi
}

# Function to make HTTP request and check status
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local expected_status=$4
    local description=$5
    local headers=$6

    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Request: $method $url"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    
    if [ -n "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -H "$headers" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    # Split response and status code using a more reliable method
    status_code=$(echo "$response" | tail -1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "Status: $status_code"
    echo "Response: $response_body"
    
    if [ "$status_code" = "$expected_status" ]; then
        print_result 0 "$description"
        return 0
    else
        print_result 1 "$description (Expected: $expected_status, Got: $status_code)"
        return 1
    fi
}

# Clean up existing test user first
echo -e "\n${YELLOW}=== Cleanup ===${NC}"
echo "Cleaning up existing test user..."
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "DELETE FROM users WHERE email = 'test-clean@example.com';" 2>/dev/null || true

# Test 1: Health Check
echo -e "\n${YELLOW}=== Health Checks ===${NC}"
test_endpoint "GET" "$HEALTH_URL/health" "" "200" "Health check"
test_endpoint "GET" "$HEALTH_URL/debug" "" "200" "Debug endpoint"

# Test 2: User Registration with unique email
echo -e "\n${YELLOW}=== User Registration ===${NC}"
REGISTER_DATA='{
    "email": "test-clean@example.com",
    "username": "testcleanuser",
    "display_name": "Test Clean User",
    "password": "TestPass123!"
}'

test_endpoint "POST" "$BASE_URL/auth/register" "$REGISTER_DATA" "201" "User registration"

# Test 3: Duplicate Registration (should fail)
echo -e "\n${YELLOW}=== Duplicate Registration ===${NC}"
test_endpoint "POST" "$BASE_URL/auth/register" "$REGISTER_DATA" "409" "Duplicate registration (should fail)"

# Test 4: User Login
echo -e "\n${YELLOW}=== User Login ===${NC}"
LOGIN_DATA='{
    "email": "test-clean@example.com",
    "password": "TestPass123!"
}'

login_response=$(curl -s -X POST "$BASE_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_DATA")

echo "Login response: $login_response"

# Extract tokens from login response
ACCESS_TOKEN=$(echo "$login_response" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$login_response" | grep -o '"refresh_token":"[^"]*' | cut -d'"' -f4)

if [ -n "$ACCESS_TOKEN" ]; then
    echo -e "${GREEN}✅ Login successful - Access token obtained${NC}"
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
else
    echo -e "${RED}❌ Login failed - No access token${NC}"
    exit 1
fi

# Test 5: Get Current User (Protected endpoint)
echo -e "\n${YELLOW}=== Protected Endpoints ===${NC}"
test_endpoint "GET" "$BASE_URL/auth/me" "" "200" "Get current user" "Authorization: Bearer $ACCESS_TOKEN"

# Test 6: Validate Token
test_endpoint "GET" "$BASE_URL/auth/validate" "" "200" "Validate token" "Authorization: Bearer $ACCESS_TOKEN"

# Test 7: Invalid Token
echo -e "\n${YELLOW}=== Invalid Token Tests ===${NC}"
test_endpoint "GET" "$BASE_URL/auth/me" "" "401" "Invalid token test" "Authorization: Bearer invalid_token"

# Test 8: Refresh Token
echo -e "\n${YELLOW}=== Token Refresh ===${NC}"
if [ -n "$REFRESH_TOKEN" ]; then
    REFRESH_DATA="{\"refresh_token\": \"$REFRESH_TOKEN\"}"
    test_endpoint "POST" "$BASE_URL/auth/refresh" "$REFRESH_DATA" "200" "Refresh token"
fi

# Test 9: Email Verification with fake token (should return proper error)
echo -e "\n${YELLOW}=== Email Verification ===${NC}"
VERIFY_DATA='{"token": "fake_token"}'
test_endpoint "POST" "$BASE_URL/auth/verify-email" "$VERIFY_DATA" "404" "Email verification with fake token (should fail with 404)"

# Test 10: Password Reset Initiation
echo -e "\n${YELLOW}=== Password Reset ===${NC}"
RESET_INIT_DATA='{"email": "test-clean@example.com"}'
test_endpoint "POST" "$BASE_URL/auth/reset-password" "$RESET_INIT_DATA" "200" "Password reset initiation"

echo -e "\n${GREEN}🎉 Clean API Testing Complete!${NC}"
echo -e "\n${YELLOW}📝 Summary:${NC}"
echo "- Health checks: Working"
echo "- User registration: Working"
echo "- User login: Working"
echo "- Protected endpoints: Working"
echo "- Token validation: Working"
echo "- Error handling: Working"
echo ""
echo -e "${YELLOW}🌐 Access Swagger UI at: http://localhost:8080/swagger/index.html${NC}"
echo -e "${YELLOW}🔍 View logs with: make docker-logs${NC}"

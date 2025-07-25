#!/bin/bash

echo "🔌 Testing gRPC Service (Fixed)"
echo "=============================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check if grpcurl is installed
if ! command -v grpcurl &> /dev/null; then
    echo -e "${YELLOW}Installing grpcurl...${NC}"
    if command -v brew &> /dev/null; then
        brew install grpcurl
    elif command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y grpcurl
    else
        echo -e "${RED}Please install grpcurl manually${NC}"
        echo "Visit: https://github.com/fullstorydev/grpcurl"
        exit 1
    fi
fi

GRPC_HOST="localhost:9090"

echo -e "\n${YELLOW}Testing gRPC server connection...${NC}"

# Wait for gRPC server to be ready
echo -e "${YELLOW}Waiting for gRPC server to be ready...${NC}"
sleep 5

# List available services
echo -e "\n${YELLOW}Available gRPC services:${NC}"
if grpcurl -plaintext $GRPC_HOST list; then
    echo -e "${GREEN}✅ gRPC reflection working!${NC}"
else
    echo -e "${RED}❌ gRPC reflection not working${NC}"
    echo "Trying to test without reflection..."
fi

# Clean up existing test user first
echo -e "\n${YELLOW}Cleaning up existing gRPC test user...${NC}"
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "DELETE FROM users WHERE email = 'grpc-test@example.com';" 2>/dev/null || true

# Test user registration via gRPC
echo -e "\n${YELLOW}Testing gRPC user registration...${NC}"
if grpcurl -plaintext -d '{
    "email": "grpc-test@example.com",
    "username": "grpctestuser",
    "display_name": "gRPC Test User",
    "password": "GrpcPass123!"
}' $GRPC_HOST auth.v1.AuthService/Register; then
    echo -e "${GREEN}✅ gRPC registration successful!${NC}"
else
    echo -e "${RED}❌ gRPC registration failed${NC}"
fi

# Test user login via gRPC
echo -e "\n${YELLOW}Testing gRPC user login...${NC}"
if grpcurl -plaintext -d '{
    "email": "grpc-test@example.com",
    "password": "GrpcPass123!"
}' $GRPC_HOST auth.v1.AuthService/Login; then
    echo -e "${GREEN}✅ gRPC login successful!${NC}"
else
    echo -e "${RED}❌ gRPC login failed${NC}"
fi

echo -e "\n${GREEN}🎉 gRPC testing complete!${NC}"

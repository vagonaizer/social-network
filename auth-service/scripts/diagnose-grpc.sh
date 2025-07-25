#!/bin/bash

echo "🔍 Diagnosing gRPC Service"
echo "========================="

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

# Check if container is running
echo -e "\n${YELLOW}Checking if auth-service container is running...${NC}"
if docker ps | grep -q auth_service_app; then
    echo -e "${GREEN}✅ Container is running${NC}"
    
    # Show container logs
    echo -e "\n${YELLOW}Last 20 lines of container logs:${NC}"
    docker logs auth_service_app --tail 20
else
    echo -e "${RED}❌ Container is not running${NC}"
    exit 1
fi

# Check if gRPC port is open
echo -e "\n${YELLOW}Checking if gRPC port is open...${NC}"
if nc -z localhost 9090; then
    echo -e "${GREEN}✅ Port 9090 is open${NC}"
else
    echo -e "${RED}❌ Port 9090 is not open${NC}"
    exit 1
fi

# Try to connect with grpcurl
echo -e "\n${YELLOW}Trying to connect with grpcurl...${NC}"
if grpcurl -plaintext -v $GRPC_HOST list; then
    echo -e "${GREEN}✅ Successfully connected to gRPC server${NC}"
else
    echo -e "${RED}❌ Failed to connect to gRPC server${NC}"
    
    # Check if reflection is enabled in the code
    echo -e "\n${YELLOW}Checking if reflection is enabled in the code...${NC}"
    if grep -q "reflection.Register" transport/grpc/server.go; then
        echo -e "${GREEN}✅ Reflection is enabled in the code${NC}"
    else
        echo -e "${RED}❌ Reflection is not enabled in the code${NC}"
    fi
    
    # Check if the server is listening on the correct port
    echo -e "\n${YELLOW}Checking if server is listening on port 9090...${NC}"
    if docker exec auth_service_app netstat -tulpn | grep -q ":9090"; then
        echo -e "${GREEN}✅ Server is listening on port 9090${NC}"
    else
        echo -e "${RED}❌ Server is not listening on port 9090${NC}"
    fi
fi

# Check if proto files are properly generated
echo -e "\n${YELLOW}Checking if proto files are properly generated...${NC}"
if [ -d "pkg/api/auth/v1" ]; then
    echo -e "${GREEN}✅ Proto directory exists${NC}"
    
    # List proto files
    echo -e "\n${YELLOW}Proto files:${NC}"
    ls -la pkg/api/auth/v1
else
    echo -e "${RED}❌ Proto directory does not exist${NC}"
fi

echo -e "\n${YELLOW}Checking imports in gRPC server...${NC}"
grep -n "import" transport/grpc/server.go

echo -e "\n${GREEN}🎉 Diagnosis complete!${NC}"

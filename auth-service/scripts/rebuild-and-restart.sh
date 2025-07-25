#!/bin/bash

echo "🔄 Rebuilding and restarting services"
echo "===================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Stop the auth service
echo -e "\n${YELLOW}Stopping auth service...${NC}"
docker-compose stop auth-service

# Rebuild the auth service
echo -e "\n${YELLOW}Rebuilding auth service...${NC}"
docker-compose build auth-service

# Start the auth service
echo -e "\n${YELLOW}Starting auth service...${NC}"
docker-compose up -d auth-service

# Wait for service to be ready
echo -e "\n${YELLOW}Waiting for service to be ready...${NC}"
for i in {1..10}; do
    echo "Attempt $i: Checking if service is ready..."
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        echo -e "${GREEN}✅ Service is ready!${NC}"
        break
    fi
    
    if [ $i -eq 10 ]; then
        echo -e "${RED}❌ Service failed to start properly${NC}"
        exit 1
    fi
    
    echo "Service not ready yet, waiting..."
    sleep 3
done

echo -e "\n${GREEN}🎉 Rebuild and restart complete!${NC}"

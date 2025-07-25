#!/bin/bash

echo "🗄️  Testing Database Connection and Structure"
echo "============================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Test database connection
echo -e "\n${YELLOW}Testing database connection...${NC}"
if docker exec auth_service_db pg_isready -U postgres; then
    echo -e "${GREEN}✅ Database is ready${NC}"
else
    echo -e "${RED}❌ Database connection failed${NC}"
    exit 1
fi

# Show database info
echo -e "\n${YELLOW}Database Information:${NC}"
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "
SELECT 
    current_database() as database_name,
    current_user as current_user,
    version() as postgresql_version;
"

# Show all tables
echo -e "\n${YELLOW}Database Tables:${NC}"
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "\dt"

# Show table structures
echo -e "\n${YELLOW}Table Structures:${NC}"

tables=("users" "user_auth" "user_roles" "refresh_tokens" "email_verifications" "password_resets")

for table in "${tables[@]}"; do
    echo -e "\n${YELLOW}Structure of $table:${NC}"
    docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "\d $table"
done

# Show migration status
echo -e "\n${YELLOW}Migration Status:${NC}"
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "
SELECT * FROM schema_migrations ORDER BY version;
" 2>/dev/null || echo "No migration table found"

# Show sample data (if any)
echo -e "\n${YELLOW}Sample Data:${NC}"
for table in "${tables[@]}"; do
    # Check if table exists first
    table_exists=$(docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" 2>/dev/null | tr -d ' ')
    
    if [ "$table_exists" = "t" ]; then
        count=$(docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | tr -d ' ')
        if [ -n "$count" ] && [ "$count" -gt 0 ]; then
            echo -e "${GREEN}$table: $count records${NC}"
            docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "SELECT * FROM $table LIMIT 3;"
        else
            echo -e "${YELLOW}$table: 0 records${NC}"
        fi
    else
        echo -e "${RED}$table: table does not exist${NC}"
    fi
done

echo -e "\n${GREEN}🎉 Database testing complete!${NC}"

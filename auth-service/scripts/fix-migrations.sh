#!/bin/bash

echo "🔧 Fixing Database Migrations"
echo "============================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "\n${YELLOW}Current migration status:${NC}"
make migrate-version

echo -e "\n${YELLOW}Checking missing tables...${NC}"
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "\dt"

echo -e "\n${YELLOW}Running all migrations again...${NC}"
make migrate-up

echo -e "\n${YELLOW}Final table list:${NC}"
docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "\dt"

# Check if all expected tables exist
expected_tables=("users" "user_auth" "user_roles" "refresh_tokens" "email_verifications" "password_resets")
missing_tables=()

for table in "${expected_tables[@]}"; do
    exists=$(docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" | tr -d ' ')
    if [ "$exists" != "t" ]; then
        missing_tables+=("$table")
    fi
done

if [ ${#missing_tables[@]} -eq 0 ]; then
    echo -e "\n${GREEN}✅ All tables created successfully!${NC}"
else
    echo -e "\n${RED}❌ Missing tables: ${missing_tables[*]}${NC}"
    echo -e "${YELLOW}Let's create them manually...${NC}"
    
    # Create missing tables manually
    for table in "${missing_tables[@]}"; do
        case $table in
            "user_roles")
                echo "Creating user_roles table..."
                docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "
                CREATE TABLE IF NOT EXISTS user_roles (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    user_id UUID NOT NULL,
                    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'moderator', 'admin')),
                    granted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                    is_active BOOLEAN DEFAULT TRUE,
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
                );
                CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
                CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);
                CREATE INDEX IF NOT EXISTS idx_user_roles_active ON user_roles(is_active);
                CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique_active ON user_roles(user_id, role) WHERE is_active = TRUE;
                "
                ;;
            "refresh_tokens")
                echo "Creating refresh_tokens table..."
                docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "
                CREATE TABLE IF NOT EXISTS refresh_tokens (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    user_id UUID NOT NULL,
                    token VARCHAR(255) UNIQUE NOT NULL,
                    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
                    is_revoked BOOLEAN DEFAULT FALSE,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
                );
                CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
                CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
                CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
                CREATE INDEX IF NOT EXISTS idx_refresh_tokens_active ON refresh_tokens(user_id, expires_at) WHERE is_revoked = FALSE;
                "
                ;;
            "email_verifications")
                echo "Creating email_verifications table..."
                docker exec auth_service_db psql -U maxon_auth -d maxon_auth_db -c "
                CREATE TABLE IF NOT EXISTS email_verifications (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    user_id UUID NOT NULL,
                    token VARCHAR(255) UNIQUE NOT NULL,
                    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
                    is_used BOOLEAN DEFAULT FALSE,
                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
                );
                CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);
                CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);
                CREATE INDEX IF NOT EXISTS idx_email_verifications_expires_at ON email_verifications(expires_at);
                CREATE INDEX IF NOT EXISTS idx_email_verifications_active ON email_verifications(user_id, expires_at) WHERE is_used = FALSE;
                "
                ;;
        esac
    done
fi

echo -e "\n${GREEN}🎉 Migration fix complete!${NC}"

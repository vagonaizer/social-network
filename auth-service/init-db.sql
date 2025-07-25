-- Create application user and database
CREATE USER maxon_auth WITH PASSWORD '123123';
CREATE DATABASE maxon_auth_db OWNER maxon_auth;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE maxon_auth_db TO maxon_auth;

-- Connect to the new database and grant schema privileges
\c maxon_auth_db;

-- Grant privileges on schema
GRANT ALL ON SCHEMA public TO maxon_auth;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO maxon_auth;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO maxon_auth;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO maxon_auth;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO maxon_auth;

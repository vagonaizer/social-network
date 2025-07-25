.PHONY: build run test clean proto docker migrate-up migrate-down dev-setup db-reset migrate-fix db-status swagger

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=auth-service
BINARY_UNIX=$(BINARY_NAME)_unix

# Protobuf parameters
PROTO_DIR=api/proto
PROTO_OUT_DIR=pkg/api
PROTO_FILES=$(shell find $(PROTO_DIR) -name "*.proto")

# Migration parameters
MIGRATE_CMD=migrate
MIGRATIONS_DIR=migrations
DATABASE_URL=postgres://maxon_auth:123123@localhost:5433/maxon_auth_db?sslmode=disable

# Load environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/auth-service

# Run the application locally
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/auth-service
	./$(BINARY_NAME)

# Run with hot reload (requires air)
dev:
	air -c .air.toml

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Clean build files
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -rf $(PROTO_OUT_DIR)
	rm -rf docs/

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/auth-service/main.go -o docs/
	@echo "Swagger documentation generated in docs/ directory"

# Install Swagger CLI
install-swagger:
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate protobuf files
proto:
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_OUT_DIR)
	@for proto_file in $(PROTO_FILES); do \
		echo "Processing $$proto_file"; \
		protoc --go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
			--go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
			$$proto_file; \
	done
	@echo "Protobuf generation completed"

# Clean and regenerate protobuf files
proto-clean:
	rm -rf $(PROTO_OUT_DIR)
	$(MAKE) proto

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/auth-service

# Docker commands
docker-build:
	docker build -t $(BINARY_NAME):latest .

docker-run:
	docker run -p 8080:8080 -p 9090:9090 $(BINARY_NAME):latest

# Start all services with Docker Compose
docker-up:
	docker-compose up -d

# Stop all services
docker-down:
	docker-compose down

# Show logs
docker-logs:
	docker logs auth_service_app -f | jq

# Restart services
docker-restart:
	docker-compose restart

# Full Docker setup (build + up)
docker-full:
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 15
	@echo "Running migrations..."
	$(MAKE) migrate-up
	@echo "Setup complete!"

# Development setup
dev-setup:
	$(GOGET) -u google.golang.org/protobuf/cmd/protoc-gen-go
	$(GOGET) -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
	$(GOGET) -u -d github.com/golang-migrate/migrate/v4/cmd/migrate
	$(MAKE) install-swagger
	cp .env.example .env
	@echo "Please edit .env file with your database credentials"

# Install golang-migrate
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Database setup with Docker
db-setup:
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 15
	$(MAKE) migrate-up

# Database migrations
migrate-up:
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down:
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

migrate-create:
	@read -p "Enter migration name: " name; \
	$(MIGRATE_CMD) create -ext sql -dir $(MIGRATIONS_DIR) $$name

migrate-force:
	@read -p "Enter version to force: " version; \
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version

migrate-version:
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

# Manual migration execution (alternative method)
migrate-manual:
	@echo "Running migrations manually..."
	@for file in $(shell ls migrations/*up.sql | sort); do \
		echo "Executing $$file"; \
		docker exec -i auth_service_db psql -U maxon_auth -d maxon_auth_db < $$file; \
	done
	@echo "Manual migrations completed"

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Generate mocks (if using mockgen)
mocks:
	mockgen -source=repository/user_repository.go -destination=mocks/user_repository_mock.go
	mockgen -source=repository/user_auth_repository.go -destination=mocks/user_auth_repository_mock.go
	mockgen -source=service/auth_service.go -destination=mocks/auth_service_mock.go

# Validate protobuf files
proto-validate:
	@echo "Validating protobuf files..."
	@for proto_file in $(PROTO_FILES); do \
		echo "Validating $$proto_file"; \
		protoc --descriptor_set_out=/dev/null $$proto_file; \
	done
	@echo "Protobuf validation completed"

# Show protobuf file structure
proto-tree:
	@echo "Protobuf file structure:"
	@find $(PROTO_DIR) -name "*.proto" -type f | sort

# Reset database completely
db-reset:
	docker-compose down -v
	docker volume prune -f
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 10
	@echo "Checking PostgreSQL connection..."
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		if docker exec auth_service_db pg_isready -U postgres >/dev/null 2>&1; then \
			echo "PostgreSQL is ready!"; \
			break; \
		fi; \
		echo "Attempt $$i: PostgreSQL not ready, waiting..."; \
		sleep 3; \
	done
	@echo "Running migrations..."
	$(MAKE) migrate-up || $(MAKE) migrate-manual

# Fix dirty migration state
migrate-fix:
	@echo "Current migration version:"
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version
	@read -p "Enter version to force (0 to reset): " version; \
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $$version

# Check database state
db-status:
	@echo "=== Migration Version ==="
	$(MIGRATE_CMD) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version || echo "Migration tool not working, checking manually..."
	@echo "=== Tables ==="
	docker exec -it auth_service_db psql -U maxon_auth -d maxon_auth_db -c "\dt"
	@echo "=== Migration History ==="
	docker exec -it auth_service_db psql -U maxon_auth -d maxon_auth_db -c "SELECT * FROM schema_migrations;" || echo "No migration table found"

# Wait for database to be ready
db-wait:
	@echo "Waiting for database to be ready..."
	@until docker exec auth_service_db pg_isready -U postgres; do \
		echo "Waiting for PostgreSQL..."; \
		sleep 2; \
	done
	@echo "Database is ready!"

# Build with Swagger generation
build-with-docs: swagger build

# Full development build
dev-build: deps swagger build

# Help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Build and run the application locally"
	@echo "  dev             - Run with hot reload (requires air)"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  clean           - Clean build files"
	@echo "  deps            - Download dependencies"
	@echo "  swagger         - Generate Swagger documentation"
	@echo "  install-swagger - Install Swagger CLI tool"
	@echo "  proto           - Generate protobuf files"
	@echo "  proto-clean     - Clean and regenerate protobuf files"
	@echo "  proto-validate  - Validate protobuf files"
	@echo "  proto-tree      - Show protobuf file structure"
	@echo "  build-linux     - Build for Linux"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-up       - Start all services with docker-compose"
	@echo "  docker-down     - Stop all services"
	@echo "  docker-logs     - Show docker-compose logs"
	@echo "  docker-restart  - Restart services"
	@echo "  docker-full     - Full Docker setup (build + up + migrate)"
	@echo "  dev-setup       - Setup development environment"
	@echo "  install-migrate - Install golang-migrate tool"
	@echo "  db-setup        - Setup database with Docker"
	@echo "  migrate-up      - Run database migrations up"
	@echo "  migrate-down    - Run database migrations down"
	@echo "  migrate-create  - Create new migration"
	@echo "  migrate-force   - Force migration version"
	@echo "  migrate-version - Show current migration version"
	@echo "  migrate-manual  - Run migrations manually via docker exec"
	@echo "  lint            - Lint code"
	@echo "  fmt             - Format code"
	@echo "  mocks           - Generate mocks"
	@echo "  db-reset        - Reset database completely"
	@echo "  migrate-fix     - Fix dirty migration state"
	@echo "  db-status       - Check database state"
	@echo "  db-wait         - Wait for database to be ready"
	@echo "  build-with-docs - Build application with Swagger docs"
	@echo "  dev-build       - Full development build (deps + swagger + build)"

.PHONY: run build test seed migrate rollback automigrate swagger docker-up docker-down lint help

# Default target
help:
	@echo "Available commands:"
	@echo "  make run         - Run the application with Air (hot reload) if available, otherwise go run"
	@echo "  make build       - Build the application binary"
	@echo "  make test        - Run unit tests"
	@echo "  make seed        - Run database seeders"
	@echo "  make migrate     - Run versioned migrations (up)"
	@echo "  make rollback    - Rollback the last migration"
	@echo "  make automigrate - Run GORM auto-migration (dev mode)"
	@echo "  make swagger     - Generate Swagger documentation"
	@echo "  make docker-up   - Start docker containers (postgres, pgadmin)"
	@echo "  make docker-down - Stop docker containers"
	@echo "  make lint        - Run fmt and vet on the code"

run:
	@if command -v air > /dev/null; then \
		air; \
	else \
		go run cmd/api/main.go; \
	fi

build:
	@echo "Building application..."
	@go build -o bin/api cmd/api/main.go

test:
	@echo "Running tests..."
	@go test -v ./...

seed:
	@echo "Seeding database..."
	@go run cmd/seed/main.go

migrate:
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

rollback:
	@echo "Rolling back migration..."
	@go run cmd/migrate/main.go rollback

automigrate:
	@echo "Running GORM auto-migration..."
	@go run cmd/migrate/main.go automigrate

swagger:
	@echo "Generating swagger docs..."
	@if command -v swag > /dev/null; then \
		swag init -g cmd/api/main.go -o docs; \
	else \
		go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go -o docs; \
	fi

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

lint:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Vetting code..."
	@go vet ./...

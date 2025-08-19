.PHONY: build run test clean docker-build docker-run docker-stop help deps fmt lint docs prod-build dev-tools migrate dev-setup dev quick-dev prod-prep

# Default target
help:
	@echo "Asset Tagging Backend - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  build        - Build the Go application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  fmt          - Format Go code"
	@echo "  lint         - Lint Go code"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker containers"
	@echo "  docker-rebuild - Rebuild and run with Docker Compose"
	@echo ""
	@echo "Database:"
	@echo "  migrate      - Run database migrations"
	@echo ""
	@echo "Production:"
	@echo "  prod-build   - Create production build"
	@echo "  docs         - Generate documentation"
	@echo ""
	@echo "Tools:"
	@echo "  dev-tools    - Install development tools"

# Build the application
build:
	@echo "Building asset-tagging-backend..."
	go build -o asset-tagging-backend .

# Run the application
run:
	@echo "Running asset-tagging-backend..."
	go run .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f asset-tagging-backend
	rm -f *.pdf
	rm -f *.xlsx
	rm -f barcode_*.png
	rm -rf build/
	rm -rf dist/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t asset-tagging-backend .

# Run with Docker Compose
docker-run:
	@echo "Starting Docker containers..."
	docker-compose up -d

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# Run with Docker Compose and rebuild
docker-rebuild:
	@echo "Rebuilding and starting Docker containers..."
	docker-compose up -d --build

# View logs
logs:
	docker-compose logs -f app

# Database logs
db-logs:
	docker-compose logs -f db

# Access database
db-access:
	docker-compose exec db mysql -u root -ppassword asset_management

# Format code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting Go code..."
	golangci-lint run

# Generate documentation
docs:
	@echo "Starting documentation server on http://localhost:6060"
	godoc -http=:6060

# Create production build
prod-build:
	@echo "Creating production build..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o asset-tagging-backend .

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/godoc@latest

# Run database migrations
migrate:
	@echo "Running database migrations..."
	@if [ -f migrations.sql ]; then \
		echo "Found migrations.sql - please run manually with your database credentials:"; \
		echo "mysql -h HOST -u USERNAME -pPASSWORD DATABASE < migrations.sql"; \
	else \
		echo "No migrations.sql file found"; \
	fi

# Development setup
dev-setup: deps dev-tools
	@echo "Development environment setup complete!"

# Quick development cycle
dev: fmt lint test build
	@echo "Development cycle complete!"

# Production deployment preparation
prod-prep: clean deps fmt lint test prod-build
	@echo "Production preparation complete!" 
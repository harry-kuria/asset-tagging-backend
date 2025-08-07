.PHONY: build run test clean docker-build docker-run docker-stop help

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the Go application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker containers"
	@echo "  deps         - Install dependencies"

# Build the application
build:
	go build -o asset-tagging-backend .

# Run the application
run:
	go run .

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f asset-tagging-backend
	rm -f *.pdf
	rm -f *.xlsx
	rm -f barcode_*.png

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	docker build -t asset-tagging-backend .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker containers
docker-stop:
	docker-compose down

# Run with Docker Compose and rebuild
docker-rebuild:
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
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Generate documentation
docs:
	godoc -http=:6060

# Create production build
prod-build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o asset-tagging-backend .

# Install development tools
dev-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/godoc@latest 
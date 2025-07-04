# File Watch Server Makefile

.PHONY: build clean test deps install-deps run-uploader run-watcher init-config help

# Variables
APP_NAME = file-watch-server
GO_VERSION = 1.21
BUILD_DIR = build
BINARY_NAME = fws

# Build flags
LDFLAGS = -ldflags "-s -w"

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/fws-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/fws-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/fws-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/fws-darwin-arm64 .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Install development dependencies
install-deps:
	@echo "Installing development dependencies..."
	go install -a std
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint:
	@echo "Linting code..."
	golangci-lint run

# Format the code
format:
	@echo "Formatting code..."
	go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f fws
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Initialize configuration
init-config:
	@echo "Initializing configuration..."
	./fws init

# Run in uploader mode
run-uploader:
	@echo "Running in uploader mode..."
	./fws --mode uploader --config config.json --verbose

# Run in watcher mode
run-watcher:
	@echo "Running in watcher mode..."
	./fws --mode watcher --config config.json --verbose

# Run in watcher daemon mode
run-watcher-daemon:
	@echo "Running in watcher daemon mode..."
	./fws --mode watcher --config config.json --daemon

# Show status
status:
	@echo "Checking container status..."
	./fws status --config config.json

# Show logs
logs:
	@echo "Showing container logs..."
	./fws logs --config config.json

# Install the binary system-wide
install: build
	@echo "Installing fws to /usr/local/bin..."
	sudo cp fws /usr/local/bin/
	sudo chmod +x /usr/local/bin/fws

# Uninstall the binary
uninstall:
	@echo "Uninstalling fws..."
	sudo rm -f /usr/local/bin/fws

# Create systemd service
create-service:
	@echo "Creating systemd service..."
	@echo '[Unit]' | sudo tee /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'Description=File Watch Server' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'After=network.target docker.service' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'Requires=docker.service' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo '' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo '[Service]' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'Type=simple' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'User=deploy' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'WorkingDirectory=/opt/$(APP_NAME)' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'ExecStart=/usr/local/bin/fws --config /opt/$(APP_NAME)/config.json --daemon' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'Restart=always' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'RestartSec=10' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo '' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo '[Install]' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	@echo 'WantedBy=multi-user.target' | sudo tee -a /etc/systemd/system/$(APP_NAME).service > /dev/null
	sudo systemctl daemon-reload
	@echo "Service created. Enable with: sudo systemctl enable $(APP_NAME)"

# Remove systemd service
remove-service:
	@echo "Removing systemd service..."
	sudo systemctl stop $(APP_NAME) || true
	sudo systemctl disable $(APP_NAME) || true
	sudo rm -f /etc/systemd/system/$(APP_NAME).service
	sudo systemctl daemon-reload

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

# Docker run uploader
docker-run-uploader:
	@echo "Running Docker container in uploader mode..."
	docker run --rm -v $(PWD)/config.json:/app/config.json -v /var/run/docker.sock:/var/run/docker.sock $(APP_NAME):latest fws --mode uploader --config /app/config.json

# Docker run watcher
docker-run-watcher:
	@echo "Running Docker container in watcher mode..."
	docker run --rm -v $(PWD)/config.json:/app/config.json -v /var/run/docker.sock:/var/run/docker.sock -v /opt/docker-uploads:/opt/docker-uploads $(APP_NAME):latest fws --mode watcher --config /app/config.json

# Show help
help:
	@echo "Available targets:"
	@echo "  build                 - Build the application"
	@echo "  build-all             - Build for multiple platforms"
	@echo "  deps                  - Install dependencies"
	@echo "  install-deps          - Install development dependencies"
	@echo "  test                  - Run tests"
	@echo "  test-coverage         - Run tests with coverage"
	@echo "  lint                  - Lint the code"
	@echo "  format                - Format the code"
	@echo "  clean                 - Clean build artifacts"
	@echo "  init-config           - Initialize configuration"
	@echo "  run-uploader          - Run in uploader mode"
	@echo "  run-watcher           - Run in watcher mode"
	@echo "  run-watcher-daemon    - Run in watcher daemon mode"
	@echo "  status                - Check container status"
	@echo "  logs                  - Show container logs"
	@echo "  install               - Install binary system-wide"
	@echo "  uninstall             - Uninstall binary"
	@echo "  create-service        - Create systemd service"
	@echo "  remove-service        - Remove systemd service"
	@echo "  docker-build          - Build Docker image"
	@echo "  docker-run-uploader   - Run Docker container in uploader mode"
	@echo "  docker-run-watcher    - Run Docker container in watcher mode"
	@echo "  help                  - Show this help message" 
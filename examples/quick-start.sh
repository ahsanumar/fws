#!/bin/bash

# File Watch Server Quick Start Script
# This script demonstrates how to use the file-watch-server application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if fws binary exists
check_binary() {
    if [ ! -f "./fws" ]; then
        log_error "fws binary not found. Please build it first:"
        echo "  make build"
        exit 1
    fi
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi
}

# Create necessary directories
setup_directories() {
    log_info "Setting up directories..."
    mkdir -p ./tarballs
    mkdir -p ./docker-uploads
    log_success "Directories created"
}

# Build the sample application
build_sample_app() {
    log_info "Building sample application..."
    
    cd examples/sample-app
    docker build -t sample-web-app:latest .
    cd ../..
    
    log_success "Sample application built"
}

# Demonstrate uploader mode
demo_uploader() {
    log_info "Demonstrating uploader mode..."
    
    # Create a temporary config for demo
    cat > demo-uploader-config.json << EOF
{
  "mode": "uploader",
  "log_level": "info",
  "uploader": {
    "docker_build_path": "./examples/sample-app",
    "image_name": "sample-web-app",
    "image_tag": "demo",
    "tarball_path": "./tarballs",
    "remote_host": "localhost",
    "remote_port": 22,
    "remote_user": "$(whoami)",
    "remote_key_path": "~/.ssh/id_rsa",
    "remote_upload_path": "./docker-uploads",
    "pre_build_commands": [
      "echo 'Starting build process for demo...'",
      "date"
    ],
    "post_build_commands": [
      "echo 'Build process completed'",
      "ls -la ./tarballs/"
    ]
  }
}
EOF

    # For demo purposes, we'll just create the tarball locally
    log_info "Creating tarball locally (simulating upload)..."
    docker save sample-web-app:demo -o ./docker-uploads/sample-web-app_demo_$(date +%Y%m%d-%H%M%S).tar
    
    log_success "Tarball created in ./docker-uploads/"
    ls -la ./docker-uploads/
}

# Demonstrate watcher mode
demo_watcher() {
    log_info "Demonstrating watcher mode..."
    
    # Create a temporary config for demo
    cat > demo-watcher-config.json << EOF
{
  "mode": "watcher",
  "log_level": "info",
  "watcher": {
    "watch_directory": "./docker-uploads",
    "container_name": "sample-web-app-demo",
    "container_ports": ["8080:80"],
    "container_env": [
      "NGINX_HOST=localhost",
      "ENVIRONMENT=demo"
    ],
    "container_volumes": [],
    "pre_load_commands": [
      "echo 'Preparing to load new image...'",
      "docker ps --filter name=sample-web-app-demo || true"
    ],
    "post_load_commands": [
      "echo 'New container deployed successfully'",
      "sleep 3",
      "curl -f http://localhost:8080/health || echo 'Health check will be available in a moment'",
      "echo 'Container is running. Access it at http://localhost:8080'"
    ],
    "restart_policy": "unless-stopped"
  }
}
EOF

    log_info "Starting watcher in background..."
    ./fws --config demo-watcher-config.json &
    WATCHER_PID=$!
    
    log_info "Watcher started with PID: $WATCHER_PID"
    log_info "Waiting for container to be ready..."
    
    # Wait for container to be ready
    sleep 10
    
    # Check if container is running
    if docker ps --filter name=sample-web-app-demo --format "{{.Names}}" | grep -q sample-web-app-demo; then
        log_success "Container is running!"
        log_info "Access the demo application at: http://localhost:8080"
        log_info "Health check endpoint: http://localhost:8080/health"
        log_info "API status endpoint: http://localhost:8080/api/status"
        
        # Test the endpoints
        echo
        log_info "Testing health endpoint..."
        curl -s http://localhost:8080/health || log_warning "Health check failed"
        
        echo
        log_info "Testing API endpoint..."
        curl -s http://localhost:8080/api/status | jq . || log_warning "API check failed"
        
        echo
        log_info "Container status:"
        docker ps --filter name=sample-web-app-demo --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    else
        log_error "Container failed to start"
    fi
    
    log_info "Stopping watcher (PID: $WATCHER_PID)..."
    kill $WATCHER_PID 2>/dev/null || true
    wait $WATCHER_PID 2>/dev/null || true
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    
    # Stop and remove demo container
    docker stop sample-web-app-demo 2>/dev/null || true
    docker rm sample-web-app-demo 2>/dev/null || true
    
    # Remove demo files
    rm -f demo-uploader-config.json demo-watcher-config.json
    rm -rf ./docker-uploads/*.tar
    
    log_success "Cleanup completed"
}

# Main function
main() {
    echo "======================================"
    echo "File Watch Server Quick Start Demo"
    echo "======================================"
    echo
    
    # Check prerequisites
    check_binary
    check_docker
    
    # Setup
    setup_directories
    build_sample_app
    
    # Demo uploader mode
    echo
    echo "======================================"
    echo "UPLOADER MODE DEMO"
    echo "======================================"
    demo_uploader
    
    # Demo watcher mode
    echo
    echo "======================================"
    echo "WATCHER MODE DEMO"
    echo "======================================"
    demo_watcher
    
    # Cleanup
    echo
    echo "======================================"
    echo "CLEANUP"
    echo "======================================"
    cleanup
    
    echo
    log_success "Demo completed!"
    echo
    echo "Next steps:"
    echo "1. Edit the configuration files to match your environment"
    echo "2. Set up SSH keys for remote deployment"
    echo "3. Run the application in your production environment"
    echo
    echo "For more information, see the README.md file"
}

# Handle script interruption
trap cleanup EXIT

# Parse command line arguments
case "${1:-}" in
    "build")
        check_binary
        check_docker
        build_sample_app
        ;;
    "uploader")
        check_binary
        check_docker
        setup_directories
        build_sample_app
        demo_uploader
        ;;
    "watcher")
        check_binary
        check_docker
        setup_directories
        demo_watcher
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        main
        ;;
esac 
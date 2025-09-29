#!/bin/bash

# =============================================================================
# Astra-xmod-shim Deployment Script
# =============================================================================
# Description: Comprehensive deployment script for astra-xmod-shim middleware
# Author: Codegen AI Assistant
# Version: 1.0.0
# =============================================================================

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_NAME="astra-xmod-shim"
BINARY_NAME="astra-xmod-shim"
DEFAULT_PORT="7777"
DEFAULT_CONFIG_PATH="./conf/base/conf.yaml"

# Deployment options
DEPLOYMENT_MODE="${DEPLOYMENT_MODE:-local}"  # local, docker, kubernetes
GO_VERSION_REQUIRED="1.20"
HEALTH_CHECK_TIMEOUT=30

# Logging function
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[WARN] $1${NC}"
}

info() {
    echo -e "${BLUE}[INFO] $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        error "Go is not installed. Please install Go ${GO_VERSION_REQUIRED}+ first."
        exit 1
    fi
    
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local required_version=$(echo $GO_VERSION_REQUIRED | sed 's/\.//')
    local current_version=$(echo $go_version | sed 's/\.//')
    
    if [ "$current_version" -lt "$required_version" ]; then
        error "Go version $go_version is too old. Required: ${GO_VERSION_REQUIRED}+"
        exit 1
    fi
    
    log "Go version $go_version is compatible"
    
    # Check Docker if needed
    if [ "$DEPLOYMENT_MODE" = "docker" ] && ! command -v docker &> /dev/null; then
        error "Docker is not installed but required for docker deployment mode"
        exit 1
    fi
    
    # Check kubectl if needed
    if [ "$DEPLOYMENT_MODE" = "kubernetes" ] && ! command -v kubectl &> /dev/null; then
        error "kubectl is not installed but required for kubernetes deployment mode"
        exit 1
    fi
}

# Clone or update repository
setup_repository() {
    log "Setting up repository..."
    
    if [ ! -d "$PROJECT_NAME" ]; then
        log "Cloning repository..."
        git clone https://github.com/Zeeeepa/astra-xmod-shim.git
        cd "$PROJECT_NAME"
    else
        log "Repository exists, updating..."
        cd "$PROJECT_NAME"
        git pull origin main
    fi
}

# Build the application
build_application() {
    log "Building application..."
    
    # Clean previous builds
    make clean || true
    
    # Build binary
    log "Compiling Go binary..."
    make build
    
    if [ ! -f "./bin/$BINARY_NAME" ]; then
        error "Build failed - binary not found"
        exit 1
    fi
    
    log "Build completed successfully"
}

# Setup configuration
setup_configuration() {
    log "Setting up configuration..."
    
    # Create logs directory
    mkdir -p logs
    
    # Copy default config if custom doesn't exist
    if [ ! -f "conf/base/conf.yaml" ]; then
        warn "Default configuration not found, creating basic config..."
        mkdir -p conf/base
        cat > conf/base/conf.yaml << EOF
# Basic configuration for astra-xmod-shim
server:
  port: ":${DEFAULT_PORT}"

log:
  level: "info"
  path: "./logs/app.log"
  max-size: 100
  max-age: 7
  compress: false
  show-line: true
  enable-console: true

current-shimlet: "k8s"

shimlets:
  k8s:
    config-path: "/opt/modserv-shim/conf/shimlets/k8s-shimlet.yaml"

model-manage:
  model-root: "/mnt/maasmodels/"

tracer:
  interval: 30
EOF
    fi
    
    log "Configuration setup completed"
}

# Deploy locally
deploy_local() {
    log "Deploying locally..."
    
    # Kill existing process if running
    if pgrep -f "$BINARY_NAME" > /dev/null; then
        warn "Stopping existing $BINARY_NAME process..."
        pkill -f "$BINARY_NAME" || true
        sleep 2
    fi
    
    # Start the service
    log "Starting $BINARY_NAME service..."
    nohup ./bin/$BINARY_NAME --config="$DEFAULT_CONFIG_PATH" > logs/service.log 2>&1 &
    
    local pid=$!
    echo $pid > "$BINARY_NAME.pid"
    
    log "Service started with PID: $pid"
    
    # Health check
    health_check_local
}

# Deploy with Docker
deploy_docker() {
    log "Deploying with Docker..."
    
    # Build Docker image
    log "Building Docker image..."
    make docker
    
    # Stop existing container
    if docker ps -q -f name="$PROJECT_NAME" | grep -q .; then
        warn "Stopping existing container..."
        docker stop "$PROJECT_NAME" || true
        docker rm "$PROJECT_NAME" || true
    fi
    
    # Run container
    log "Starting Docker container..."
    docker run -d \
        --name "$PROJECT_NAME" \
        --restart unless-stopped \
        -p "${DEFAULT_PORT}:${DEFAULT_PORT}" \
        -v "$(pwd)/conf:/opt/modserv-shim/conf" \
        -v "$(pwd)/logs:/opt/modserv-shim/logs" \
        "ghcr.io/iflytek/$PROJECT_NAME:latest"
    
    log "Docker container started"
    
    # Health check
    health_check_docker
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log "Deploying to Kubernetes..."
    
    # Check if Helm chart exists
    if [ ! -d "deploy/helm/$PROJECT_NAME" ]; then
        error "Helm chart not found at deploy/helm/$PROJECT_NAME"
        exit 1
    fi
    
    # Package Helm chart
    log "Packaging Helm chart..."
    make helm-package
    
    # Deploy with Helm
    log "Deploying with Helm..."
    helm upgrade --install "$PROJECT_NAME" \
        "deploy/helm/$PROJECT_NAME" \
        --namespace "$PROJECT_NAME" \
        --create-namespace \
        --wait \
        --timeout=300s
    
    log "Kubernetes deployment completed"
    
    # Health check
    health_check_kubernetes
}

# Health check for local deployment
health_check_local() {
    log "Performing health check..."
    
    local count=0
    while [ $count -lt $HEALTH_CHECK_TIMEOUT ]; do
        if curl -s "http://localhost:${DEFAULT_PORT}/api/v1/plugins" > /dev/null 2>&1; then
            log "Health check passed - service is responding"
            return 0
        fi
        
        sleep 1
        count=$((count + 1))
    done
    
    error "Health check failed - service not responding after ${HEALTH_CHECK_TIMEOUT}s"
    return 1
}

# Health check for Docker deployment
health_check_docker() {
    log "Performing Docker health check..."
    
    local count=0
    while [ $count -lt $HEALTH_CHECK_TIMEOUT ]; do
        if docker exec "$PROJECT_NAME" curl -s "http://localhost:${DEFAULT_PORT}/api/v1/plugins" > /dev/null 2>&1; then
            log "Docker health check passed"
            return 0
        fi
        
        sleep 1
        count=$((count + 1))
    done
    
    error "Docker health check failed"
    return 1
}

# Health check for Kubernetes deployment
health_check_kubernetes() {
    log "Performing Kubernetes health check..."
    
    if kubectl wait --for=condition=ready pod -l app="$PROJECT_NAME" -n "$PROJECT_NAME" --timeout=300s; then
        log "Kubernetes health check passed"
        return 0
    else
        error "Kubernetes health check failed"
        return 1
    fi
}

# Show service status
show_status() {
    log "Service Status:"
    
    case $DEPLOYMENT_MODE in
        "local")
            if pgrep -f "$BINARY_NAME" > /dev/null; then
                info "✅ Service is running (PID: $(pgrep -f "$BINARY_NAME"))"
                info "📊 API Status: $(curl -s http://localhost:${DEFAULT_PORT}/api/v1/plugins | jq -r 'length // "N/A"') plugins loaded"
            else
                warn "❌ Service is not running"
            fi
            ;;
        "docker")
            if docker ps -q -f name="$PROJECT_NAME" | grep -q .; then
                info "✅ Docker container is running"
                info "📊 Container Status: $(docker inspect --format='{{.State.Status}}' "$PROJECT_NAME")"
            else
                warn "❌ Docker container is not running"
            fi
            ;;
        "kubernetes")
            local pods=$(kubectl get pods -l app="$PROJECT_NAME" -n "$PROJECT_NAME" --no-headers 2>/dev/null | wc -l)
            if [ "$pods" -gt 0 ]; then
                info "✅ Kubernetes deployment is running ($pods pods)"
                kubectl get pods -l app="$PROJECT_NAME" -n "$PROJECT_NAME"
            else
                warn "❌ Kubernetes deployment not found"
            fi
            ;;
    esac
}

# Stop service
stop_service() {
    log "Stopping service..."
    
    case $DEPLOYMENT_MODE in
        "local")
            if [ -f "$BINARY_NAME.pid" ]; then
                local pid=$(cat "$BINARY_NAME.pid")
                if kill "$pid" 2>/dev/null; then
                    log "Service stopped (PID: $pid)"
                    rm -f "$BINARY_NAME.pid"
                else
                    warn "Process $pid not found, cleaning up PID file"
                    rm -f "$BINARY_NAME.pid"
                fi
            else
                pkill -f "$BINARY_NAME" || warn "No running process found"
            fi
            ;;
        "docker")
            docker stop "$PROJECT_NAME" || warn "Container not running"
            docker rm "$PROJECT_NAME" || warn "Container not found"
            ;;
        "kubernetes")
            helm uninstall "$PROJECT_NAME" -n "$PROJECT_NAME" || warn "Helm release not found"
            kubectl delete namespace "$PROJECT_NAME" || warn "Namespace not found"
            ;;
    esac
}

# Show logs
show_logs() {
    log "Showing logs..."
    
    case $DEPLOYMENT_MODE in
        "local")
            if [ -f "logs/service.log" ]; then
                tail -f logs/service.log
            else
                error "Log file not found"
            fi
            ;;
        "docker")
            docker logs -f "$PROJECT_NAME"
            ;;
        "kubernetes")
            kubectl logs -f -l app="$PROJECT_NAME" -n "$PROJECT_NAME"
            ;;
    esac
}

# Main deployment function
main() {
    log "Starting $PROJECT_NAME deployment (mode: $DEPLOYMENT_MODE)"
    
    check_prerequisites
    setup_repository
    build_application
    setup_configuration
    
    case $DEPLOYMENT_MODE in
        "local")
            deploy_local
            ;;
        "docker")
            deploy_docker
            ;;
        "kubernetes")
            deploy_kubernetes
            ;;
        *)
            error "Unknown deployment mode: $DEPLOYMENT_MODE"
            exit 1
            ;;
    esac
    
    show_status
    
    log "Deployment completed successfully!"
    log "Service URL: http://localhost:${DEFAULT_PORT}"
    log "API Endpoints:"
    log "  - GET  /api/v1/plugins - List available plugins"
    log "  - POST /api/v1/modserv/deploy - Deploy model service"
    log "  - GET  /api/v1/modserv/{id} - Get service status"
}

# Command line interface
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "status")
        show_status
        ;;
    "stop")
        stop_service
        ;;
    "logs")
        show_logs
        ;;
    "restart")
        stop_service
        sleep 2
        main
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo "Commands:"
        echo "  deploy   - Deploy the service (default)"
        echo "  status   - Show service status"
        echo "  stop     - Stop the service"
        echo "  logs     - Show service logs"
        echo "  restart  - Restart the service"
        echo "  help     - Show this help"
        echo ""
        echo "Environment Variables:"
        echo "  DEPLOYMENT_MODE - local|docker|kubernetes (default: local)"
        echo "  DEFAULT_PORT    - Service port (default: 7777)"
        ;;
    *)
        error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac


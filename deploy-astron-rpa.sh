#!/bin/bash

# =============================================================================
# Astron RPA Deployment Script
# =============================================================================
# Description: Comprehensive deployment script for Astron RPA platform
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
PROJECT_NAME="astron-rpa"
DEFAULT_PORT="8080"
DOCKER_COMPOSE_PATH="docker/docker-compose.yml"

# Deployment options
DEPLOYMENT_MODE="${DEPLOYMENT_MODE:-docker}"  # docker, source, tauri
PYTHON_VERSION_REQUIRED="3.13"
NODE_VERSION_REQUIRED="22"
JAVA_VERSION_REQUIRED="8"
RUST_VERSION_REQUIRED="1.90"
HEALTH_CHECK_TIMEOUT=60

# Database configuration
DATABASE_NAME="${DATABASE_NAME:-rpa_opensource}"
DATABASE_USERNAME="${DATABASE_USERNAME:-root}"
DATABASE_PASSWORD="${DATABASE_PASSWORD:-123456}"
DATABASE_PORT="${DATABASE_PORT:-3306}"
REDIS_PASSWORD="${REDIS_PASSWORD:-123456}"

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
    
    case $DEPLOYMENT_MODE in
        "docker")
            # Check Docker and Docker Compose
            if ! command -v docker &> /dev/null; then
                error "Docker is not installed. Please install Docker first."
                exit 1
            fi
            
            if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
                error "Docker Compose is not installed. Please install Docker Compose first."
                exit 1
            fi
            
            log "Docker environment is ready"
            ;;
            
        "source"|"tauri")
            # Check Python
            if ! command -v python3 &> /dev/null; then
                error "Python 3 is not installed. Please install Python ${PYTHON_VERSION_REQUIRED}+ first."
                exit 1
            fi
            
            local python_version=$(python3 --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
            local required_version=$(echo $PYTHON_VERSION_REQUIRED | sed 's/\.//')
            local current_version=$(echo $python_version | sed 's/\.//')
            
            if [ "$current_version" -lt "$required_version" ]; then
                warn "Python version $python_version might be too old. Recommended: ${PYTHON_VERSION_REQUIRED}+"
            fi
            
            # Check Node.js
            if ! command -v node &> /dev/null; then
                error "Node.js is not installed. Please install Node.js ${NODE_VERSION_REQUIRED}+ first."
                exit 1
            fi
            
            local node_version=$(node -v | sed 's/v//' | cut -d'.' -f1)
            if [ "$node_version" -lt "$NODE_VERSION_REQUIRED" ]; then
                error "Node.js version $node_version is too old. Required: ${NODE_VERSION_REQUIRED}+"
                exit 1
            fi
            
            # Check pnpm
            if ! command -v pnpm &> /dev/null; then
                error "pnpm is not installed. Please install pnpm first: npm install -g pnpm"
                exit 1
            fi
            
            # Check Java
            if ! command -v java &> /dev/null; then
                error "Java is not installed. Please install Java ${JAVA_VERSION_REQUIRED}+ first."
                exit 1
            fi
            
            # Check UV (Python package manager)
            if ! command -v uv &> /dev/null; then
                warn "UV is not installed. Installing UV..."
                curl -LsSf https://astral.sh/uv/install.sh | sh
                export PATH="$HOME/.cargo/bin:$PATH"
            fi
            
            if [ "$DEPLOYMENT_MODE" = "tauri" ]; then
                # Check Rust for Tauri
                if ! command -v rustc &> /dev/null; then
                    error "Rust is not installed. Please install Rust ${RUST_VERSION_REQUIRED}+ first."
                    exit 1
                fi
                
                local rust_version=$(rustc --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
                log "Rust version $rust_version detected"
            fi
            
            log "Source build environment is ready"
            ;;
    esac
}

# Clone or update repository
setup_repository() {
    log "Setting up repository..."
    
    if [ ! -d "$PROJECT_NAME" ]; then
        log "Cloning repository..."
        git clone https://github.com/Zeeeepa/astron-rpa.git
        cd "$PROJECT_NAME"
    else
        log "Repository exists, updating..."
        cd "$PROJECT_NAME"
        git pull origin main
    fi
}

# Setup environment configuration
setup_environment() {
    log "Setting up environment configuration..."
    
    case $DEPLOYMENT_MODE in
        "docker")
            # Setup Docker environment
            if [ ! -f "docker/.env" ]; then
                log "Creating Docker environment file..."
                cp docker/.env.example docker/.env
                
                # Update environment variables
                sed -i "s/DATABASE_NAME=.*/DATABASE_NAME=${DATABASE_NAME}/" docker/.env
                sed -i "s/DATABASE_USERNAME=.*/DATABASE_USERNAME=${DATABASE_USERNAME}/" docker/.env
                sed -i "s/DATABASE_PASSWORD=.*/DATABASE_PASSWORD=${DATABASE_PASSWORD}/" docker/.env
                sed -i "s/DATABASE_PORT=.*/DATABASE_PORT=${DATABASE_PORT}/" docker/.env
                sed -i "s/REDIS_PASSWORD=.*/REDIS_PASSWORD=${REDIS_PASSWORD}/" docker/.env
            fi
            ;;
            
        "source"|"tauri")
            # Setup source environment
            log "Setting up source build environment..."
            
            # Create application configuration
            mkdir -p config
            cat > config/application.yml << EOF
server:
  port: ${DEFAULT_PORT}

spring:
  datasource:
    url: jdbc:mysql://localhost:${DATABASE_PORT}/${DATABASE_NAME}
    username: ${DATABASE_USERNAME}
    password: ${DATABASE_PASSWORD}
    driver-class-name: com.mysql.cj.jdbc.Driver
  
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: false
    
  redis:
    host: localhost
    port: 6379
    password: ${REDIS_PASSWORD}
    
logging:
  level:
    com.iflytek: INFO
    org.springframework: WARN
EOF
            
            # Setup Python environment
            log "Setting up Python virtual environment..."
            if [ ! -d "venv" ]; then
                python3 -m venv venv
            fi
            source venv/bin/activate
            
            # Install Python dependencies
            if [ -f "engine/requirements.txt" ]; then
                pip install -r engine/requirements.txt
            fi
            ;;
    esac
    
    log "Environment configuration completed"
}

# Build application components
build_application() {
    log "Building application components..."
    
    case $DEPLOYMENT_MODE in
        "source"|"tauri")
            # Build backend services
            log "Building Java backend services..."
            if [ -d "backend/robot-service" ]; then
                cd backend/robot-service
                ./mvnw clean package -DskipTests
                cd ../..
            fi
            
            if [ -d "backend/resource-service" ]; then
                cd backend/resource-service
                ./mvnw clean package -DskipTests
                cd ../..
            fi
            
            # Install frontend dependencies
            log "Installing frontend dependencies..."
            cd frontend
            pnpm install
            
            if [ "$DEPLOYMENT_MODE" = "tauri" ]; then
                # Build Tauri desktop application
                log "Building Tauri desktop application..."
                pnpm build:tauri
            else
                # Build web application
                log "Building web application..."
                pnpm build:web
            fi
            
            cd ..
            
            # Setup Python RPA engine
            log "Setting up Python RPA engine..."
            source venv/bin/activate
            cd engine
            
            # Install engine dependencies
            if [ -f "requirements.txt" ]; then
                pip install -r requirements.txt
            fi
            
            # Install component packages
            for component_dir in components/*/; do
                if [ -d "$component_dir" ] && [ -f "${component_dir}requirements.txt" ]; then
                    log "Installing dependencies for $(basename "$component_dir")"
                    pip install -r "${component_dir}requirements.txt"
                fi
            done
            
            cd ..
            
            log "Build completed successfully"
            ;;
            
        "docker")
            log "Docker build will be handled by docker-compose"
            ;;
    esac
}

# Deploy with Docker
deploy_docker() {
    log "Deploying with Docker..."
    
    # Navigate to docker directory
    cd docker
    
    # Stop existing services
    log "Stopping existing services..."
    docker-compose down || true
    
    # Pull latest images
    log "Pulling latest images..."
    docker-compose pull || warn "Some images might need to be built locally"
    
    # Start services
    log "Starting services..."
    docker-compose up -d
    
    # Wait for services to be ready
    log "Waiting for services to start..."
    sleep 15
    
    # Health check
    health_check_docker
    
    cd ..
}

# Deploy from source
deploy_source() {
    log "Deploying from source..."
    
    # Start database services (assuming they're running externally)
    log "Checking database services..."
    if ! nc -z localhost $DATABASE_PORT; then
        warn "MySQL is not running on port $DATABASE_PORT. Please start MySQL manually."
        warn "Example: sudo systemctl start mysql"
    fi
    
    if ! nc -z localhost 6379; then
        warn "Redis is not running on port 6379. Please start Redis manually."
        warn "Example: sudo systemctl start redis"
    fi
    
    # Start backend services
    log "Starting backend services..."
    
    # Start robot service
    if [ -d "backend/robot-service" ]; then
        cd backend/robot-service
        nohup java -jar target/*.jar --spring.config.location=../../config/application.yml > ../../logs/robot-service.log 2>&1 &
        echo $! > ../../robot-service.pid
        cd ../..
    fi
    
    # Start resource service
    if [ -d "backend/resource-service" ]; then
        cd backend/resource-service
        nohup java -jar target/*.jar --server.port=8081 --spring.config.location=../../config/application.yml > ../../logs/resource-service.log 2>&1 &
        echo $! > ../../resource-service.pid
        cd ../..
    fi
    
    # Start Python RPA engine
    log "Starting Python RPA engine..."
    source venv/bin/activate
    cd engine
    nohup python3 -m uvicorn main:app --host 0.0.0.0 --port 8002 > ../logs/rpa-engine.log 2>&1 &
    echo $! > ../rpa-engine.pid
    cd ..
    
    # Start frontend (if in development mode)
    if [ "$NODE_ENV" = "development" ]; then
        log "Starting frontend development server..."
        cd frontend
        nohup pnpm dev:web > ../logs/frontend.log 2>&1 &
        echo $! > ../frontend.pid
        cd ..
    fi
    
    log "Source deployment completed"
    
    # Health check
    health_check_source
}

# Deploy Tauri desktop application
deploy_tauri() {
    log "Deploying Tauri desktop application..."
    
    # Build and run Tauri app
    cd frontend
    
    if [ "$NODE_ENV" = "development" ]; then
        log "Starting Tauri development mode..."
        pnpm dev:tauri
    else
        log "Building Tauri production app..."
        pnpm build:tauri
        
        # The built application will be in src-tauri/target/release/
        log "Tauri application built successfully"
        log "Executable location: src-tauri/target/release/"
        
        # Optionally start the built application
        if [ -f "src-tauri/target/release/astron-rpa" ]; then
            log "Starting Tauri application..."
            nohup ./src-tauri/target/release/astron-rpa > ../logs/tauri-app.log 2>&1 &
            echo $! > ../tauri-app.pid
        fi
    fi
    
    cd ..
}

# Health check for Docker deployment
health_check_docker() {
    log "Performing Docker health check..."
    
    local count=0
    while [ $count -lt $HEALTH_CHECK_TIMEOUT ]; do
        if docker-compose ps | grep -q "Up"; then
            if curl -s "http://localhost:${DEFAULT_PORT}" > /dev/null 2>&1; then
                log "Docker health check passed - services are responding"
                return 0
            fi
        fi
        
        sleep 2
        count=$((count + 2))
    done
    
    error "Docker health check failed - services not responding after ${HEALTH_CHECK_TIMEOUT}s"
    docker-compose logs --tail=20
    return 1
}

# Health check for source deployment
health_check_source() {
    log "Performing source deployment health check..."
    
    local count=0
    while [ $count -lt $HEALTH_CHECK_TIMEOUT ]; do
        if curl -s "http://localhost:${DEFAULT_PORT}" > /dev/null 2>&1; then
            log "Source deployment health check passed"
            return 0
        fi
        
        sleep 2
        count=$((count + 2))
    done
    
    error "Source deployment health check failed"
    return 1
}

# Show service status
show_status() {
    log "Service Status:"
    
    case $DEPLOYMENT_MODE in
        "docker")
            cd docker
            if docker-compose ps | grep -q "Up"; then
                info "✅ Docker services are running"
                docker-compose ps
                info "📊 Web UI: http://localhost:${DEFAULT_PORT}"
            else
                warn "❌ Docker services are not running"
                docker-compose ps
            fi
            cd ..
            ;;
            
        "source")
            local running_services=0
            
            if [ -f "robot-service.pid" ] && kill -0 "$(cat robot-service.pid)" 2>/dev/null; then
                info "✅ Robot service is running (PID: $(cat robot-service.pid))"
                running_services=$((running_services + 1))
            else
                warn "❌ Robot service is not running"
            fi
            
            if [ -f "resource-service.pid" ] && kill -0 "$(cat resource-service.pid)" 2>/dev/null; then
                info "✅ Resource service is running (PID: $(cat resource-service.pid))"
                running_services=$((running_services + 1))
            else
                warn "❌ Resource service is not running"
            fi
            
            if [ -f "rpa-engine.pid" ] && kill -0 "$(cat rpa-engine.pid)" 2>/dev/null; then
                info "✅ RPA engine is running (PID: $(cat rpa-engine.pid))"
                running_services=$((running_services + 1))
            else
                warn "❌ RPA engine is not running"
            fi
            
            if [ -f "frontend.pid" ] && kill -0 "$(cat frontend.pid)" 2>/dev/null; then
                info "✅ Frontend service is running (PID: $(cat frontend.pid))"
                running_services=$((running_services + 1))
            fi
            
            info "📊 Running services: $running_services"
            info "📊 Web UI: http://localhost:${DEFAULT_PORT}"
            ;;
            
        "tauri")
            if [ -f "tauri-app.pid" ] && kill -0 "$(cat tauri-app.pid)" 2>/dev/null; then
                info "✅ Tauri desktop application is running (PID: $(cat tauri-app.pid))"
            else
                info "ℹ️ Tauri application status unknown (desktop app)"
            fi
            ;;
    esac
}

# Stop service
stop_service() {
    log "Stopping service..."
    
    case $DEPLOYMENT_MODE in
        "docker")
            cd docker
            docker-compose down
            cd ..
            ;;
            
        "source")
            # Stop all services
            for pid_file in robot-service.pid resource-service.pid rpa-engine.pid frontend.pid; do
                if [ -f "$pid_file" ]; then
                    local pid=$(cat "$pid_file")
                    if kill "$pid" 2>/dev/null; then
                        log "Stopped service (PID: $pid)"
                    fi
                    rm -f "$pid_file"
                fi
            done
            ;;
            
        "tauri")
            if [ -f "tauri-app.pid" ]; then
                local pid=$(cat "tauri-app.pid")
                if kill "$pid" 2>/dev/null; then
                    log "Stopped Tauri application (PID: $pid)"
                fi
                rm -f "tauri-app.pid"
            fi
            ;;
    esac
}

# Show logs
show_logs() {
    log "Showing logs..."
    
    case $DEPLOYMENT_MODE in
        "docker")
            cd docker
            docker-compose logs -f
            cd ..
            ;;
            
        "source")
            if [ -f "logs/robot-service.log" ]; then
                tail -f logs/robot-service.log
            else
                error "Robot service log file not found"
            fi
            ;;
            
        "tauri")
            if [ -f "logs/tauri-app.log" ]; then
                tail -f logs/tauri-app.log
            else
                error "Tauri application log file not found"
            fi
            ;;
    esac
}

# Run packaging script (Windows-specific)
run_packaging() {
    log "Running packaging script..."
    
    if [ -f "pack.bat" ]; then
        if command -v cmd.exe &> /dev/null; then
            # Running on Windows with WSL
            cmd.exe /c pack.bat
        else
            warn "pack.bat is a Windows batch file. Please run it on Windows or use WSL."
            warn "Alternative: Use the Docker deployment mode for cross-platform compatibility."
        fi
    else
        error "pack.bat not found in repository root"
    fi
}

# Main deployment function
main() {
    log "Starting $PROJECT_NAME deployment (mode: $DEPLOYMENT_MODE)"
    
    # Create logs directory
    mkdir -p logs
    
    check_prerequisites
    setup_repository
    setup_environment
    build_application
    
    case $DEPLOYMENT_MODE in
        "docker")
            deploy_docker
            ;;
        "source")
            deploy_source
            ;;
        "tauri")
            deploy_tauri
            ;;
        *)
            error "Unknown deployment mode: $DEPLOYMENT_MODE"
            exit 1
            ;;
    esac
    
    show_status
    
    log "Deployment completed successfully!"
    log "🤖 Astron RPA Platform is ready!"
    log "📖 Features available:"
    log "   - Visual drag-and-drop process designer"
    log "   - 25+ professional RPA components"
    log "   - AI-powered automation (OCR, CAPTCHA recognition)"
    log "   - Multi-platform support (Web, Desktop, Mobile)"
    log "   - Real-time execution monitoring"
    log "   - Enterprise-grade security and scalability"
    
    if [ "$DEPLOYMENT_MODE" != "tauri" ]; then
        log "🌐 Access the web interface at: http://localhost:${DEFAULT_PORT}"
    fi
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
        sleep 5
        main
        ;;
    "package")
        run_packaging
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo "Commands:"
        echo "  deploy   - Deploy the service (default)"
        echo "  status   - Show service status"
        echo "  stop     - Stop the service"
        echo "  logs     - Show service logs"
        echo "  restart  - Restart the service"
        echo "  package  - Run Windows packaging script"
        echo "  help     - Show this help"
        echo ""
        echo "Environment Variables:"
        echo "  DEPLOYMENT_MODE - docker|source|tauri (default: docker)"
        echo "  DEFAULT_PORT    - Service port (default: 8080)"
        echo "  DATABASE_NAME   - MySQL database name (default: rpa_opensource)"
        echo "  DATABASE_USERNAME - MySQL username (default: root)"
        echo "  DATABASE_PASSWORD - MySQL password (default: 123456)"
        echo "  REDIS_PASSWORD  - Redis password (default: 123456)"
        echo ""
        echo "Deployment Modes:"
        echo "  docker  - Full Docker Compose deployment (recommended)"
        echo "  source  - Build and run from source code"
        echo "  tauri   - Build and run Tauri desktop application"
        ;;
    *)
        error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac


#!/bin/bash

# =============================================================================
# Astron Agent Deployment Script
# =============================================================================
# Description: Comprehensive deployment script for Astron Agent platform
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
PROJECT_NAME="astron-agent"
DEFAULT_PORT="8080"
DOCKER_COMPOSE_PATH="docker/astronAgent/docker-compose.yaml"

# Deployment options
DEPLOYMENT_MODE="${DEPLOYMENT_MODE:-docker}"  # docker, kubernetes, source
JAVA_VERSION_REQUIRED="21"
NODE_VERSION_REQUIRED="18"
PYTHON_VERSION_REQUIRED="3.11"
HEALTH_CHECK_TIMEOUT=60

# Database configuration
DB_TYPE="${DB_TYPE:-postgres}"  # postgres, mysql
POSTGRES_USER="${POSTGRES_USER:-spark}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-spark123}"
POSTGRES_DB="${POSTGRES_DB:-sparkdb_manager}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root123}"

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
            
        "kubernetes")
            # Check kubectl and helm
            if ! command -v kubectl &> /dev/null; then
                error "kubectl is not installed. Please install kubectl first."
                exit 1
            fi
            
            if ! command -v helm &> /dev/null; then
                error "Helm is not installed. Please install Helm first."
                exit 1
            fi
            
            log "Kubernetes environment is ready"
            ;;
            
        "source")
            # Check Java
            if ! command -v java &> /dev/null; then
                error "Java is not installed. Please install Java ${JAVA_VERSION_REQUIRED}+ first."
                exit 1
            fi
            
            local java_version=$(java -version 2>&1 | head -n1 | cut -d'"' -f2 | cut -d'.' -f1)
            if [ "$java_version" -lt "$JAVA_VERSION_REQUIRED" ]; then
                error "Java version $java_version is too old. Required: ${JAVA_VERSION_REQUIRED}+"
                exit 1
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
            
            # Check Python
            if ! command -v python3 &> /dev/null; then
                error "Python 3 is not installed. Please install Python ${PYTHON_VERSION_REQUIRED}+ first."
                exit 1
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
        git clone https://github.com/Zeeeepa/astron-agent.git
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
            if [ ! -f "docker/astronAgent/.env" ]; then
                log "Creating Docker environment file..."
                cp docker/astronAgent/.env.example docker/astronAgent/.env
                
                # Update environment variables
                sed -i "s/POSTGRES_USER=.*/POSTGRES_USER=${POSTGRES_USER}/" docker/astronAgent/.env
                sed -i "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" docker/astronAgent/.env
                sed -i "s/MYSQL_ROOT_PASSWORD=.*/MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}/" docker/astronAgent/.env
            fi
            ;;
            
        "source")
            # Setup source environment
            log "Setting up source build environment..."
            
            # Create application.yml for backend services
            mkdir -p config
            cat > config/application.yml << EOF
server:
  port: ${DEFAULT_PORT}

spring:
  datasource:
    url: jdbc:${DB_TYPE}://localhost:5432/${POSTGRES_DB}
    username: ${POSTGRES_USER}
    password: ${POSTGRES_PASSWORD}
    driver-class-name: org.postgresql.Driver
  
  jpa:
    hibernate:
      ddl-auto: update
    show-sql: false
    
logging:
  level:
    com.iflytek: INFO
    org.springframework: WARN
EOF
            ;;
    esac
    
    log "Environment configuration completed"
}

# Build application components
build_application() {
    log "Building application components..."
    
    case $DEPLOYMENT_MODE in
        "source")
            # Build backend services
            log "Building Java backend services..."
            if [ -d "console/backend" ]; then
                cd console/backend
                ./mvnw clean package -DskipTests
                cd ../..
            fi
            
            # Build frontend
            log "Building TypeScript frontend..."
            if [ -d "console/frontend" ]; then
                cd console/frontend
                npm install
                npm run build
                cd ../..
            fi
            
            # Build Python AI services
            log "Setting up Python AI services..."
            if [ -d "core/plugin/aitools/service" ]; then
                cd core/plugin/aitools/service
                python3 -m pip install -r requirements.txt
                cd ../../../..
            fi
            
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
    cd docker/astronAgent
    
    # Stop existing services
    log "Stopping existing services..."
    docker-compose down || true
    
    # Pull latest images
    log "Pulling latest images..."
    docker-compose pull
    
    # Start services
    log "Starting services..."
    docker-compose up -d
    
    # Wait for services to be ready
    log "Waiting for services to start..."
    sleep 10
    
    # Health check
    health_check_docker
    
    cd ../..
}

# Deploy from source
deploy_source() {
    log "Deploying from source..."
    
    # Start database (assuming PostgreSQL)
    log "Starting database..."
    if ! pgrep -x "postgres" > /dev/null; then
        warn "PostgreSQL is not running. Please start PostgreSQL manually."
        warn "Example: sudo systemctl start postgresql"
    fi
    
    # Start backend services
    log "Starting backend services..."
    if [ -d "console/backend" ]; then
        cd console/backend
        nohup java -jar target/*.jar --spring.config.location=../../config/application.yml > ../../logs/backend.log 2>&1 &
        echo $! > ../../backend.pid
        cd ../..
    fi
    
    # Start frontend (if needed for development)
    if [ "$NODE_ENV" = "development" ] && [ -d "console/frontend" ]; then
        log "Starting frontend development server..."
        cd console/frontend
        nohup npm run dev > ../../logs/frontend.log 2>&1 &
        echo $! > ../../frontend.pid
        cd ../..
    fi
    
    # Start AI services
    if [ -d "core/plugin/aitools/service" ]; then
        log "Starting AI services..."
        cd core/plugin/aitools/service
        nohup python3 -m uvicorn main:app --host 0.0.0.0 --port 8001 > ../../../../logs/ai-service.log 2>&1 &
        echo $! > ../../../../ai-service.pid
        cd ../../../..
    fi
    
    log "Source deployment completed"
    
    # Health check
    health_check_source
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log "Deploying to Kubernetes..."
    
    # Check if Helm charts exist
    if [ ! -d "deploy/helm" ]; then
        error "Helm charts not found. Creating basic Kubernetes manifests..."
        create_kubernetes_manifests
    fi
    
    # Create namespace
    kubectl create namespace "$PROJECT_NAME" --dry-run=client -o yaml | kubectl apply -f -
    
    # Deploy with kubectl or helm
    if [ -d "deploy/helm" ]; then
        helm upgrade --install "$PROJECT_NAME" deploy/helm/"$PROJECT_NAME" \
            --namespace "$PROJECT_NAME" \
            --wait \
            --timeout=600s
    else
        kubectl apply -f deploy/k8s/ -n "$PROJECT_NAME"
    fi
    
    log "Kubernetes deployment completed"
    
    # Health check
    health_check_kubernetes
}

# Create basic Kubernetes manifests
create_kubernetes_manifests() {
    log "Creating basic Kubernetes manifests..."
    
    mkdir -p deploy/k8s
    
    # Create deployment manifest
    cat > deploy/k8s/deployment.yaml << EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: astron-agent
  labels:
    app: astron-agent
spec:
  replicas: 1
  selector:
    matchLabels:
      app: astron-agent
  template:
    metadata:
      labels:
        app: astron-agent
    spec:
      containers:
      - name: astron-agent
        image: astron-agent:latest
        ports:
        - containerPort: 8080
        env:
        - name: SPRING_PROFILES_ACTIVE
          value: "production"
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: astron-agent-service
spec:
  selector:
    app: astron-agent
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
EOF
    
    log "Kubernetes manifests created"
}

# Health check for Docker deployment
health_check_docker() {
    log "Performing Docker health check..."
    
    local count=0
    while [ $count -lt $HEALTH_CHECK_TIMEOUT ]; do
        if docker-compose ps | grep -q "Up"; then
            if curl -s "http://localhost:${DEFAULT_PORT}/health" > /dev/null 2>&1 || \
               curl -s "http://localhost:${DEFAULT_PORT}" > /dev/null 2>&1; then
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
        if curl -s "http://localhost:${DEFAULT_PORT}/health" > /dev/null 2>&1 || \
           curl -s "http://localhost:${DEFAULT_PORT}" > /dev/null 2>&1; then
            log "Source deployment health check passed"
            return 0
        fi
        
        sleep 2
        count=$((count + 2))
    done
    
    error "Source deployment health check failed"
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
        kubectl describe pods -l app="$PROJECT_NAME" -n "$PROJECT_NAME"
        return 1
    fi
}

# Show service status
show_status() {
    log "Service Status:"
    
    case $DEPLOYMENT_MODE in
        "docker")
            cd docker/astronAgent
            if docker-compose ps | grep -q "Up"; then
                info "✅ Docker services are running"
                docker-compose ps
                info "📊 Service URL: http://localhost:${DEFAULT_PORT}"
            else
                warn "❌ Docker services are not running"
                docker-compose ps
            fi
            cd ../..
            ;;
            
        "source")
            local running_services=0
            
            if [ -f "backend.pid" ] && kill -0 "$(cat backend.pid)" 2>/dev/null; then
                info "✅ Backend service is running (PID: $(cat backend.pid))"
                running_services=$((running_services + 1))
            else
                warn "❌ Backend service is not running"
            fi
            
            if [ -f "ai-service.pid" ] && kill -0 "$(cat ai-service.pid)" 2>/dev/null; then
                info "✅ AI service is running (PID: $(cat ai-service.pid))"
                running_services=$((running_services + 1))
            else
                warn "❌ AI service is not running"
            fi
            
            if [ -f "frontend.pid" ] && kill -0 "$(cat frontend.pid)" 2>/dev/null; then
                info "✅ Frontend service is running (PID: $(cat frontend.pid))"
                running_services=$((running_services + 1))
            fi
            
            info "📊 Running services: $running_services"
            ;;
            
        "kubernetes")
            local pods=$(kubectl get pods -l app="$PROJECT_NAME" -n "$PROJECT_NAME" --no-headers 2>/dev/null | wc -l)
            if [ "$pods" -gt 0 ]; then
                info "✅ Kubernetes deployment is running ($pods pods)"
                kubectl get pods -l app="$PROJECT_NAME" -n "$PROJECT_NAME"
                
                # Get service URL
                local service_url=$(kubectl get svc "$PROJECT_NAME-service" -n "$PROJECT_NAME" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null)
                if [ -n "$service_url" ]; then
                    info "📊 Service URL: http://$service_url"
                fi
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
        "docker")
            cd docker/astronAgent
            docker-compose down
            cd ../..
            ;;
            
        "source")
            # Stop all services
            for pid_file in backend.pid ai-service.pid frontend.pid; do
                if [ -f "$pid_file" ]; then
                    local pid=$(cat "$pid_file")
                    if kill "$pid" 2>/dev/null; then
                        log "Stopped service (PID: $pid)"
                    fi
                    rm -f "$pid_file"
                fi
            done
            ;;
            
        "kubernetes")
            if command -v helm &> /dev/null; then
                helm uninstall "$PROJECT_NAME" -n "$PROJECT_NAME" || warn "Helm release not found"
            else
                kubectl delete -f deploy/k8s/ -n "$PROJECT_NAME" || warn "Kubernetes resources not found"
            fi
            kubectl delete namespace "$PROJECT_NAME" || warn "Namespace not found"
            ;;
    esac
}

# Show logs
show_logs() {
    log "Showing logs..."
    
    case $DEPLOYMENT_MODE in
        "docker")
            cd docker/astronAgent
            docker-compose logs -f
            cd ../..
            ;;
            
        "source")
            if [ -f "logs/backend.log" ]; then
                tail -f logs/backend.log
            else
                error "Backend log file not found"
            fi
            ;;
            
        "kubernetes")
            kubectl logs -f -l app="$PROJECT_NAME" -n "$PROJECT_NAME"
            ;;
    esac
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
    log "🌟 Astron Agent Platform is ready!"
    log "📖 Access the documentation at: http://localhost:${DEFAULT_PORT}/docs"
    log "🎯 Main features:"
    log "   - Enterprise-grade Agent development platform"
    log "   - Intelligent RPA integration"
    log "   - Multi-language backend support (Java, Go, Python)"
    log "   - TypeScript + React frontend"
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
        echo "  DEPLOYMENT_MODE - docker|kubernetes|source (default: docker)"
        echo "  DB_TYPE         - postgres|mysql (default: postgres)"
        echo "  DEFAULT_PORT    - Service port (default: 8080)"
        echo "  POSTGRES_USER   - PostgreSQL username (default: spark)"
        echo "  POSTGRES_PASSWORD - PostgreSQL password (default: spark123)"
        ;;
    *)
        error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac


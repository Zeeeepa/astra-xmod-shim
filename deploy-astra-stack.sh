#!/bin/bash

# =============================================================================
# Astra Ecosystem Stack Deployment Script
# =============================================================================
# Description: Master orchestration script for deploying the complete Astra ecosystem
# Components: astra-xmod-shim, astron-agent, astron-rpa
# Author: Codegen AI Assistant
# Version: 1.0.0
# =============================================================================

set -euo pipefail

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STACK_NAME="astra-ecosystem"
DEPLOYMENT_MODE="${DEPLOYMENT_MODE:-docker}"  # docker, kubernetes, source, mixed

# Component configuration
COMPONENTS=("astra-xmod-shim" "astron-agent" "astron-rpa")
COMPONENT_PORTS=("7777" "8080" "8080")
COMPONENT_HEALTH_ENDPOINTS=("/api/v1/plugins" "/health" "/")

# Deployment order and dependencies
DEPLOYMENT_ORDER=("astra-xmod-shim" "astron-agent" "astron-rpa")
STARTUP_DELAYS=(10 15 20)  # Seconds to wait between component startups

# Global settings
HEALTH_CHECK_TIMEOUT=120
PARALLEL_DEPLOYMENT="${PARALLEL_DEPLOYMENT:-false}"
SKIP_HEALTH_CHECKS="${SKIP_HEALTH_CHECKS:-false}"
DRY_RUN="${DRY_RUN:-false}"

# Logging functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] [STACK] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] [STACK] $1${NC}" >&2
}

warn() {
    echo -e "${YELLOW}[WARN] [STACK] $1${NC}"
}

info() {
    echo -e "${BLUE}[INFO] [STACK] $1${NC}"
}

component_log() {
    local component=$1
    shift
    echo -e "${PURPLE}[$(date +'%Y-%m-%d %H:%M:%S')] [$component] $*${NC}"
}

# Banner function
show_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
    ╔═══════════════════════════════════════════════════════════════╗
    ║                                                               ║
    ║                🚀 ASTRA ECOSYSTEM DEPLOYMENT 🚀               ║
    ║                                                               ║
    ║  Components:                                                  ║
    ║  • astra-xmod-shim  - AI Service Orchestration Middleware    ║
    ║  • astron-agent     - Enterprise Agent Development Platform  ║
    ║  • astron-rpa       - Robotic Process Automation Platform    ║
    ║                                                               ║
    ╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Check prerequisites for all components
check_global_prerequisites() {
    log "Checking global prerequisites..."
    
    case $DEPLOYMENT_MODE in
        "docker")
            if ! command -v docker &> /dev/null; then
                error "Docker is not installed. Please install Docker first."
                exit 1
            fi
            
            if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
                error "Docker Compose is not installed. Please install Docker Compose first."
                exit 1
            fi
            
            # Check Docker daemon
            if ! docker info &> /dev/null; then
                error "Docker daemon is not running. Please start Docker first."
                exit 1
            fi
            
            log "Docker environment is ready"
            ;;
            
        "kubernetes")
            if ! command -v kubectl &> /dev/null; then
                error "kubectl is not installed. Please install kubectl first."
                exit 1
            fi
            
            if ! command -v helm &> /dev/null; then
                error "Helm is not installed. Please install Helm first."
                exit 1
            fi
            
            # Check cluster connectivity
            if ! kubectl cluster-info &> /dev/null; then
                error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
                exit 1
            fi
            
            log "Kubernetes environment is ready"
            ;;
            
        "source"|"mixed")
            # Check basic tools
            local missing_tools=()
            
            if ! command -v git &> /dev/null; then
                missing_tools+=("git")
            fi
            
            if ! command -v curl &> /dev/null; then
                missing_tools+=("curl")
            fi
            
            if ! command -v jq &> /dev/null; then
                missing_tools+=("jq")
            fi
            
            if [ ${#missing_tools[@]} -gt 0 ]; then
                error "Missing required tools: ${missing_tools[*]}"
                error "Please install these tools before proceeding."
                exit 1
            fi
            
            log "Basic tools are available"
            ;;
    esac
    
    # Check available ports
    check_port_availability
}

# Check if required ports are available
check_port_availability() {
    log "Checking port availability..."
    
    local unavailable_ports=()
    
    for i in "${!COMPONENT_PORTS[@]}"; do
        local port="${COMPONENT_PORTS[$i]}"
        local component="${COMPONENTS[$i]}"
        
        if netstat -tuln 2>/dev/null | grep -q ":$port " || ss -tuln 2>/dev/null | grep -q ":$port "; then
            warn "Port $port is already in use (needed for $component)"
            unavailable_ports+=("$port:$component")
        fi
    done
    
    if [ ${#unavailable_ports[@]} -gt 0 ]; then
        warn "The following ports are in use:"
        for port_info in "${unavailable_ports[@]}"; do
            warn "  - ${port_info}"
        done
        
        if [ "$DRY_RUN" = "false" ]; then
            read -p "Continue anyway? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                error "Deployment cancelled due to port conflicts"
                exit 1
            fi
        fi
    else
        log "All required ports are available"
    fi
}

# Create deployment workspace
setup_workspace() {
    log "Setting up deployment workspace..."
    
    local workspace_dir="astra-ecosystem-deployment"
    
    if [ ! -d "$workspace_dir" ]; then
        mkdir -p "$workspace_dir"
        cd "$workspace_dir"
        
        # Create directory structure
        mkdir -p logs config scripts
        
        # Copy deployment scripts
        cp "$SCRIPT_DIR"/deploy-*.sh scripts/ 2>/dev/null || true
        
        log "Workspace created: $(pwd)"
    else
        cd "$workspace_dir"
        log "Using existing workspace: $(pwd)"
    fi
    
    # Create master log file
    exec 1> >(tee -a logs/stack-deployment.log)
    exec 2> >(tee -a logs/stack-deployment.log >&2)
}

# Deploy individual component
deploy_component() {
    local component=$1
    local index=$2
    
    component_log "$component" "Starting deployment..."
    
    if [ "$DRY_RUN" = "true" ]; then
        component_log "$component" "[DRY RUN] Would deploy $component"
        return 0
    fi
    
    # Set component-specific environment variables
    case $component in
        "astra-xmod-shim")
            export DEPLOYMENT_MODE="$DEPLOYMENT_MODE"
            export DEFAULT_PORT="${COMPONENT_PORTS[$index]}"
            ;;
        "astron-agent")
            export DEPLOYMENT_MODE="$DEPLOYMENT_MODE"
            export DEFAULT_PORT="${COMPONENT_PORTS[$index]}"
            export DB_TYPE="postgres"
            ;;
        "astron-rpa")
            export DEPLOYMENT_MODE="$DEPLOYMENT_MODE"
            export DEFAULT_PORT="${COMPONENT_PORTS[$index]}"
            ;;
    esac
    
    # Run component deployment script
    local script_path="../scripts/deploy-${component}.sh"
    
    if [ -f "$script_path" ]; then
        component_log "$component" "Executing deployment script..."
        
        # Run deployment in background if parallel deployment is enabled
        if [ "$PARALLEL_DEPLOYMENT" = "true" ]; then
            (
                bash "$script_path" deploy 2>&1 | while IFS= read -r line; do
                    component_log "$component" "$line"
                done
            ) &
            
            # Store PID for later waiting
            eval "${component}_pid=$!"
        else
            bash "$script_path" deploy 2>&1 | while IFS= read -r line; do
                component_log "$component" "$line"
            done
            
            if [ $? -eq 0 ]; then
                component_log "$component" "✅ Deployment completed successfully"
            else
                component_log "$component" "❌ Deployment failed"
                return 1
            fi
        fi
    else
        error "Deployment script not found: $script_path"
        return 1
    fi
}

# Wait for parallel deployments to complete
wait_for_parallel_deployments() {
    if [ "$PARALLEL_DEPLOYMENT" = "true" ]; then
        log "Waiting for parallel deployments to complete..."
        
        local failed_components=()
        
        for component in "${COMPONENTS[@]}"; do
            local pid_var="${component//-/_}_pid"
            local pid=${!pid_var:-}
            
            if [ -n "$pid" ]; then
                component_log "$component" "Waiting for deployment to complete (PID: $pid)..."
                
                if wait "$pid"; then
                    component_log "$component" "✅ Deployment completed successfully"
                else
                    component_log "$component" "❌ Deployment failed"
                    failed_components+=("$component")
                fi
            fi
        done
        
        if [ ${#failed_components[@]} -gt 0 ]; then
            error "The following components failed to deploy: ${failed_components[*]}"
            return 1
        fi
        
        log "All parallel deployments completed successfully"
    fi
}

# Health check for individual component
health_check_component() {
    local component=$1
    local index=$2
    local port="${COMPONENT_PORTS[$index]}"
    local endpoint="${COMPONENT_HEALTH_ENDPOINTS[$index]}"
    
    if [ "$SKIP_HEALTH_CHECKS" = "true" ]; then
        component_log "$component" "⏭️ Health check skipped"
        return 0
    fi
    
    component_log "$component" "Performing health check on port $port..."
    
    local count=0
    local url="http://localhost:${port}${endpoint}"
    
    while [ $count -lt $HEALTH_CHECK_TIMEOUT ]; do
        if curl -s --max-time 5 "$url" > /dev/null 2>&1; then
            component_log "$component" "✅ Health check passed"
            return 0
        fi
        
        sleep 2
        count=$((count + 2))
        
        if [ $((count % 20)) -eq 0 ]; then
            component_log "$component" "⏳ Still waiting for health check... (${count}s/${HEALTH_CHECK_TIMEOUT}s)"
        fi
    done
    
    component_log "$component" "❌ Health check failed after ${HEALTH_CHECK_TIMEOUT}s"
    return 1
}

# Deploy all components
deploy_stack() {
    log "Starting stack deployment (mode: $DEPLOYMENT_MODE, parallel: $PARALLEL_DEPLOYMENT)"
    
    if [ "$PARALLEL_DEPLOYMENT" = "true" ]; then
        # Deploy all components in parallel
        for i in "${!COMPONENTS[@]}"; do
            deploy_component "${COMPONENTS[$i]}" "$i"
        done
        
        # Wait for all deployments to complete
        wait_for_parallel_deployments
        
        # Perform health checks
        log "Performing health checks for all components..."
        local failed_health_checks=()
        
        for i in "${!COMPONENTS[@]}"; do
            if ! health_check_component "${COMPONENTS[$i]}" "$i"; then
                failed_health_checks+=("${COMPONENTS[$i]}")
            fi
        done
        
        if [ ${#failed_health_checks[@]} -gt 0 ]; then
            error "Health checks failed for: ${failed_health_checks[*]}"
            return 1
        fi
        
    else
        # Deploy components sequentially
        for i in "${!DEPLOYMENT_ORDER[@]}"; do
            local component="${DEPLOYMENT_ORDER[$i]}"
            local component_index
            
            # Find component index
            for j in "${!COMPONENTS[@]}"; do
                if [ "${COMPONENTS[$j]}" = "$component" ]; then
                    component_index=$j
                    break
                fi
            done
            
            # Deploy component
            if ! deploy_component "$component" "$component_index"; then
                error "Failed to deploy $component"
                return 1
            fi
            
            # Wait before next deployment
            if [ $i -lt $((${#DEPLOYMENT_ORDER[@]} - 1)) ]; then
                local delay="${STARTUP_DELAYS[$i]}"
                log "Waiting ${delay}s before deploying next component..."
                sleep "$delay"
            fi
            
            # Health check
            if ! health_check_component "$component" "$component_index"; then
                error "Health check failed for $component"
                return 1
            fi
        done
    fi
    
    log "✅ Stack deployment completed successfully!"
}

# Show stack status
show_stack_status() {
    log "Astra Ecosystem Stack Status:"
    echo
    
    local all_healthy=true
    
    for i in "${!COMPONENTS[@]}"; do
        local component="${COMPONENTS[$i]}"
        local port="${COMPONENT_PORTS[$i]}"
        local endpoint="${COMPONENT_HEALTH_ENDPOINTS[$i]}"
        local url="http://localhost:${port}${endpoint}"
        
        printf "%-20s " "$component:"
        
        if curl -s --max-time 3 "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Running${NC} (http://localhost:$port)"
        else
            echo -e "${RED}❌ Not responding${NC}"
            all_healthy=false
        fi
    done
    
    echo
    if [ "$all_healthy" = "true" ]; then
        info "🎉 All components are healthy!"
        echo
        info "🌟 Astra Ecosystem URLs:"
        info "   • Middleware API:    http://localhost:7777/api/v1/plugins"
        info "   • Agent Platform:    http://localhost:8080"
        info "   • RPA Studio:        http://localhost:8080 (if different port configured)"
        echo
        info "📚 Integration Flow:"
        info "   1. astra-xmod-shim provides AI service orchestration"
        info "   2. astron-agent manages enterprise agent workflows"
        info "   3. astron-rpa handles robotic process automation"
        info "   4. Components communicate via REST APIs and event buses"
    else
        warn "⚠️ Some components are not responding. Check individual component logs."
    fi
}

# Stop all components
stop_stack() {
    log "Stopping Astra ecosystem stack..."
    
    # Stop in reverse order
    for ((i=${#DEPLOYMENT_ORDER[@]}-1; i>=0; i--)); do
        local component="${DEPLOYMENT_ORDER[$i]}"
        local script_path="../scripts/deploy-${component}.sh"
        
        if [ -f "$script_path" ]; then
            component_log "$component" "Stopping..."
            
            if [ "$DRY_RUN" = "true" ]; then
                component_log "$component" "[DRY RUN] Would stop $component"
            else
                bash "$script_path" stop 2>&1 | while IFS= read -r line; do
                    component_log "$component" "$line"
                done
            fi
        fi
    done
    
    log "Stack stopped"
}

# Show logs for all components
show_stack_logs() {
    log "Showing logs for all components..."
    
    # Create a combined log viewer
    local log_files=()
    
    for component in "${COMPONENTS[@]}"; do
        local log_file="logs/${component}.log"
        if [ -f "$log_file" ]; then
            log_files+=("$log_file")
        fi
    done
    
    if [ ${#log_files[@]} -gt 0 ]; then
        tail -f "${log_files[@]}"
    else
        warn "No log files found. Components may not be running."
    fi
}

# Create Docker Compose for entire stack
create_stack_compose() {
    log "Creating Docker Compose configuration for entire stack..."
    
    cat > docker-compose.stack.yml << 'EOF'
version: '3.8'

networks:
  astra-network:
    driver: bridge

volumes:
  postgres_data:
  mysql_data:
  redis_data:

services:
  # Shared infrastructure
  postgres:
    image: postgres:14
    container_name: astra-postgres
    environment:
      POSTGRES_DB: sparkdb_manager
      POSTGRES_USER: spark
      POSTGRES_PASSWORD: spark123
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - astra-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U spark -d sparkdb_manager"]
      interval: 30s
      timeout: 10s
      retries: 5

  mysql:
    image: mysql:8.4
    container_name: astra-mysql
    environment:
      MYSQL_ROOT_PASSWORD: root123
      MYSQL_DATABASE: rpa_opensource
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - astra-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 30s
      timeout: 10s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: astra-redis
    command: redis-server --requirepass 123456
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - astra-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5

  # Astra XMod Shim
  astra-xmod-shim:
    build:
      context: ./astra-xmod-shim
      dockerfile: deploy/docker/Dockerfile
    container_name: astra-xmod-shim
    ports:
      - "7777:7777"
    volumes:
      - ./astra-xmod-shim/conf:/opt/modserv-shim/conf
      - ./astra-xmod-shim/logs:/opt/modserv-shim/logs
    networks:
      - astra-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7777/api/v1/plugins"]
      interval: 30s
      timeout: 10s
      retries: 5
    depends_on:
      - postgres
      - redis

  # Astron Agent
  astron-agent:
    build:
      context: ./astron-agent
      dockerfile: Dockerfile
    container_name: astron-agent
    ports:
      - "8080:8080"
    environment:
      - SPRING_PROFILES_ACTIVE=docker
      - POSTGRES_HOST=postgres
      - REDIS_HOST=redis
    networks:
      - astra-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 5
    depends_on:
      - postgres
      - redis
      - astra-xmod-shim

  # Astron RPA
  astron-rpa:
    build:
      context: ./astron-rpa
      dockerfile: Dockerfile
    container_name: astron-rpa
    ports:
      - "8081:8080"
    environment:
      - MYSQL_HOST=mysql
      - REDIS_HOST=redis
    networks:
      - astra-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080"]
      interval: 30s
      timeout: 10s
      retries: 5
    depends_on:
      - mysql
      - redis
      - astron-agent
EOF
    
    log "Docker Compose configuration created: docker-compose.stack.yml"
    info "To use: docker-compose -f docker-compose.stack.yml up -d"
}

# Main function
main() {
    show_banner
    
    log "Initializing Astra ecosystem deployment..."
    log "Deployment mode: $DEPLOYMENT_MODE"
    log "Parallel deployment: $PARALLEL_DEPLOYMENT"
    log "Dry run: $DRY_RUN"
    
    check_global_prerequisites
    setup_workspace
    
    if [ "$DRY_RUN" = "true" ]; then
        log "🔍 DRY RUN MODE - No actual deployments will be performed"
    fi
    
    deploy_stack
    
    echo
    show_stack_status
    
    log "🎉 Astra ecosystem deployment completed successfully!"
    log "📝 Logs are available in: $(pwd)/logs/"
    log "🔧 Individual component management scripts are in: $(pwd)/scripts/"
}

# Command line interface
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "status")
        setup_workspace
        show_stack_status
        ;;
    "stop")
        setup_workspace
        stop_stack
        ;;
    "logs")
        setup_workspace
        show_stack_logs
        ;;
    "restart")
        setup_workspace
        stop_stack
        sleep 10
        main
        ;;
    "compose")
        setup_workspace
        create_stack_compose
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deploy   - Deploy the entire Astra ecosystem (default)"
        echo "  status   - Show status of all components"
        echo "  stop     - Stop all components"
        echo "  logs     - Show logs for all components"
        echo "  restart  - Restart the entire stack"
        echo "  compose  - Generate Docker Compose file for the stack"
        echo "  help     - Show this help"
        echo ""
        echo "Environment Variables:"
        echo "  DEPLOYMENT_MODE      - docker|kubernetes|source|mixed (default: docker)"
        echo "  PARALLEL_DEPLOYMENT  - true|false (default: false)"
        echo "  SKIP_HEALTH_CHECKS   - true|false (default: false)"
        echo "  DRY_RUN             - true|false (default: false)"
        echo "  HEALTH_CHECK_TIMEOUT - Health check timeout in seconds (default: 120)"
        echo ""
        echo "Examples:"
        echo "  $0 deploy                                    # Deploy with defaults"
        echo "  DEPLOYMENT_MODE=kubernetes $0 deploy         # Deploy to Kubernetes"
        echo "  PARALLEL_DEPLOYMENT=true $0 deploy           # Deploy components in parallel"
        echo "  DRY_RUN=true $0 deploy                      # Dry run mode"
        echo ""
        echo "Components:"
        echo "  • astra-xmod-shim  - AI Service Orchestration Middleware (port 7777)"
        echo "  • astron-agent     - Enterprise Agent Development Platform (port 8080)"
        echo "  • astron-rpa       - Robotic Process Automation Platform (port 8080)"
        ;;
    *)
        error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac


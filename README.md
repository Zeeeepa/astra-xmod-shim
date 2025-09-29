# Astra Ecosystem Deployment Scripts

## 🚀 Quick Start

Deploy the complete Astra ecosystem with a single command:

```bash
./deploy-astra-stack.sh deploy
```

## 📦 Components

This repository contains deployment scripts for the complete Astra ecosystem:

- **astra-xmod-shim** - AI Service Orchestration Middleware (Port 7777)
- **astron-agent** - Enterprise Agent Development Platform (Port 8080)  
- **astron-rpa** - Robotic Process Automation Platform (Port 8080)

## 🛠️ Individual Component Deployment

Deploy components individually:

```bash
# Deploy middleware
./deploy-astra-xmod-shim.sh deploy

# Deploy agent platform
./deploy-astron-agent.sh deploy

# Deploy RPA platform
./deploy-astron-rpa.sh deploy
```

## 🔧 Configuration

### Environment Variables

```bash
# Deployment mode
export DEPLOYMENT_MODE=docker  # docker, kubernetes, source

# Enable parallel deployment (faster)
export PARALLEL_DEPLOYMENT=true

# Skip health checks (for testing)
export SKIP_HEALTH_CHECKS=false

# Dry run mode (preview only)
export DRY_RUN=true
```

### Database Configuration

```bash
# PostgreSQL (astron-agent)
export POSTGRES_USER=spark
export POSTGRES_PASSWORD=spark123
export POSTGRES_DB=sparkdb_manager

# MySQL (astron-rpa)
export DATABASE_NAME=rpa_opensource
export DATABASE_USERNAME=root
export DATABASE_PASSWORD=123456

# Redis (shared)
export REDIS_PASSWORD=123456
```

## 📊 Deployment Modes

### 🐳 Docker (Recommended)
- **Best for**: Development, testing, small-scale production
- **Requirements**: Docker, Docker Compose
- **Command**: `DEPLOYMENT_MODE=docker ./deploy-astra-stack.sh deploy`

### ☸️ Kubernetes
- **Best for**: Production, enterprise environments
- **Requirements**: kubectl, helm, Kubernetes cluster
- **Command**: `DEPLOYMENT_MODE=kubernetes ./deploy-astra-stack.sh deploy`

### 🔧 Source
- **Best for**: Development, customization
- **Requirements**: Go 1.24+, Java 21+, Node.js 22+, Python 3.13+
- **Command**: `DEPLOYMENT_MODE=source ./deploy-astra-stack.sh deploy`

## 📋 Management Commands

```bash
# Check status
./deploy-astra-stack.sh status

# View logs
./deploy-astra-stack.sh logs

# Stop all services
./deploy-astra-stack.sh stop

# Restart all services
./deploy-astra-stack.sh restart

# Generate Docker Compose file
./deploy-astra-stack.sh compose
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  astron-agent   │───▶│ astra-xmod-shim │───▶│   astron-rpa    │
│   (Port 8080)   │    │   (Port 7777)   │    │   (Port 8080)   │
│                 │    │                 │    │                 │
│ Agent Platform  │    │   Middleware    │    │  RPA Platform   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │      Redis      │    │      MySQL      │
│   (Port 5432)   │    │   (Port 6379)   │    │   (Port 3306)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🔍 Health Checks

After deployment, verify all components are healthy:

```bash
# Check middleware
curl http://localhost:7777/api/v1/plugins

# Check agent platform
curl http://localhost:8080/health

# Check RPA platform
curl http://localhost:8080/
```

## 📚 Documentation

- **[Complete Analysis](ASTRA_ECOSYSTEM_ANALYSIS.md)** - Comprehensive ecosystem analysis
- **Component Documentation**:
  - [astra-xmod-shim](https://github.com/Zeeeepa/astra-xmod-shim)
  - [astron-agent](https://github.com/Zeeeepa/astron-agent)
  - [astron-rpa](https://github.com/Zeeeepa/astron-rpa)

## 🆘 Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check what's using the port
   netstat -tuln | grep :8080
   
   # Change port if needed
   export DEFAULT_PORT=8081
   ```

2. **Database Connection Issues**
   ```bash
   # Test database connectivity
   nc -zv localhost 5432  # PostgreSQL
   nc -zv localhost 3306  # MySQL
   ```

3. **Memory Issues**
   ```bash
   # Check available memory
   free -h
   
   # Increase Docker memory if needed
   ```

### Getting Help

- **View Logs**: `./deploy-astra-stack.sh logs`
- **Check Status**: `./deploy-astra-stack.sh status`
- **Restart Services**: `./deploy-astra-stack.sh restart`

## 🎯 Integration Examples

### Deploy AI Model via astron-agent
```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "text-classifier",
    "type": "model-deployment",
    "configuration": {
      "modelName": "bert-base",
      "endpoint": "http://localhost:7777/api/v1/modserv/deploy"
    }
  }'
```

### Execute RPA Workflow
```bash
curl -X POST http://localhost:8080/api/workflows/execute \
  -H "Content-Type: application/json" \
  -d '{
    "workflowId": "data-extraction",
    "parameters": {
      "source": "web-form",
      "target": "database"
    }
  }'
```

## 🌟 Features

### astra-xmod-shim
- ✅ AI model deployment and management
- ✅ Kubernetes and Docker orchestration
- ✅ Plugin-based architecture
- ✅ Event-driven observability

### astron-agent
- ✅ Enterprise agent development
- ✅ Multi-language backend support
- ✅ Intelligent RPA integration
- ✅ One-click deployment

### astron-rpa
- ✅ Visual drag-and-drop designer
- ✅ 25+ professional RPA components
- ✅ AI-powered automation
- ✅ Multi-platform support

---

**🎉 Ready to deploy? Run `./deploy-astra-stack.sh deploy` to get started!**


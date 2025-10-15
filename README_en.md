<div align="center">
<img src="xmod-shim.svg?v=2" alt="Astron-xmod-shim Logo" width="600" />
<br>

[![License](https://img.shields.io/github/license/iflytek/Astron-xmod-shim)](https://github.com/iflytek/Astron-xmod-shim/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/iflytek/Astron-xmod-shim?include_prereleases)](https://github.com/iflytek/Astron-xmod-shim/releases)
[![CI Status](https://img.shields.io/github/actions/workflow/status/iflytek/Astron-xmod-shim/ci.yml?branch=main)](https://github.com/iflytek/Astron-xmod-shim/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/iflytek/Astron-xmod-shim)](https://github.com/iflytek/Astron-xmod-shim/blob/main/go.mod)
[![Coverage](https://img.shields.io/codecov/c/github/iflytek/Astron-xmod-shim)](https://codecov.io/gh/iflytek/Astron-xmod-shim)
[![Multi-Arch](https://img.shields.io/badge/Multi--Arch-linux%2Famd64%20%7C%20linux%2Farm64-blue?logo=docker)]()
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Native-blue?logo=kubernetes&logoColor=white)](docs/k8s.md)
[![Helm](https://img.shields.io/badge/Helm-Chart-blue?logo=helm&logoColor=white)](charts/)
[![Cloud Native](https://img.shields.io/badge/Cloud-Native-blue?logo=cloudnative&logoColor=white)](https://cncf.io)
[![Metrics](https://img.shields.io/badge/Metrics-Prometheus-green?logo=prometheus)](docs/metrics.md)
[![Contributors](https://img.shields.io/github/contributors/iflytek/Astron-xmod-shim)](https://github.com/iflytek/Astron-xmod-shim/graphs/contributors)
[![Stars](https://img.shields.io/github/stars/iflytek/Astron-xmod-shim?style=social)](https://github.com/iflytek/Astron-xmod-shim)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](http://makeapullrequest.com)

<span style="font-size:0.9em; color:#586375;">**Language**: **English** | [简体中文](README.md)</span>
</div>

# Astron-xmod-shim

A lightweight, declarative middleware for AI service orchestration.

## Project Overview

**Astron-xmod-shim** is a **goal-driven**, declarative middleware for managing AI services. It translates user-declared `DeploySpec` into verifiable and idempotent `GoalSet`s—supporting standard scenarios like LLM deployment out-of-the-box, while also enabling third-party extensions for custom `GoalSet`s. Each `Goal` is executed by a **Shimlet** plugin that interfaces with underlying runtimes (e.g., Kubernetes, Docker), ensuring reliable delivery across environments through a unified convergence engine.

## 🌟 Core Design Philosophy: From Intent to Eventual Consistency

Astron-xmod-shim is built around one core idea: **deployment is convergence toward a set of explicit goals**.  
The system does **not** dictate *what to check*, but rather *how to reliably converge to your defined goals*.

- **Deployment Intent: `DeploySpec` (User-facing)**  
  Users declare *what they want* via `DeploySpec`, for example:
  > “Deploy a model service named `qwen-test` with 1 replica, using 1 NVIDIA GPU, and model `qwen3-1.5b`.”  
  `DeploySpec` is purely intent-based—free of implementation details or environment binding—ensuring a clean, platform-agnostic interface.

- **`Goal`, `GoalSet`, and Execution Engine**
    1. A **`Goal`** represents a specific system target or convergence step (e.g., “model file exists”) and includes:
        - `IsAchieved()`: checks if the goal is already met;
        - `Ensure()`: performs an idempotent action to achieve the goal if not.
    2. A **`GoalSet`** is an ordered collection of `Goal`s representing the convergence path for a deployment scenario (e.g., LLM rollout, service teardown). It is fully extensible by third parties.
    3. The **execution engine** consists of a **`WorkQueue` + `reconcile loop`**:
        - `WorkQueue` provides reliable task scheduling (deduplication, rate-limited retries, backpressure);
        - `reconcile loop` continuously processes tasks, converging each `Goal` until system state matches intent.

- **`Shimlet` (Runtime Adapter Plugin)**  
  A `Shimlet` implements the `shim.Runtime` interface, abstracting environment-specific operations (e.g., Kubernetes, Docker). This decouples the core engine from runtime details, enabling seamless multi-environment support.

- **Lightweight Monolithic Architecture**  
  Delivered as a single binary with no external dependencies—ideal for edge, local, and cloud-native deployments.

## 🏗️ Technical Architecture

Astron-xmod-shim adopts a **“core engine + dual plugin”** architecture, separating concerns through abstraction layers and a workflow engine for high extensibility and runtime agnosticism.

[Architecture Diagram](xmod-shim.diagram.svg)

## Quick Start

### Prerequisites

- Go 1.24+ (for development)
- Target runtime (e.g., Kubernetes v1.19+, if using the K8s Shimlet)

### Helm Deployment (Kubernetes)

Astron-xmod-shim provides a Helm Chart for Kubernetes deployment.

#### Requirements

- Helm 3.x installed
- `kubectl` configured to access your target Kubernetes cluster
- Host directories for config and models already exist

#### Deployment Commands

```bash
# Navigate to Helm chart directory
cd deploy/helm

# Install or upgrade
helm upgrade --install astron-xmod-shim astron-xmod-shim/ -f astron-xmod-shim/values.yaml

# Verify deployment
kubectl get pods -l app.kubernetes.io/name=astron-xmod-shim
```

#### Helm Chart Features

- Runs in host network mode
- Mounts config directory to `/app/conf` in the container
- Mounts host model directory `/mnt/maasmodels/` to the same path in the container
- Supports mounting K8s Shimlet config files

#### Custom Configuration

To override defaults:

1. **Edit `values.yaml` directly**
   ```bash
   vi deploy/helm/astron-xmod-shim/values.yaml
   ```

2. **Use a custom values file**
   ```bash
   helm upgrade --install astron-xmod-shim deploy/helm/astron-xmod-shim/ -f your-custom-values.yaml
   ```

#### Uninstall

```bash
helm uninstall astron-xmod-shim
```

## API Reference

### Deploy a Model Service

```bash
curl -X POST http://localhost:8080/api/v1/modserv/deploy \
  -H "Content-Type: application/json" \
  -d '{
    "modelName": "example-model",
    "modelFile": "/path/to/model",
    "resourceRequirements": {
      "acceleratorType": "NVIDIA GPU",
      "acceleratorCount": 1,
      "cpu": "4",
      "memory": "16Gi"
    },
    "replicaCount": 1
  }'
```

### Query Service Status

```bash
curl http://localhost:8080/api/v1/modserv/{serviceId}
```

### List Loaded Plugins

```bash
curl http://localhost:8080/api/v1/plugins
```

## Plugin Development Guide

### Shimlet Development (Runtime Adapter)

A Shimlet translates abstract deployment requests into concrete runtime operations. Below is an example of a custom Shimlet for Docker.

#### Built-in Example: Kubernetes Shimlet

Astron-xmod-shim includes a native Kubernetes Shimlet that maps deployment specs to Kubernetes resources (e.g., Deployments, Services).

#### Extension Example: Docker Shimlet

The following demonstrates how to implement a Docker Shimlet to deploy model services in Docker containers.

##### Step 1: Implement the Shimlet Interface

```go
package dockershimlet

import (
	"astron-xmod-shim/internal/core/shimlet"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/log"
	// Import Docker SDK packages as needed
	// "github.com/docker/docker/client"
)

// DockerShimlet implements the Docker runtime adapter
type DockerShimlet struct {
	// Add Docker client or other fields as needed
	// client *client.Client
}

// Compile-time interface check
var _ shimlet.Shimlet = (*DockerShimlet)(nil)

func (d *DockerShimlet) InitWithConfig(confPath string) error {
	log.Info("Initializing Docker shimlet with config: %s", confPath)
	// Initialize Docker client, parse config, validate connection
	return nil
}

func (d *DockerShimlet) Apply(spec *dto.RequirementSpec) error {
	log.Info("Applying deployment spec to Docker: %s", spec.ModelName)
	// Create or update Docker container based on spec
	return nil
}

func (d *DockerShimlet) Delete(resourceID string) error {
	log.Info("Deleting Docker resource: %s", resourceID)
	// Stop and remove container
	return nil
}

func (d *DockerShimlet) Status(resourceID string) (*dto.RuntimeStatus, error) {
	log.Info("Getting status for Docker resource: %s", resourceID)
	// Query container status and return RuntimeStatus
	return &dto.RuntimeStatus{}, nil
}

func (d *DockerShimlet) ID() string {
	return "docker"
}

func (d *DockerShimlet) Description() string {
	return "Docker runtime adapter for deploying model services in containers"
}

func (d *DockerShimlet) ListDeployedServices() ([]string, error) {
	log.Info("Listing all deployed services in Docker")
	// Return list of service IDs (container names or labels)
	return []string{}, nil
}
```

##### Step 2: Register the Shimlet

```go
package dockershimlet

import "astron-xmod-shim/internal/core/shimlet"

func init() {
	shimlet.Registry.AutoRegister(&DockerShimlet{})
}
```

### Predefined GoalSets

GoalSets define the convergence logic for deployments. Astron-xmod-shim uses a **Builder pattern** for GoalSet construction.

#### Built-in Example: OpenSourceLLM GoalSet

This built-in GoalSet handles open-source LLM deployment, including model path mapping, deployment validation, spec consistency checks, and service exposure.

##### Step 1: Define Goals

```go
package mygoalset

import (
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/pkg/log"
)

var validateModel = goal.Goal{
	Name: "validate-model",
	IsAchieved: func(ctx *goal.Context) bool {
		return ctx.DeploySpec.ModelName != ""
	},
	Ensure: func(ctx *goal.Context) error {
		log.Info("Validating model: %s", ctx.DeploySpec.ModelName)
		return nil
	},
}

var prepareResources = goal.Goal{
	Name: "prepare-resources",
	IsAchieved: func(ctx *goal.Context) bool {
		resourceReady, exists := ctx.Get("resourceReady").(bool)
		return exists && resourceReady
	},
	Ensure: func(ctx *goal.Context) error {
		log.Info("Preparing deployment resources")
		ctx.Set("resourceReady", true)
		return nil
	},
}
```

##### Step 2: Create and Register GoalSet

```go
package mygoalset

import (
	"astron-xmod-shim/internal/core/goal"
	"time"
)

func init() {
	newMyGoalSet()
}

func newMyGoalSet() {
	goal.NewGoalSetBuilder("my-goalset").
		AddGoal(validateModel).
		AddGoal(prepareResources).
		WithMaxRetries(3).
		WithTimeout(2 * time.Minute).
		BuildAndRegister()
}
```

### Extension Examples

- **Multimodal Model GoalSet**: Add validation for text+image inputs, optimize GPU allocation, set inference parameters.
- **Edge Deployment GoalSet**: Enforce resource limits, apply model quantization, enable offline inference.
- **Enterprise Security GoalSet**: Integrate auth, encrypted transport, and access control.

### Plugin Integration

Astron-xmod-shim uses Go’s `init()` registration mechanism—**not** dynamic loading or shared libraries.

#### Built-in Plugins

Registered automatically in `init()`:
```go
func init() {
	shimlet.Registry.AutoRegister(&K8sShimlet{})
}
```

#### Custom Plugins

1. Implement the standard interface (`Shimlet` or `GoalSet`)
2. Auto-register in `init()`
3. Recompile the binary with your plugin code included

#### Plugin Selection & Configuration

Specify plugins via CLI or config:

```bash
./model-serve-shim --shimlet=k8s --goalset=opensource-llm-deploy
```

Or in `config.yaml`:
```yaml
plugins:
  defaultShimlet: k8s
  defaultGoalSet: opensource-llm-deploy
```

## Configuration

### Command-Line Flags

```bash
./model-serve-shim --help

Usage:
  --port int              HTTP server port (default: 8080)
  --config string         Path to config file
  --shimlet string        Default shimlet plugin
  --goalset string        Default goalset plugin
  --plugin-dir string     Plugin directory (unused in static build)
  --log-level string      Log level: debug, info, warn, error (default: "info")
```

### Config File (YAML)

```yaml
# config.yaml
service:
  port: 8080
  readTimeout: 30s
  writeTimeout: 30s

plugins:
  defaultShimlet: k8s
  defaultGoalSet: opensource-llm-deploy

logging:
  level: info
  format: text
  output: stdout
```

## Contributing

We welcome community contributions! Please follow these steps:

1. Fork the repo and create a feature branch
2. Adhere to coding standards (use `pre-commit` hooks)
3. Ensure all tests pass before submitting
4. Open a PR with a clear description of changes and issues resolved

## 🌟 Star History

<div align="center">
  <img src="https://api.star-history.com/svg?repos=iflytek/Astron-xmod-shim&type=Date" alt="Star History Chart" width="600">
</div>

## License

Astron-xmod-shim is licensed under the **Apache License 2.0**.

## Contact Us

For questions or feedback:

- GitHub Issues: https://github.com/iflytek/Astron-xmod-shim/issues
- Email: hxli28@iflytek.com
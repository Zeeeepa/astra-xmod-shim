package dto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGlobalConfig 测试 GlobalConfig 结构体的基本功能
func TestGlobalConfig(t *testing.T) {
	// 准备测试数据
	globalConfig := &GlobalConfig{
		K8s: K8sConfig{
			Kubeconfig: "/path/to/kubeconfig",
			Context:    "test-context",
			QPS:        10.0,
			Burst:      20,
			Timeout:    30,
		},
		Server: Server{
			Port: "8080",
		},
		Log: LogConfig{
			Level:         "info",
			Path:          "/var/log/app.log",
			MaxSize:       100,
			MaxAge:        7,
			Compress:      true,
			ShowLine:      true,
			EnableConsole: true,
		},
		CurrentShimlet: "kubernetes",
		Shimlets: map[string]ShimletConfig{
			"kubernetes": {
				ConfigPath: "/etc/kubernetes/config.yaml",
			},
		},
		ModelManage: ModelManageConfig{
			ModelRoot: "/models",
		},
	}

	// 验证结构体字段值
	assert.Equal(t, "/path/to/kubeconfig", globalConfig.K8s.Kubeconfig)
	assert.Equal(t, "test-context", globalConfig.K8s.Context)
	assert.Equal(t, float32(10.0), globalConfig.K8s.QPS)
	assert.Equal(t, 20, globalConfig.K8s.Burst)
	assert.Equal(t, int64(30), globalConfig.K8s.Timeout)
	assert.Equal(t, "8080", globalConfig.Server.Port)
	assert.Equal(t, "info", globalConfig.Log.Level)
	assert.Equal(t, "/var/log/app.log", globalConfig.Log.Path)
	assert.Equal(t, 100, globalConfig.Log.MaxSize)
	assert.Equal(t, 7, globalConfig.Log.MaxAge)
	assert.Equal(t, true, globalConfig.Log.Compress)
	assert.Equal(t, true, globalConfig.Log.ShowLine)
	assert.Equal(t, true, globalConfig.Log.EnableConsole)
	assert.Equal(t, "kubernetes", globalConfig.CurrentShimlet)
	assert.Equal(t, "/etc/kubernetes/config.yaml", globalConfig.Shimlets["kubernetes"].ConfigPath)
	assert.Equal(t, "/models", globalConfig.ModelManage.ModelRoot)

	// 测试 JSON 序列化和反序列化
	jsonData, err := json.Marshal(globalConfig)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	var unmarshaled GlobalConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, globalConfig.K8s.Kubeconfig, unmarshaled.K8s.Kubeconfig)
	assert.Equal(t, globalConfig.Server.Port, unmarshaled.Server.Port)
}

// TestK8sConfig 测试 K8sConfig 结构体的基本功能
func TestK8sConfig(t *testing.T) {
	k8sConfig := &K8sConfig{
		Kubeconfig: "/another/path/to/kubeconfig",
		Context:    "another-context",
		QPS:        5.0,
		Burst:      10,
		Timeout:    15,
	}

	assert.Equal(t, "/another/path/to/kubeconfig", k8sConfig.Kubeconfig)
	assert.Equal(t, "another-context", k8sConfig.Context)
	assert.Equal(t, float32(5.0), k8sConfig.QPS)
	assert.Equal(t, 10, k8sConfig.Burst)
	assert.Equal(t, int64(15), k8sConfig.Timeout)
}

// TestServerConfig 测试 Server 结构体的基本功能
func TestServerConfig(t *testing.T) {
	server := &Server{
		Port: "9090",
	}

	assert.Equal(t, "9090", server.Port)
}

// TestLogConfig 测试 LogConfig 结构体的基本功能
func TestLogConfig(t *testing.T) {
	logConfig := &LogConfig{
		Level:         "debug",
		Path:          "/var/log/test.log",
		MaxSize:       50,
		MaxAge:        3,
		Compress:      false,
		ShowLine:      false,
		EnableConsole: false,
	}

	assert.Equal(t, "debug", logConfig.Level)
	assert.Equal(t, "/var/log/test.log", logConfig.Path)
	assert.Equal(t, 50, logConfig.MaxSize)
	assert.Equal(t, 3, logConfig.MaxAge)
	assert.Equal(t, false, logConfig.Compress)
	assert.Equal(t, false, logConfig.ShowLine)
	assert.Equal(t, false, logConfig.EnableConsole)
}

// TestShimletConfig 测试 ShimletConfig 结构体的基本功能
func TestShimletConfig(t *testing.T) {
	shimletConfig := &ShimletConfig{
		ConfigPath: "/etc/shimlet/config.yaml",
	}

	assert.Equal(t, "/etc/shimlet/config.yaml", shimletConfig.ConfigPath)
}

// TestModelManageConfig 测试 ModelManageConfig 结构体的基本功能
func TestModelManageConfig(t *testing.T) {
	modelManageConfig := &ModelManageConfig{
		ModelRoot: "/test/models",
	}

	assert.Equal(t, "/test/models", modelManageConfig.ModelRoot)
}

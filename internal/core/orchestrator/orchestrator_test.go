package orchestrator

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/internal/core/shimlet"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/typereg"
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShimlet 是shimlet.Shimlet接口的简化mock实现
type MockShimlet struct {
	mock.Mock
}

func (m *MockShimlet) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockShimlet) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockShimlet) InitWithConfig(confPath string) error {
	args := m.Called(confPath)
	return args.Error(0)
}

func (m *MockShimlet) Apply(spec *dto.RequirementSpec) error {
	args := m.Called(spec)
	return args.Error(0)
}

func (m *MockShimlet) Delete(resourceId string) error {
	args := m.Called(resourceId)
	return args.Error(0)
}

func (m *MockShimlet) Status(resourceId string) (*dto.RuntimeStatus, error) {
	args := m.Called(resourceId)
	status, _ := args.Get(0).(*dto.RuntimeStatus)
	return status, args.Error(1)
}

func (m *MockShimlet) ListDeployedServices() ([]string, error) {
	args := m.Called()
	services, _ := args.Get(0).([]string)
	return services, args.Error(1)
}

// 简化的测试用例
func TestOrchestrator(t *testing.T) {
	// 创建临时配置文件
	configContent := `
current-shimlet: "k8s"
shimlets:
  k8s:
    config-path: "/conf/k8s-shimlet.yaml"
`
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	assert.NoError(t, err)
	assert.NoError(t, tmpFile.Close())

	// 设置配置文件路径
	config.SetConfigPath(tmpFile.Name())

	// 测试Provision方法
	t.Run("Provision", func(t *testing.T) {
		// 创建测试需要的组件
		mockQueue := workqueue.New()
		defer mockQueue.ShutDown()

		mockSpecStore := spec.NewMemoryStore()
		shimReg := typereg.New[shimlet.Shimlet]()

		// 创建并配置mock shimlet
		mockShim := new(MockShimlet)
		mockShim.On("ID").Return("k8s")
		mockShim.On("InitWithConfig", "/conf/k8s-shimlet.yaml").Return(nil)

		// 注册mock shimlet到注册中心
		shimReg.AutoRegister(mockShim)

		// 创建orchestrator实例，使用固定的shimlet ID "k8s"
		orchestrator := NewOrchestratorWithShimletGetter(
			shimReg,
			map[string]*goal.GoalSet{},
			mockQueue,
			mockSpecStore,
			func() string {
				return "k8s"
			},
		)

		serviceSpec := &dto.RequirementSpec{
			ServiceId:            "test-service",
			GoalSetName:          "test-goalset",
			ResourceRequirements: &dto.ResourceRequirements{},
		}

		err := orchestrator.Provision(serviceSpec)
		assert.NoError(t, err)

		// 验证规格被保存
		savedSpec := mockSpecStore.Get("test-service")
		assert.NotNil(t, savedSpec)

		// 验证项目被添加到队列
		assert.Equal(t, 1, mockQueue.Len())

		// 清理队列
		if mockQueue.Len() > 0 {
			_, done := mockQueue.Get()
			done()
		}
	})

}

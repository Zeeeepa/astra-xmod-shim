package goal

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShimlet 是一个模拟的 shimlet 实现
type MockShimlet struct {
	mock.Mock
}

func (m *MockShimlet) InitWithConfig(confPath string) error {
	args := m.Called(confPath)
	return args.Error(0)
}

func (m *MockShimlet) Apply(spec interface{}) error {
	args := m.Called(spec)
	return args.Error(0)
}

func (m *MockShimlet) Delete(resourceId string) error {
	args := m.Called(resourceId)
	return args.Error(0)
}

func (m *MockShimlet) Status(resourceId string) (interface{}, error) {
	args := m.Called(resourceId)
	return args.Get(0), args.Error(1)
}

func (m *MockShimlet) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockShimlet) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockShimlet) ListDeployedServices() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func TestGoalSetBuilder(t *testing.T) {
	// 创建一个GoalSetBuilder
	builder := NewGoalSetBuilder("test-goal-set")

	// 创建测试用的Goal
	testGoal := Goal{
		Name: "test-goal",
		IsAchieved: func(ctx *Context) bool {
			return true
		},
		Ensure: func(ctx *Context) error {
			return nil
		},
	}

	// 测试添加Goal
	builder.AddGoal(testGoal)

	// 测试设置重试次数
	builder.WithMaxRetries(5)

	// 测试设置超时时间
	builder.WithTimeout(30 * time.Second)

	// 构建并注册
	builder.BuildAndRegister()

	// 验证注册成功
	registeredGoalSet, exists := Registry["test-goal-set"]
	assert.True(t, exists, "GoalSet should be registered")
	assert.Equal(t, "test-goal-set", registeredGoalSet.Name)
	assert.Equal(t, 1, len(registeredGoalSet.Goals))
	assert.Equal(t, 5, registeredGoalSet.MaxRetries)
	assert.Equal(t, 30*time.Second, registeredGoalSet.Timeout)
}

func TestGoalSetBuilderChaining(t *testing.T) {
	// 测试链式调用
	NewGoalSetBuilder("chaining-test").
		AddGoal(Goal{
			Name: "goal1",
			IsAchieved: func(ctx *Context) bool { return true },
			Ensure:     func(ctx *Context) error { return nil },
		}).
		AddGoal(Goal{
			Name: "goal2",
			IsAchieved: func(ctx *Context) bool { return false },
			Ensure:     func(ctx *Context) error { return errors.New("not achieved") },
		}).
		WithMaxRetries(3).
		WithTimeout(1 * time.Minute).
		BuildAndRegister()

	// 验证链式调用的结果
	registeredGoalSet, exists := Registry["chaining-test"]
	assert.True(t, exists)
	assert.Equal(t, 2, len(registeredGoalSet.Goals))
	assert.Equal(t, 3, registeredGoalSet.MaxRetries)
	assert.Equal(t, 1*time.Minute, registeredGoalSet.Timeout)
}

func TestGoalExecution(t *testing.T) {
	// 创建测试用的Goal
	achievedGoal := Goal{
		Name: "achieved-goal",
		IsAchieved: func(ctx *Context) bool {
			return true // 已达成
		},
		Ensure: func(ctx *Context) error {
			return nil
		},
	}

	notAchievedGoal := Goal{
		Name: "not-achieved-goal",
		IsAchieved: func(ctx *Context) bool {
			return false // 未达成
		},
		Ensure: func(ctx *Context) error {
			return errors.New("goal not achieved")
		},
	}

	successGoal := Goal{
		Name: "success-goal",
		IsAchieved: func(ctx *Context) bool {
			return false // 未达成，但Ensure会成功
		},
		Ensure: func(ctx *Context) error {
			return nil // Ensure成功
		},
	}

	// 测试已达成的Goal
	ctx := NewContext()
	assert.True(t, achievedGoal.IsAchieved(ctx))

	// 测试未达成且Ensure失败的Goal
	err := notAchievedGoal.Ensure(ctx)
	assert.Error(t, err)
	assert.Equal(t, "goal not achieved", err.Error())

	// 测试未达成但Ensure成功的Goal
	err = successGoal.Ensure(ctx)
	assert.NoError(t, err)
}

func TestContextOperations(t *testing.T) {
	// 创建Context
	ctx := NewContext()

	// 测试Set和Get操作
	ctx.Set("key1", "value1")
	ctx.Set("key2", 42)
	ctx.Set("key3", true)

	// 测试获取字符串值
	assert.Equal(t, "value1", ctx.GetString("key1"))
	assert.Equal(t, "", ctx.GetString("key2")) // 不是字符串类型
	assert.Equal(t, "", ctx.GetString("non-existent")) // 不存在的键

	// 测试通用Get操作
	assert.Equal(t, "value1", ctx.Get("key1"))
	assert.Equal(t, 42, ctx.Get("key2"))
	assert.Equal(t, true, ctx.Get("key3"))
	assert.Nil(t, ctx.Get("non-existent")) // 不存在的键
}

// 测试Context的默认值
func TestContextDefaultValues(t *testing.T) {
	ctx := NewContext()
	
	// 测试空Context的GetString
	assert.Equal(t, "", ctx.GetString("non-existent"))
	
	// 测试空Context的Get
	assert.Nil(t, ctx.Get("non-existent"))
}

// 测试GoalSetBuilder的默认值
func TestGoalSetBuilderDefaults(t *testing.T) {
	builder := NewGoalSetBuilder("default-test")
	
	// 验证默认值
	// 通过反射获取私有字段值比较困难，所以我们通过构建后的结果来验证
	builder.BuildAndRegister()
	
	registeredGoalSet, exists := Registry["default-test"]
	assert.True(t, exists)
	assert.Equal(t, 0, registeredGoalSet.MaxRetries) // 默认不重试
	assert.Equal(t, 10*time.Second, registeredGoalSet.Timeout) // 默认10秒超时
}

// 测试多次注册同一个名称的GoalSet
func TestGoalSetBuilderOverwrite(t *testing.T) {
	// 第一次注册
	NewGoalSetBuilder("overwrite-test").
		AddGoal(Goal{
			Name: "goal1",
			IsAchieved: func(ctx *Context) bool { return true },
			Ensure:     func(ctx *Context) error { return nil },
		}).
		BuildAndRegister()
	
	// 验证第一次注册
	registeredGoalSet, exists := Registry["overwrite-test"]
	assert.True(t, exists)
	assert.Equal(t, 1, len(registeredGoalSet.Goals))
	
	// 第二次注册同名GoalSet
	NewGoalSetBuilder("overwrite-test").
		AddGoal(Goal{
			Name: "goal2",
			IsAchieved: func(ctx *Context) bool { return false },
			Ensure:     func(ctx *Context) error { return nil },
		}).
		AddGoal(Goal{
			Name: "goal3",
			IsAchieved: func(ctx *Context) bool { return true },
			Ensure:     func(ctx *Context) error { return nil },
		}).
		BuildAndRegister()
	
	// 验证第二次注册覆盖了第一次
	registeredGoalSet, exists = Registry["overwrite-test"]
	assert.True(t, exists)
	assert.Equal(t, 2, len(registeredGoalSet.Goals))
	assert.Equal(t, "goal2", registeredGoalSet.Goals[0].Name)
	assert.Equal(t, "goal3", registeredGoalSet.Goals[1].Name)
}

// 测试GoalSetBuilder的各个方法
func TestGoalSetBuilderMethods(t *testing.T) {
	builder := NewGoalSetBuilder("method-test")
	
	// 测试AddGoal返回builder本身（链式调用）
	result := builder.AddGoal(Goal{
		Name: "test-goal",
		IsAchieved: func(ctx *Context) bool { return true },
		Ensure:     func(ctx *Context) error { return nil },
	})
	assert.Equal(t, builder, result)
	
	// 测试WithMaxRetries返回builder本身（链式调用）
	result = builder.WithMaxRetries(5)
	assert.Equal(t, builder, result)
	
	// 测试WithTimeout返回builder本身（链式调用）
	result = builder.WithTimeout(30 * time.Second)
	assert.Equal(t, builder, result)
}

// 测试Goal的执行
func TestGoalIsAchievedAndEnsure(t *testing.T) {
	// 创建一个始终未达成的Goal
	goal := Goal{
		Name: "test-goal",
		IsAchieved: func(ctx *Context) bool {
			return false
		},
		Ensure: func(ctx *Context) error {
			// 在Ensure中设置一些数据来验证执行
			ctx.Set("executed", true)
			return nil
		},
	}
	
	ctx := NewContext()
	
	// 验证Goal未达成
	assert.False(t, goal.IsAchieved(ctx))
	
	// 执行Ensure
	err := goal.Ensure(ctx)
	assert.NoError(t, err)
	
	// 验证Ensure被执行（通过检查设置的数据）
	assert.Equal(t, true, ctx.Get("executed"))
}

// 测试Goal执行失败的情况
func TestGoalEnsureFailure(t *testing.T) {
	// 创建一个Ensure会失败的Goal
	goal := Goal{
		Name: "failing-goal",
		IsAchieved: func(ctx *Context) bool {
			return false
		},
		Ensure: func(ctx *Context) error {
			return errors.New("ensure failed")
		},
	}
	
	ctx := NewContext()
	
	// 执行Ensure应该返回错误
	err := goal.Ensure(ctx)
	assert.Error(t, err)
	assert.Equal(t, "ensure failed", err.Error())
}
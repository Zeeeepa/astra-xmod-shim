#### 步骤 1：定义 Goal

```go
package mygoalset

import (
	"astron-xmod-shim/internal/core/goal"
	"astron-xmod-shim/pkg/log"
)

// 定义一个验证模型的 Goal
var validateModel = goal.Goal{
	Name: "validate-model",
	IsAchieved: func(ctx *goal.Context) bool {
		// 检查目标是否已达成
		return ctx.DeploySpec.ModelName != ""
	},
	Ensure: func(ctx *goal.Context) error {
		log.Info("开始验证模型: %s", ctx.DeploySpec.ModelName)
		// 实现模型验证逻辑
		return nil
	},
}

// 定义一个准备资源的 Goal
var prepareResources = goal.Goal{
	Name: "prepare-resources",
	IsAchieved: func(ctx *goal.Context) bool {
		// 检查资源是否已准备好
		// 这里可以检查模型文件、配置等是否存在
		resourceReady, exists := ctx.Get("resourceReady").(bool)
		return exists && resourceReady
	},
	Ensure: func(ctx *goal.Context) error {
		log.Info("准备部署资源")
		// 实现资源准备逻辑
		// 准备完成后设置标记
		ctx.Set("resourceReady", true)
		return nil
	},
}
```

#### 步骤 2：创建并注册 GoalSet

```go
package mygoalset

import (
	"astron-xmod-shim/internal/core/goal"
	"time"
)

// init 函数在插件加载时自动调用
func init() {
	// 使用 Builder 模式创建并注册自定义 GoalSet
	newMyGoalSet()
}

// newMyGoalSet 创建自定义 GoalSet 实例
func newMyGoalSet() {
	goal.NewGoalSetBuilder("my-goalset").
		AddGoal(validateModel).
		AddGoal(prepareResources).
		WithMaxRetries(3).           // 失败最多重试 3 次
		WithTimeout(2 * time.Minute). // 整体超时 2 分钟
		BuildAndRegister()
}
```
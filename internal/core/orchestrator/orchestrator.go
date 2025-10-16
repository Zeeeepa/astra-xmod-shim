package orchestrator

import (
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/internal/core/goal"
	_ "astron-xmod-shim/internal/core/goal/goalset"
	"astron-xmod-shim/internal/core/shimlet"
	"astron-xmod-shim/internal/core/spec"
	"astron-xmod-shim/internal/core/typereg"
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
	"fmt"
)

// CurrentShimletGetter 定义获取当前shimlet ID的函数类型
type CurrentShimletGetter func() string

type Orchestrator struct {
	shimReg              *typereg.TypeReg[shimlet.Shimlet]
	goalSetReg           map[string]*goal.GoalSet
	specStore            spec.Store
	queue                *workqueue.Queue
	currentShimletGetter CurrentShimletGetter
}

func NewOrchestrator(
	shimReg *typereg.TypeReg[shimlet.Shimlet],
	pipeReg map[string]*goal.GoalSet,
	queue *workqueue.Queue,
	specStore spec.Store,
) *Orchestrator {
	return &Orchestrator{
		queue:      queue,
		shimReg:    shimReg,
		goalSetReg: pipeReg,
		specStore:  specStore,
		currentShimletGetter: func() string {
			return config.Get().CurrentShimlet
		},
	}
}

// NewOrchestratorWithShimletGetter 创建一个可以自定义currentShimletGetter的Orchestrator实例
// 主要用于测试
func NewOrchestratorWithShimletGetter(
	shimReg *typereg.TypeReg[shimlet.Shimlet],
	pipeReg map[string]*goal.GoalSet,
	queue *workqueue.Queue,
	specStore spec.Store,
	currentShimletGetter CurrentShimletGetter,
) *Orchestrator {
	return &Orchestrator{
		queue:                queue,
		shimReg:              shimReg,
		goalSetReg:           pipeReg,
		specStore:            specStore,
		currentShimletGetter: currentShimletGetter,
	}
}

var GlobalOrchestrator *Orchestrator

func (o *Orchestrator) Provision(spec *dto.RequirementSpec) error {

	// 覆盖掉 nvidia.com/gpu 的 limit
	spec.ResourceRequirements.AcceleratorType = "nvidia.com/gpu"

	// goalset 已在api handler 层 确定
	// shimlet 已在启动时配置全局确定

	// RequirementSpec 持久化 部署期望
	spec.ReplicaCount = 1
	spec.ShimletName = o.currentShimletGetter()
	// 如果这里是更新, 则需要 对应goalset reconcile 检测到 不一致 并调用ensure 闭环
	o.specStore.Set(spec.ServiceId, spec)

	// 投递到队列
	o.queue.Add(spec.ServiceId)

	return nil
}

//// DeleteService 删除指定的模型服务
//func (o *Orchestrator) DeleteService(serviceID string) error {
//	// 获取当前使用的shimlet
//	currentShimletId := o.currentShimletGetter()
//
//	runtimeShimlet, err := o.shimReg.GetSingleton(currentShimletId)
//	if err != nil {
//		log.Error("get runtime shimlet error", err)
//		return err
//	}
//
//	// 调用shimlet的Delete方法删除资源
//	if err := runtimeShimlet.Delete(serviceID); err != nil {
//		log.Error("delete service failed", err)
//		return err
//	}
//	o.specStore.Delete(serviceID)
//
//	go log.Info("service deleted successfully", "serviceID", serviceID)
//	return nil
//}

// GetServiceStatus 获取指定服务的状态信息
func (o *Orchestrator) GetServiceStatus(serviceID string) (*dto.RuntimeStatus, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("serviceID is required")
	}

	currentShimletId := o.currentShimletGetter()
	runtimeShimlet, err := o.shimReg.GetSingleton(currentShimletId)
	if err != nil {
		return nil, err
	}
	status, err := runtimeShimlet.Status(serviceID)
	if err != nil {
		return nil, err
	}
	if status.EndPoint != "" {
		status.EndPoint += "/v1/chat/completions"
	}
	return status, nil
}

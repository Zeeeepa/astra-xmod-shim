package spec

import (
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
)

// Store 是 StateManager 的简单内存实现
// 修改 ReloadAll 方法签名，接受 workqueue 作为参数

type Store interface {
	Set(serviceID string, spec *dto.RequirementSpec)
	Delete(serviceID string)
	Get(serviceID string) *dto.RequirementSpec
	ReloadAll(queue *workqueue.Queue)
}

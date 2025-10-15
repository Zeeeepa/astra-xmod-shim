package spec

import (
	"astron-xmod-shim/internal/core/workqueue"
	dto "astron-xmod-shim/internal/dto/deploy"
)

// MemoryStore 是 Store 的简单内存实现
type MemoryStore struct {
	specMap map[string]*dto.RequirementSpec
}

// NewMemoryStore 创建一个新的 StateManager 实例
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		specMap: make(map[string]*dto.RequirementSpec),
	}
}

// Set 保存用户部署期望 以及 runtime shimlet 和 部署 goal set (目标集合)
func (m *MemoryStore) Set(serviceID string, spec *dto.RequirementSpec) {
	m.specMap[serviceID] = spec
}

func (m *MemoryStore) Get(serviceID string) *dto.RequirementSpec {
	return m.specMap[serviceID]
}

// Delete 删除服务的状态记录
func (m *MemoryStore) Delete(serviceID string) {
	delete(m.specMap, serviceID)
}

// ReloadAll 实现Store接口的ReloadAll方法
// 在MemoryStore实现中，此方法实际不需要任何操作
func (m *MemoryStore) ReloadAll(queue *workqueue.Queue) {
	// MemoryStore实现中无需特殊处理
	// 如果需要，这里也可以添加清空队列和重新投递的逻辑
}

func (m *MemoryStore) GetStatus(id string) {

	// TODO 判断goal set 所有 goals is achieved
}

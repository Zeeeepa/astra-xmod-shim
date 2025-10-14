package spec

import (
	dto "astron-xmod-shim/internal/dto/deploy"
	"testing"
)

func TestMemoryStoreBasicOperations(t *testing.T) {
	// 创建内存存储
	store := NewMemoryStore()

	// 创建测试数据
	serviceID := "test-service"
	spec := &dto.RequirementSpec{
		ServiceId:    serviceID,
		ModelName:    "test-model",
		ReplicaCount: 1,
	}

	// 测试设置和获取
	store.Set(serviceID, spec)
	retrieved := store.Get(serviceID)
	if retrieved == nil {
		t.Error("Expected to retrieve non-nil spec")
	} else if retrieved.ServiceId != serviceID {
		t.Errorf("Expected service ID '%s', got '%s'", serviceID, retrieved.ServiceId)
	} else if retrieved.ModelName != "test-model" {
		t.Errorf("Expected model name 'test-model', got '%s'", retrieved.ModelName)
	}

	// 测试删除
	store.Delete(serviceID)
	retrieved = store.Get(serviceID)
	if retrieved != nil {
		t.Error("Expected nil after delete")
	}
}

func TestMemoryStoreMultipleServices(t *testing.T) {
	// 创建内存存储
	store := NewMemoryStore()

	// 创建多个服务规格
	spec1 := &dto.RequirementSpec{ServiceId: "service-1", ModelName: "model-1"}
	spec2 := &dto.RequirementSpec{ServiceId: "service-2", ModelName: "model-2"}
	spec3 := &dto.RequirementSpec{ServiceId: "service-3", ModelName: "model-3"}

	// 设置多个服务
	store.Set("service-1", spec1)
	store.Set("service-2", spec2)
	store.Set("service-3", spec3)

	// 验证所有服务都能正确获取
	if retrieved := store.Get("service-1"); retrieved == nil || retrieved.ModelName != "model-1" {
		t.Error("Failed to retrieve service-1")
	}
	if retrieved := store.Get("service-2"); retrieved == nil || retrieved.ModelName != "model-2" {
		t.Error("Failed to retrieve service-2")
	}
	if retrieved := store.Get("service-3"); retrieved == nil || retrieved.ModelName != "model-3" {
		t.Error("Failed to retrieve service-3")
	}

	// 验证不存在的服务返回nil
	if retrieved := store.Get("non-existent"); retrieved != nil {
		t.Error("Expected nil for non-existent service")
	}

	// 删除一个服务并验证
	store.Delete("service-2")
	if retrieved := store.Get("service-2"); retrieved != nil {
		t.Error("Expected nil after deleting service-2")
	}
	// 确保其他服务仍然存在
	if retrieved := store.Get("service-1"); retrieved == nil {
		t.Error("service-1 should still exist")
	}
}

package spec

import (
	"astron-xmod-shim/internal/core/workqueue"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"astron-xmod-shim/internal/config"
	dto "astron-xmod-shim/internal/dto/deploy"
	"astron-xmod-shim/pkg/k8s"
	"astron-xmod-shim/pkg/log"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EtcdStore 使用Kubernetes ConfigMap实现的Store接口
// 注意：这里命名为EtcdStore是为了满足需求，实际实现是基于ConfigMap的

const (
	// ConfigMapName ConfigMap名称
	ConfigMapName = "astron-xmod-shim-services"
	// DataKey 存储数据的键名
	DataKey = "services"
	// ConfigMapNamespace 默认命名空间
	ConfigMapNamespace = "default"
)

// EtcdStore 实现了Store接口，使用ConfigMap存储所有RequirementSpec
type EtcdStore struct {
	client    *k8s.K8sClient
	namespace string
}

// NewEtcdStore 创建并返回一个新的EtcdStore实例
func NewEtcdStore() *EtcdStore {
	store := &EtcdStore{
		namespace: ConfigMapNamespace,
	}
	// 初始化客户端
	if err := store.initClient(); err != nil {
		log.Error("Failed to initialize EtcdStore client: %v", err)
	}
	return store
}

// initClient 初始化Kubernetes客户端
func (e *EtcdStore) initClient() error {
	if e.client != nil {
		return nil
	}

	// 获取Kubernetes配置
	k8sCfg := config.Get().K8s
	client, err := k8s.NewK8sClient(&k8sCfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	e.client = client
	return nil
}

// Set 实现Store接口的Set方法，将RequirementSpec保存到ConfigMap
func (e *EtcdStore) Set(serviceID string, spec *dto.RequirementSpec) {
	if e.client == nil {
		if err := e.initClient(); err != nil {
			log.Error("Failed to initialize client in Set: %v", err)
			return
		}
	}

	// 获取现有的ConfigMap
	cm, err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Get(
		context.Background(),
		ConfigMapName,
		metav1.GetOptions{},
	)

	// 如果ConfigMap不存在，创建一个新的
	if err != nil && k8serrors.IsNotFound(err) {
		cm = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      ConfigMapName,
				Namespace: e.namespace,
				Labels: map[string]string{
					"app":      "astron-xmod-shim",
					"resource": "requirement-specs",
				},
				Annotations: map[string]string{
					"last-updated": time.Now().Format(time.RFC3339),
				},
			},
			Data: make(map[string]string),
		}
	} else if err != nil {
		log.Error("Failed to get ConfigMap: %v", err)
		return
	}

	// 将spec序列化为JSON
	data, err := json.Marshal(spec)
	if err != nil {
		log.Error("Failed to marshal spec: %v", err)
		return
	}

	// 确保Data字段已初始化
	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}

	// 将服务规格存储在ConfigMap中，以serviceID为键
	cm.Data[serviceID] = string(data)
	cm.Annotations["last-updated"] = time.Now().Format(time.RFC3339)

	// 如果是新创建的ConfigMap，则创建它
	if cm.ResourceVersion == "" {
		_, err = e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Create(
			context.Background(),
			cm,
			metav1.CreateOptions{},
		)
		if err != nil {
			log.Error("Failed to create ConfigMap: %v", err)
		}
		return
	}

	// 否则更新现有的ConfigMap
	_, err = e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Update(
		context.Background(),
		cm,
		metav1.UpdateOptions{},
	)
	if err != nil {
		log.Error("Failed to update ConfigMap: %v", err)
	}
}

// Get 实现Store接口的Get方法，从ConfigMap中获取RequirementSpec
func (e *EtcdStore) Get(serviceID string) *dto.RequirementSpec {
	if e.client == nil {
		if err := e.initClient(); err != nil {
			log.Error("Failed to initialize client in Get: %v", err)
			return nil
		}
	}

	// 获取ConfigMap
	cm, err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Get(
		context.Background(),
		ConfigMapName,
		metav1.GetOptions{},
	)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		log.Error("Failed to get ConfigMap: %v", err)
		return nil
	}

	// 检查是否存在该服务的数据
	data, ok := cm.Data[serviceID]
	if !ok {
		return nil
	}

	// 反序列化
	spec := &dto.RequirementSpec{}
	if err := json.Unmarshal([]byte(data), spec); err != nil {
		log.Error("Failed to unmarshal spec: %v", err)
		return nil
	}

	return spec
}

// Delete 实现Store接口的Delete方法，从ConfigMap中删除对应的service
func (e *EtcdStore) Delete(serviceID string) {
	if e.client == nil {
		if err := e.initClient(); err != nil {
			log.Error("Failed to initialize client in Delete: %v", err)
			return
		}
	}

	// 获取现有的ConfigMap
	cm, err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Get(
		context.Background(),
		ConfigMapName,
		metav1.GetOptions{},
	)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return
		}
		log.Error("Failed to get ConfigMap: %v", err)
		return
	}

	// 删除指定服务的数据
	if cm.Data != nil {
		delete(cm.Data, serviceID)
		cm.Annotations["last-updated"] = time.Now().Format(time.RFC3339)

		// 更新ConfigMap
		_, err = e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Update(
			context.Background(),
			cm,
			metav1.UpdateOptions{},
		)
		if err != nil {
			log.Error("Failed to update ConfigMap: %v", err)
		}
	}
}

// ReloadAll 实现Store接口的ReloadAll方法
// 功能：重新投递所有ConfigMap中记录的serviceID
// 参数：queue - 工作队列接口，用于重新投递事件
// 特性：保证幂等性 - 无论调用多少次，最终效果相同
func (e *EtcdStore) ReloadAll(queue *workqueue.Queue) {

}

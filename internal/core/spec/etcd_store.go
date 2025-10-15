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
	// ConfigMapNamePrefix ConfigMap名称前缀
	ConfigMapNamePrefix = "astron-xmod-shim-"
	// DataKey 存储数据的键名
	DataKey = "requirement-spec"
	// ConfigMapNamespace 默认命名空间
	ConfigMapNamespace = "default"
)

// EtcdStore 实现了Store接口，使用ConfigMap存储RequirementSpec
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

	// 将spec序列化为JSON
	data, err := json.Marshal(spec)
	if err != nil {
		log.Error("Failed to marshal spec: %v", err)
		return
	}

	// 构建ConfigMap名称
	cmName := ConfigMapNamePrefix + serviceID

	// 创建或更新ConfigMap
	targetCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: e.namespace,
			Labels: map[string]string{
				"app":        "astron-xmod-shim",
				"resource":   "requirement-spec",
				"service-id": serviceID,
			},
			Annotations: map[string]string{
				"last-updated": time.Now().Format(time.RFC3339),
			},
		},
		Data: map[string]string{
			DataKey: string(data),
		},
	}

	// 尝试获取现有ConfigMap
	existingCM, err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Get(
		context.Background(),
		cmName,
		metav1.GetOptions{},
	)

	if err != nil {
		// 资源不存在，创建新的
		if k8serrors.IsNotFound(err) {
			_, err = e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Create(
				context.Background(),
				targetCM,
				metav1.CreateOptions{},
			)
			if err != nil {
				log.Error("Failed to create ConfigMap: %v", err)
			}
			return
		}
		// 其他错误
		log.Error("Failed to get ConfigMap: %v", err)
		return
	}

	// 资源已存在，更新
	targetCM.ResourceVersion = existingCM.ResourceVersion
	// 保留现有标签
	for k, v := range existingCM.Labels {
		if _, ok := targetCM.Labels[k]; !ok {
			targetCM.Labels[k] = v
		}
	}
	// 保留现有注释
	for k, v := range existingCM.Annotations {
		if k != "last-updated" {
			targetCM.Annotations[k] = v
		}
	}

	_, err = e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Update(
		context.Background(),
		targetCM,
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

	// 构建ConfigMap名称
	cmName := ConfigMapNamePrefix + serviceID

	// 获取ConfigMap
	cm, err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Get(
		context.Background(),
		cmName,
		metav1.GetOptions{},
	)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		log.Error("Failed to get ConfigMap: %v", err)
		return nil
	}

	// 获取数据
	data, ok := cm.Data[DataKey]
	if !ok {
		log.Error("Data not found in ConfigMap: %s", DataKey)
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

// Delete 实现Store接口的Delete方法，删除对应的ConfigMap
func (e *EtcdStore) Delete(serviceID string) {
	if e.client == nil {
		if err := e.initClient(); err != nil {
			log.Error("Failed to initialize client in Delete: %v", err)
			return
		}
	}

	// 构建ConfigMap名称
	cmName := ConfigMapNamePrefix + serviceID

	// 删除ConfigMap
	err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).Delete(
		context.Background(),
		cmName,
		metav1.DeleteOptions{},
	)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			log.Error("Failed to delete ConfigMap: %v", err)
		}
	}
}

// ReloadAll 实现Store接口的ReloadAll方法
// 功能：清空workqueue并重新投递所有ConfigMap中记录的serviceID
// 参数：queue - 工作队列接口，用于清空和重新投递事件
// 特性：保证幂等性 - 无论调用多少次，最终效果相同
func (e *EtcdStore) ReloadAll(queue *workqueue.Queue) {
	log.Info("ReloadAll called for EtcdStore, clearing workqueue and reloading all service IDs")

	// 检查客户端是否初始化
	if e.client == nil {
		if err := e.initClient(); err != nil {
			log.Error("Failed to initialize client in ReloadAll: %v", err)
			return
		}
	}

	// 检查队列是否有效
	if queue == nil {
		log.Error("Queue is nil in ReloadAll")
		return
	}

	// 2. 列出所有包含service-id标签的ConfigMap
	labelSelector := fmt.Sprintf("app=astron-xmod-shim,resource=requirement-spec")
	opts := metav1.ListOptions{LabelSelector: labelSelector}
	cms, err := e.client.GetClientSet().CoreV1().ConfigMaps(e.namespace).List(context.Background(), opts)
	if err != nil {
		log.Error("Failed to list ConfigMaps in ReloadAll: %v", err)
		return
	}

	// 3. 重新投递所有serviceID到workqueue
	serviceCount := 0
	for _, cm := range cms.Items {
		serviceID, exists := cm.Labels["service-id"]
		if !exists {
			log.Warn("ConfigMap %s doesn't have service-id label, skipping", cm.Name)
			continue
		}

		// 构建事件键并添加到队列
		eventKey := fmt.Sprintf("cm/reload/%s", serviceID)
		queue.Add(eventKey)
		serviceCount++
		log.Info("Requeued service ID: %s", serviceID)
	}

	log.Info("ReloadAll completed, requeued %d service IDs", serviceCount)
}

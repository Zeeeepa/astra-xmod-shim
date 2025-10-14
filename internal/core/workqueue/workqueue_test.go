package workqueue

import (
	"testing"
	"time"
)

func TestQueueBasicOperations(t *testing.T) {
	// 创建队列
	q := New()
	defer q.ShutDown()

	// 测试添加和获取
	q.Add("test-key")
	if q.Len() != 1 {
		t.Errorf("Expected queue length 1, got %d", q.Len())
	}

	key, done := q.Get()
	if key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", key)
	}
	if q.Len() != 0 {
		t.Errorf("Expected queue length 0 after Get, got %d", q.Len())
	}

	// 测试done函数
	done()
}

func TestQueueAddAfter(t *testing.T) {
	// 创建队列
	q := New()
	defer q.ShutDown()

	// 测试延迟添加
	q.AddAfter("delayed-key", 100*time.Millisecond)
	if q.Len() != 0 {
		t.Errorf("Expected queue length 0 immediately after AddAfter, got %d", q.Len())
	}

	// 等待延迟时间
	time.Sleep(150 * time.Millisecond)
	if q.Len() != 1 {
		t.Errorf("Expected queue length 1 after delay, got %d", q.Len())
	}

	// 获取延迟添加的项
	key, done := q.Get()
	if key != "delayed-key" {
		t.Errorf("Expected key 'delayed-key', got '%s'", key)
	}
	done()
}

func TestQueueRequeuing(t *testing.T) {
	// 创建队列
	q := New()
	defer q.ShutDown()

	// 添加项并获取
	q.Add("retry-key")
	_, done := q.Get()

	// 不调用done，而是重新添加到队列
	done() // 先标记完成
	q.Add("retry-key")

	// 检查重试计数
	if q.NumRequeues("retry-key") != 0 {
		t.Errorf("Expected 0 requeues, got %d", q.NumRequeues("retry-key"))
	}

	// 忘记该项（清除重试计数）
	q.Forget("retry-key")
}

func TestQueueShutDown(t *testing.T) {
	// 创建队列
	q := New()

	// 添加项
	q.Add("shutdown-key")

	// 关闭队列
	q.ShutDown()

	// 尝试添加项（应该没有效果）
	q.Add("after-shutdown")

	// 检查队列长度
	if q.Len() != 1 {
		t.Errorf("Expected queue length 1 after shutdown, got %d", q.Len())
	}
}
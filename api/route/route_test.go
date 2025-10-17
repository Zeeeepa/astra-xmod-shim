package route

import (
	// 为自定义 http 包添加别名以避免命名冲突
	xmodhttp "astron-xmod-shim/pkg/http"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockServer 模拟 HTTP 服务器
// 注意：这个结构体现在只用于模拟，不会直接传递给 RegisterRoutes
type MockServer struct {
	mock.Mock
	engine *gin.Engine
}

func (m *MockServer) GetEngine() *gin.Engine {
	args := m.Called()
	return args.Get(0).(*gin.Engine)
}

func (m *MockServer) Run() error {
	args := m.Called()
	return args.Error(0)
}

func TestRegisterRoutes(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 创建一个符合 RegisterRoutes 参数要求的服务器
	// 使用 NewServer 函数而不是直接初始化结构体
	testServer := xmodhttp.NewServer(":8080")
	// 获取引擎用于测试
	engine := testServer.GetEngine()

	// 创建一个模拟的 HTTP 服务器（仅用于验证调用）
	mockServer := new(MockServer)
	mockServer.On("GetEngine").Return(engine)

	// 注册路由
	RegisterRoutes(testServer)

	// 验证路由是否被正确注册
	// 创建测试请求来验证各个路由

	// 测试部署路由
	r1, _ := http.NewRequest("POST", "/api/v1/modserv/deploy", nil)
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, r1)
	assert.NotEqual(t, http.StatusNotFound, w1.Code)

	// 测试模型列表路由
	r2, _ := http.NewRequest("GET", "/api/v1/modserv/list", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, r2)
	assert.NotEqual(t, http.StatusNotFound, w2.Code)

	// 测试获取服务状态路由
	r3, _ := http.NewRequest("GET", "/api/v1/modserv/test-service-id", nil)
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, r3)
	assert.NotEqual(t, http.StatusNotFound, w3.Code)

	// 测试删除服务路由
	r4, _ := http.NewRequest("DELETE", "/api/v1/modserv/test-service-id", nil)
	w4 := httptest.NewRecorder()
	engine.ServeHTTP(w4, r4)
	assert.NotEqual(t, http.StatusNotFound, w4.Code)

	// 测试更新服务路由
	r5, _ := http.NewRequest("PUT", "/api/v1/modserv/test-service-id", nil)
	w5 := httptest.NewRecorder()
	engine.ServeHTTP(w5, r5)
	assert.NotEqual(t, http.StatusNotFound, w5.Code)

	// 测试指标路由
	r6, _ := http.NewRequest("GET", "/api/v1/modserv/metrics", nil)
	w6 := httptest.NewRecorder()
	engine.ServeHTTP(w6, r6)
	assert.NotEqual(t, http.StatusNotFound, w6.Code)
}

func TestRegisterRoutes_NonExistentRoute(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 创建一个符合 RegisterRoutes 参数要求的服务器
	testServer := xmodhttp.NewServer(":8080")
	// 获取引擎用于测试
	engine := testServer.GetEngine()

	// 注册路由
	RegisterRoutes(testServer)

	// 测试不存在的路由
	r, _ := http.NewRequest("GET", "/api/v1/non-existent-route", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, r)

	// 验证返回 404 状态码
	assert.Equal(t, http.StatusNotFound, w.Code)
}
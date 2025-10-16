package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// 测试创建新服务器
	addr := ":8080"
	server := NewServer(addr)

	// 验证服务器实例
	assert.NotNil(t, server)
	assert.Equal(t, addr, server.addr)
	assert.NotNil(t, server.engine)

	// 验证gin引擎已正确初始化
	assert.IsType(t, &gin.Engine{}, server.engine)
}

func TestGetEngine(t *testing.T) {
	// 测试获取引擎
	server := NewServer(":8080")
	engine := server.GetEngine()

	// 验证返回的引擎
	assert.NotNil(t, engine)
	assert.Equal(t, server.engine, engine)
}

func TestServerRun(t *testing.T) {
	// 测试服务器运行（不实际启动服务器）
	server := NewServer(":0") // 使用端口0让系统自动分配
	engine := server.GetEngine()

	// 添加测试路由
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// 创建测试请求
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 处理请求
	engine.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestServerWithMultipleRoutes(t *testing.T) {
	// 测试多个路由
	server := NewServer(":8080")
	engine := server.GetEngine()

	// 添加多个路由
	engine.GET("/api/v1/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"users": []string{"user1", "user2"}})
	})

	engine.POST("/api/v1/users", func(c *gin.Context) {
		c.JSON(201, gin.H{"message": "user created"})
	})

	engine.DELETE("/api/v1/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{"message": "user " + id + " deleted"})
	})

	// 测试GET请求
	req1, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)
	assert.Contains(t, w1.Body.String(), "users")

	// 测试POST请求
	req2, _ := http.NewRequest("POST", "/api/v1/users", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	assert.Equal(t, 201, w2.Code)
	assert.Contains(t, w2.Body.String(), "user created")

	// 测试DELETE请求
	req3, _ := http.NewRequest("DELETE", "/api/v1/users/123", nil)
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
	assert.Contains(t, w3.Body.String(), "user 123 deleted")
}

func TestServerMiddleware(t *testing.T) {
	// 测试中间件
	server := NewServer(":8080")
	engine := server.GetEngine()

	// 添加自定义中间件
	engine.Use(func(c *gin.Context) {
		c.Header("X-Custom-Header", "test-value")
		c.Next()
	})

	// 添加路由
	engine.GET("/middleware-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "middleware test"})
	})

	// 测试请求
	req, _ := http.NewRequest("GET", "/middleware-test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// 验证响应和头部
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "middleware test")
	assert.Equal(t, "test-value", w.Header().Get("X-Custom-Header"))
}

func TestServerErrorHandling(t *testing.T) {
	// 测试错误处理
	server := NewServer(":8080")
	engine := server.GetEngine()

	// 添加错误处理路由
	engine.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal server error"})
	})

	engine.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// 测试500错误
	req1, _ := http.NewRequest("GET", "/error", nil)
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	assert.Equal(t, 500, w1.Code)
	assert.Contains(t, w1.Body.String(), "internal server error")

	// 测试panic恢复（gin默认会恢复panic）
	req2, _ := http.NewRequest("GET", "/panic", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	assert.Equal(t, 500, w2.Code) // gin会返回500
}
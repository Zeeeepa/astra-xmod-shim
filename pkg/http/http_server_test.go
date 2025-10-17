package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	// Test creating a new server
	addr := ":8080"
	server := NewServer(addr)

	// Verify server instance
	assert.NotNil(t, server)
	assert.Equal(t, addr, server.addr)
	assert.NotNil(t, server.engine)

	// Verify gin engine is properly initialized
	assert.IsType(t, &gin.Engine{}, server.engine)
}

func TestGetEngine(t *testing.T) {
	// Test getting engine
	server := NewServer(":8080")
	engine := server.GetEngine()

	// Verify returned engine
	assert.NotNil(t, engine)
	assert.Equal(t, server.engine, engine)
}

func TestServerRun(t *testing.T) {
	// Test server running (without actually starting the server)
	server := NewServer(":0") // Use port 0 to let the system assign automatically
	engine := server.GetEngine()

	// Add test route
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Create test request
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// Create response recorder
	w := httptest.NewRecorder()

	// Handle request
	engine.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestServerWithMultipleRoutes(t *testing.T) {
	// Test multiple routes
	server := NewServer(":8080")
	engine := server.GetEngine()

	// Add multiple routes
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

	// Test GET request
	req1, _ := http.NewRequest("GET", "/api/v1/users", nil)
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)
	assert.Contains(t, w1.Body.String(), "users")

	// Test POST request
	req2, _ := http.NewRequest("POST", "/api/v1/users", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	assert.Equal(t, 201, w2.Code)
	assert.Contains(t, w2.Body.String(), "user created")

	// Test DELETE request
	req3, _ := http.NewRequest("DELETE", "/api/v1/users/123", nil)
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
	assert.Contains(t, w3.Body.String(), "user 123 deleted")
}

func TestServerMiddleware(t *testing.T) {
	// Test middleware
	server := NewServer(":8080")
	engine := server.GetEngine()

	// Add custom middleware
	engine.Use(func(c *gin.Context) {
		c.Header("X-Custom-Header", "test-value")
		c.Next()
	})

	// Add route
	engine.GET("/middleware-test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "middleware test"})
	})

	// Test request
	req, _ := http.NewRequest("GET", "/middleware-test", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	// Verify response and headers
	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "middleware test")
	assert.Equal(t, "test-value", w.Header().Get("X-Custom-Header"))
}

func TestServerErrorHandling(t *testing.T) {
	// Test error handling
	server := NewServer(":8080")
	engine := server.GetEngine()

	// Add error handling routes
	engine.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal server error"})
	})

	engine.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	// Test 500 error
	req1, _ := http.NewRequest("GET", "/error", nil)
	w1 := httptest.NewRecorder()
	engine.ServeHTTP(w1, req1)
	assert.Equal(t, 500, w1.Code)
	assert.Contains(t, w1.Body.String(), "internal server error")

	// Test panic recovery (gin automatically recovers from panic)
	req2, _ := http.NewRequest("GET", "/panic", nil)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	assert.Equal(t, 500, w2.Code) // gin会返回500
}
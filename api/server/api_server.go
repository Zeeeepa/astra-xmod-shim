package server

import (
	"astron-xmod-shim/api/middleware"
	"astron-xmod-shim/api/route"
	"astron-xmod-shim/internal/config"
	"astron-xmod-shim/pkg/http"
	"astron-xmod-shim/pkg/log"

	"github.com/gin-gonic/gin"
)

// Init Start HTTP server
func Init() error {

	gin.SetMode(gin.ReleaseMode) // Set before initializing Engine
	// 2. Get config on demand later (complete initialization on first Get() call)
	globalCfg := config.Get()
	log.Info("HTTP server address and port %v", globalCfg.Server.Port)

	// 3. Initialize generic HTTP server
	httpServer := http.NewServer(globalCfg.Server.Port)

	// Register business routes
	route.RegisterRoutes(httpServer)

	// Register logging middleware
	engine := httpServer.GetEngine()
	engine.Use(middleware.Logging())

	log.Info("HTTP server initialized")

	// Start server
	return httpServer.Run()
}
package http

import (
	"github.com/gin-gonic/gin"
)

// Server Generic HTTP server
 type Server struct {
	engine *gin.Engine // Internal gin engine
	addr   string
}

// NewServer Create HTTP server instance
func NewServer(addr string) *Server {
	return &Server{
		engine: gin.Default(),
		addr:   addr,
	}
}

// GetEngine Provide engine access method
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// Run Start the server
func (s *Server) Run() error {
	return s.engine.Run(s.addr)
}
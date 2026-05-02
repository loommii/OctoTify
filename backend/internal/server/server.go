package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	engine *gin.Engine
	addr   string
	logger *zap.Logger
}

func New(addr, mode string, logger *zap.Logger) *Server {
	gin.SetMode(mode)

	s := &Server{
		engine: gin.New(),
		addr:   addr,
		logger: logger,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.engine.Use(gin.Recovery())
	s.engine.Use(gin.Logger())
}

func (s *Server) setupRoutes() {
	api := s.engine.Group("/api")
	{
		s.setupAuthRoutes(api)
		s.setupSourceRoutes(api)
		s.setupChannelRoutes(api)
		s.setupMessageRoutes(api)
		s.setupPushRoutes(api)
	}
}

func (s *Server) setupAuthRoutes(api *gin.RouterGroup) {
	// POST /api/register
	// POST /api/login
	// POST /api/token/refresh
	// PUT  /api/password
	// GET  /api/user/profile
}

func (s *Server) setupSourceRoutes(api *gin.RouterGroup) {
	// POST   /api/sources
	// PUT    /api/sources/:id
	// GET    /api/sources
	// GET    /api/sources/:id
	// GET    /api/sources/:id/token
	// PUT    /api/sources/:id/token
	// PUT    /api/sources/:id/disable
	// PUT    /api/sources/:id/enable
	// DELETE /api/sources/:id
}

func (s *Server) setupChannelRoutes(api *gin.RouterGroup) {
	// POST   /api/channels
	// PUT    /api/channels/:id
	// GET    /api/channels
	// GET    /api/channels/:id
	// POST   /api/channels/:id/test
	// PUT    /api/channels/:id/disable
	// PUT    /api/channels/:id/enable
	// DELETE /api/channels/:id
}

func (s *Server) setupMessageRoutes(api *gin.RouterGroup) {
	// GET    /api/messages
	// GET    /api/messages/:id
	// DELETE /api/messages/:id
}

func (s *Server) setupPushRoutes(api *gin.RouterGroup) {
	// POST /api/message/push
}

func (s *Server) Run() error {
	s.logger.Info("server starting", zap.String("addr", s.addr))
	if err := s.engine.Run(s.addr); err != nil {
		return fmt.Errorf("server run failed: %w", err)
	}
	return nil
}

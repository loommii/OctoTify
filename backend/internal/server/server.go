package server

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"octotify/internal/handler"
	"octotify/internal/middleware"
	apperrors "octotify/pkg/errors"
	"octotify/pkg/response"
)

type Server struct {
	engine     *gin.Engine
	addr       string
	serverName string
	logger     *zap.Logger
}

func New(addr, mode, serverName string, logger *zap.Logger) *Server {
	gin.SetMode(mode)
	s := &Server{
		engine:     gin.New(),
		addr:       addr,
		serverName: serverName,
		logger:     logger,
	}

	s.setupMiddleware()
	s.setupNoRoute()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.engine.Use(middleware.RequestID())
	s.engine.Use(middleware.CustomRecovery(s.logger))
	s.engine.Use(gin.Logger())
	s.engine.Use(middleware.ErrorHandler(s.logger))
}

// setupNoRoute 注册 404/405 处理，覆盖 Gin 默认的纯文本响应
func (s *Server) setupNoRoute() {
	// 路由不存在时返回统一格式
	s.engine.NoRoute(func(c *gin.Context) {
		response.Fail(c, apperrors.ErrNotFound.Code, "接口不存在")
	})

	// 启用 405 检测并返回统一格式
	s.engine.HandleMethodNotAllowed = true
	s.engine.NoMethod(func(c *gin.Context) {
		response.Fail(c, apperrors.ErrMethodNotAllowed.Code, "请求方法不允许")
	})
}

func (s *Server) setupRoutes() {
	// 健康检查，无需鉴权
	s.engine.GET("/ping", handler.Ping(s.serverName))

	api := s.engine.Group("/api")

	s.setupUserRoutes(api)
	s.setupSourceRoutes(api)
	s.setupChannelRoutes(api)
	s.setupMessageRoutes(api)
	s.setupPushRoutes(api)
}

func (s *Server) setupUserRoutes(api *gin.RouterGroup) {
	user := api.Group("/user")
	{
		user.POST("/register", func(c *gin.Context) {})
		user.POST("/login", func(c *gin.Context) {})
		user.POST("/refresh-token", func(c *gin.Context) {})
		user.POST("/change-password", func(c *gin.Context) {})
		user.GET("/profile", func(c *gin.Context) {})
	}
}

func (s *Server) setupSourceRoutes(api *gin.RouterGroup) {
	source := api.Group("/sources")
	{
		source.POST("", func(c *gin.Context) {})
		source.PUT("/:id", func(c *gin.Context) {})
		source.GET("", func(c *gin.Context) {})
		source.GET("/:id", func(c *gin.Context) {})
		source.GET("/:id/token", func(c *gin.Context) {})
		source.POST("/:id/reset-token", func(c *gin.Context) {})
		source.POST("/:id/disable", func(c *gin.Context) {})
		source.POST("/:id/enable", func(c *gin.Context) {})
		source.DELETE("/:id", func(c *gin.Context) {})
	}
}

func (s *Server) setupChannelRoutes(api *gin.RouterGroup) {
	channel := api.Group("/channels")
	{
		channel.POST("", func(c *gin.Context) {})
		channel.PUT("/:id", func(c *gin.Context) {})
		channel.GET("", func(c *gin.Context) {})
		channel.GET("/:id", func(c *gin.Context) {})
		channel.POST("/:id/test", func(c *gin.Context) {})
		channel.POST("/:id/disable", func(c *gin.Context) {})
		channel.POST("/:id/enable", func(c *gin.Context) {})
		channel.DELETE("/:id", func(c *gin.Context) {})
	}
}

func (s *Server) setupMessageRoutes(api *gin.RouterGroup) {
	message := api.Group("/messages")
	{
		message.GET("", func(c *gin.Context) {})
		message.GET("/filter", func(c *gin.Context) {})
		message.GET("/:id", func(c *gin.Context) {})
		message.DELETE("/:id", func(c *gin.Context) {})
	}
}

func (s *Server) setupPushRoutes(api *gin.RouterGroup) {
	push := api.Group("/push")
	{
		push.POST("", func(c *gin.Context) {})
		push.POST("/multi", func(c *gin.Context) {})
		push.POST("/status", func(c *gin.Context) {})
	}
}

func (s *Server) Run() error {
	s.logger.Info("server starting", zap.String("addr", s.addr))
	if err := s.engine.Run(s.addr); err != nil {
		return err
	}
	return nil
}

package server

import (
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"octotify/internal/config"
	"octotify/internal/handler"
	"octotify/internal/middleware"
	"octotify/internal/service"
	"octotify/pkg/jwtx"
	"octotify/pkg/response"
	"octotify/pkg/validator"
	"octotify/pkg/xerr"

	_ "octotify/docs"
)

// Server HTTP 服务器
type Server struct {
	engine     *gin.Engine // Gin 引擎实例
	addr       string      // 服务器监听地址
	serverName string      // 服务器名称
	logger     *zap.Logger // 日志记录器

	authHandler     *handler.AuthHandler   // 认证处理器
	userHandler     *handler.UserHandler   // 用户管理处理器
	sourceHandler   *handler.SourceHandler // 来源管理处理器
	accessJWTHelper *jwtx.JWTHelper        // Access Token JWT 辅助工具
}

// New 创建并初始化 HTTP 服务器实例
func New(addr string, cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Server {
	gin.SetMode(cfg.Server.Mode)
	s := &Server{
		engine:     gin.New(),
		addr:       addr,
		serverName: cfg.Server.Name,
		logger:     logger,
	}

	// 禁用信任代理（默认不信任任何代理）
	s.engine.SetTrustedProxies(nil)

	// 注册自定义参数验证器
	validator.Init()

	// 初始化依赖组件
	s.initDependencies(cfg, db, logger)
	// 注册中间件
	s.setupMiddleware()
	// 注册 404/405 处理
	s.setupNoRoute()
	// 注册路由
	s.setupRoutes()

	return s
}

// initDependencies 初始化服务器依赖组件（JWT、Service、Handler）
func (s *Server) initDependencies(cfg *config.Config, db *gorm.DB, logger *zap.Logger) {
	// 检查并生成 RSA 密钥对（如果不存在）
	if err := jwtx.EnsureRSAKeyPair(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath); err != nil {
		logger.Fatal("确保 RSA 密钥对存在失败", zap.Error(err))
	}

	// 加载 RSA 私钥
	privateKeyPEM, err := os.ReadFile(cfg.JWT.PrivateKeyPath)
	if err != nil {
		logger.Fatal("读取 RSA 私钥文件失败", zap.Error(err))
	}
	privateKey, err := jwtx.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		logger.Fatal("解析 RSA 私钥失败", zap.Error(err))
	}

	// 加载 RSA 公钥
	publicKeyPEM, err := os.ReadFile(cfg.JWT.PublicKeyPath)
	if err != nil {
		logger.Fatal("读取 RSA 公钥文件失败", zap.Error(err))
	}
	publicKey, err := jwtx.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		logger.Fatal("解析 RSA 公钥失败", zap.Error(err))
	}

	// 初始化 Access Token JWT 辅助工具
	accessJWTHelper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(cfg.JWT.AccessTTL),
	)

	// 初始化 Refresh Token JWT 辅助工具
	refreshJWTHelper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(cfg.JWT.RefreshTTL),
	)

	// 初始化认证服务及处理器
	authService := service.NewAuthService(db, accessJWTHelper, refreshJWTHelper, logger)
	s.authHandler = handler.NewAuthHandler(authService)

	// 初始化用户服务及处理器
	userService := service.NewUserService(db, accessJWTHelper, refreshJWTHelper, logger)
	s.userHandler = handler.NewUserHandler(userService)

	// 初始化来源服务及处理器
	sourceService := service.NewSourceService(db, logger)
	s.sourceHandler = handler.NewSourceHandler(sourceService)

	s.accessJWTHelper = accessJWTHelper
}

// setupMiddleware 注册全局中间件
func (s *Server) setupMiddleware() {
	s.engine.Use(middleware.RequestID())              // 请求 ID 生成与注入
	s.engine.Use(middleware.RequestLogger(s.logger))  // 请求日志记录
	s.engine.Use(middleware.CustomRecovery(s.logger)) // 自定义 Panic 恢复
	s.engine.Use(middleware.ErrorHandler(s.logger))   // 统一错误处理
}

// setupNoRoute 注册 404/405 默认处理
func (s *Server) setupNoRoute() {
	// 404 接口不存在
	s.engine.NoRoute(func(c *gin.Context) {
		response.Fail(c, xerr.ErrNotFound.Code, "接口不存在")
	})

	s.engine.HandleMethodNotAllowed = true
	// 405 请求方法不允许
	s.engine.NoMethod(func(c *gin.Context) {
		response.Fail(c, xerr.ErrMethodNotAllowed.Code, "请求方法不允许")
	})
}

// setupRoutes 注册所有路由
func (s *Server) setupRoutes() {
	// 健康检查接口
	s.engine.GET("/ping", handler.Ping(s.serverName))
	// Swagger API 文档
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.engine.Group("/api")

	s.setupAuthRoutes(api)
	s.setupUserRoutes(api)
	s.setupSourceRoutes(api)
	s.setupChannelRoutes(api)
	s.setupMessageRoutes(api)
	s.setupPushRoutes(api)
}

// setupAuthRoutes 注册认证相关路由
func (s *Server) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", s.authHandler.Login)                                          // 用户登录
		auth.POST("/refresh", s.authHandler.RefreshToken)                                 // 刷新 Token
		auth.POST("/logout", middleware.JWTAuth(s.accessJWTHelper), s.authHandler.Logout) // 退出登录（需 JWT 认证）
	}
}

// setupUserRoutes 注册用户管理相关路由
func (s *Server) setupUserRoutes(api *gin.RouterGroup) {
	user := api.Group("/user")
	{
		user.POST("/register", s.userHandler.Register)                                             // 用户注册
		user.PUT("/password", middleware.JWTAuth(s.accessJWTHelper), s.userHandler.ChangePassword) // 修改密码（需 JWT 认证）
		user.GET("/profile", middleware.JWTAuth(s.accessJWTHelper), s.userHandler.GetUserProfile)  // 获取用户信息（需 JWT 认证）
	}
}

// setupSourceRoutes 注册消息来源管理相关路由
func (s *Server) setupSourceRoutes(api *gin.RouterGroup) {
	source := api.Group("/sources")
	source.Use(middleware.JWTAuth(s.accessJWTHelper)) // 所有来源接口均需 JWT 认证
	{
		source.POST("", s.sourceHandler.CreateSource)              // 创建消息来源
		source.PUT("/:id", s.sourceHandler.UpdateSource)           // 编辑消息来源
		source.GET("", s.sourceHandler.ListSources)                // 查看来源列表
		source.GET("/:id", s.sourceHandler.GetSourceDetail)        // 查看来源详情
		source.POST("/:id/token", s.sourceHandler.GetSourceToken)  // 查看来源令牌（需密码二次验证）
		source.PUT("/:id/token", s.sourceHandler.ResetSourceToken) // 重置来源令牌（需密码二次验证）
		source.PUT("/:id/disable", s.sourceHandler.DisableSource)  // 停用消息来源（需密码二次验证）
		source.PUT("/:id/enable", s.sourceHandler.EnableSource)    // 启用消息来源（需密码二次验证）
		source.DELETE("/:id", s.sourceHandler.DeleteSource)        // 删除消息来源（需密码二次验证）
	}
}

// setupChannelRoutes 注册推送渠道管理相关路由（占位）
func (s *Server) setupChannelRoutes(api *gin.RouterGroup) {
	channel := api.Group("/channels")
	{
		channel.POST("", func(c *gin.Context) {})             // 创建渠道
		channel.PUT("/:id", func(c *gin.Context) {})          // 编辑渠道
		channel.GET("", func(c *gin.Context) {})              // 查看渠道列表
		channel.GET("/:id", func(c *gin.Context) {})          // 查看渠道详情
		channel.POST("/:id/test", func(c *gin.Context) {})    // 测试渠道连接
		channel.POST("/:id/disable", func(c *gin.Context) {}) // 停用渠道
		channel.POST("/:id/enable", func(c *gin.Context) {})  // 启用渠道
		channel.DELETE("/:id", func(c *gin.Context) {})       // 删除渠道
	}
}

// setupMessageRoutes 注册消息管理相关路由（占位）
func (s *Server) setupMessageRoutes(api *gin.RouterGroup) {
	message := api.Group("/messages")
	{
		message.GET("", func(c *gin.Context) {})        // 查看消息列表
		message.GET("/filter", func(c *gin.Context) {}) // 筛选消息
		message.GET("/:id", func(c *gin.Context) {})    // 查看消息详情
		message.DELETE("/:id", func(c *gin.Context) {}) // 删除消息
	}
}

// setupPushRoutes 注册消息推送相关路由（占位）
func (s *Server) setupPushRoutes(api *gin.RouterGroup) {
	push := api.Group("/push")
	{
		push.POST("", func(c *gin.Context) {})        // 单条推送
		push.POST("/multi", func(c *gin.Context) {})  // 批量推送
		push.POST("/status", func(c *gin.Context) {}) // 推送状态回调
	}
}

// Run 启动 HTTP 服务器
func (s *Server) Run() error {
	s.logger.Info("服务器启动", zap.String("addr", s.addr))
	if err := s.engine.Run(s.addr); err != nil {
		return err
	}
	return nil
}

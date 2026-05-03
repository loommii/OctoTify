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

type Server struct {
	engine     *gin.Engine
	addr       string
	serverName string
	logger     *zap.Logger

	authHandler     *handler.AuthHandler
	userHandler     *handler.UserHandler
	sourceHandler   *handler.SourceHandler
	accessJWTHelper *jwtx.JWTHelper
}

func New(addr string, cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Server {
	gin.SetMode(cfg.Server.Mode)
	s := &Server{
		engine:     gin.New(),
		addr:       addr,
		serverName: cfg.Server.Name,
		logger:     logger,
	}

	s.engine.SetTrustedProxies(nil)

	// 注册自定义验证器
	validator.Init()

	s.initDependencies(cfg, db, logger)
	s.setupMiddleware()
	s.setupNoRoute()
	s.setupRoutes()

	return s
}

func (s *Server) initDependencies(cfg *config.Config, db *gorm.DB, logger *zap.Logger) {
	// 检查并生成 RSA 密钥对（如果不存在）
	if err := jwtx.EnsureRSAKeyPair(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath); err != nil {
		logger.Fatal("failed to ensure RSA key pair", zap.Error(err))
	}

	// 加载 RSA 私钥
	privateKeyPEM, err := os.ReadFile(cfg.JWT.PrivateKeyPath)
	if err != nil {
		logger.Fatal("failed to read private key", zap.Error(err))
	}
	privateKey, err := jwtx.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		logger.Fatal("failed to parse private key", zap.Error(err))
	}

	// 加载 RSA 公钥
	publicKeyPEM, err := os.ReadFile(cfg.JWT.PublicKeyPath)
	if err != nil {
		logger.Fatal("failed to read public key", zap.Error(err))
	}
	publicKey, err := jwtx.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		logger.Fatal("failed to parse public key", zap.Error(err))
	}

	// 初始化 JWT 辅助工具（双 Helper 实例，分别管理 access 和 refresh 的过期时间）
	accessJWTHelper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(cfg.JWT.AccessTTL),
	)

	refreshJWTHelper := jwtx.NewJWTHelper(
		jwtx.WithPrivateKey(privateKey),
		jwtx.WithPublicKey(publicKey),
		jwtx.WithExpiredTime(cfg.JWT.RefreshTTL),
	)

	authService := service.NewAuthService(db, accessJWTHelper, refreshJWTHelper, logger)
	s.authHandler = handler.NewAuthHandler(authService)

	userService := service.NewUserService(db, accessJWTHelper, refreshJWTHelper, logger)
	s.userHandler = handler.NewUserHandler(userService)

	sourceService := service.NewSourceService(db, logger)
	s.sourceHandler = handler.NewSourceHandler(sourceService)

	s.accessJWTHelper = accessJWTHelper
}

func (s *Server) setupMiddleware() {
	s.engine.Use(middleware.RequestID())
	s.engine.Use(middleware.RequestLogger(s.logger))
	s.engine.Use(middleware.CustomRecovery(s.logger))
	s.engine.Use(middleware.ErrorHandler(s.logger))
}

func (s *Server) setupNoRoute() {
	s.engine.NoRoute(func(c *gin.Context) {
		response.Fail(c, xerr.ErrNotFound.Code, "接口不存在")
	})

	s.engine.HandleMethodNotAllowed = true
	s.engine.NoMethod(func(c *gin.Context) {
		response.Fail(c, xerr.ErrMethodNotAllowed.Code, "请求方法不允许")
	})
}

func (s *Server) setupRoutes() {
	s.engine.GET("/ping", handler.Ping(s.serverName))
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.engine.Group("/api")

	s.setupAuthRoutes(api)
	s.setupUserRoutes(api)
	s.setupSourceRoutes(api)
	s.setupChannelRoutes(api)
	s.setupMessageRoutes(api)
	s.setupPushRoutes(api)
}

func (s *Server) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", s.authHandler.Login)
		auth.POST("/refresh", s.authHandler.RefreshToken)
		auth.POST("/logout", middleware.JWTAuth(s.accessJWTHelper), s.authHandler.Logout)
	}
}

func (s *Server) setupUserRoutes(api *gin.RouterGroup) {
	user := api.Group("/user")
	{
		user.POST("/register", s.userHandler.Register)
		user.PUT("/password", middleware.JWTAuth(s.accessJWTHelper), s.userHandler.ChangePassword)
		user.GET("/profile", middleware.JWTAuth(s.accessJWTHelper), s.userHandler.GetUserProfile)
	}
}

func (s *Server) setupSourceRoutes(api *gin.RouterGroup) {
	source := api.Group("/sources")
	source.Use(middleware.JWTAuth(s.accessJWTHelper))
	{
		source.POST("", s.sourceHandler.CreateSource)
		source.PUT("/:id", s.sourceHandler.UpdateSource)
		source.GET("", s.sourceHandler.ListSources)
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

package server

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"octotify/internal/client/ilink"
	"octotify/internal/config"
	"octotify/internal/handler"
	"octotify/internal/middleware"
	"octotify/internal/sender"
	"octotify/internal/service"
	"octotify/pkg/jwtx"
	"octotify/pkg/response"
	"octotify/pkg/validator"
	"octotify/pkg/xerr"
)

// Server HTTP 服务器
type Server struct {
	engine     *gin.Engine    // Gin 引擎实例
	addr       string         // 服务器监听地址
	serverName string         // 服务器名称
	logger     *zap.Logger    // 日志记录器
	cfg        *config.Config // 配置信息

	authHandler     *handler.AuthHandler    // 认证处理器
	userHandler     *handler.UserHandler    // 用户管理处理器
	sourceHandler   *handler.SourceHandler  // 来源管理处理器
	channelHandler  *handler.ChannelHandler // 渠道管理处理器
	messageHandler  *handler.MessageHandler // 消息管理处理器
	pushHandler     *handler.PushHandler    // 消息推送处理器
	accessJWTHelper *jwtx.JWTHelper         // Access Token JWT 辅助工具

	sourceService *service.SourceService // 来源服务（用于 StepUpAuth 中间件）
}

// New 创建并初始化 HTTP 服务器实例
func New(addr string, cfg *config.Config, db *gorm.DB, logger *zap.Logger) *Server {
	gin.SetMode(cfg.Server.Mode)
	s := &Server{
		engine:     gin.New(),
		addr:       addr,
		serverName: cfg.Server.Name,
		logger:     logger,
		cfg:        cfg,
	}

	// 禁用信任代理（默认不信任任何代理）
	s.engine.SetTrustedProxies(nil)

	// 注册自定义参数验证器
	validator.Init()

	// 初始化依赖组件
	s.initDependencies(cfg, db, logger)
	// 注册中间件
	s.setupMiddleware()
	// 注册 OpenAPI 文档
	s.setupOpenAPI()
	// 注册 404/405 处理
	s.setupNoRoute()
	// 注册路由
	s.setupRoutes()

	return s
}

// initDependencies 初始化服务器依赖组件（JWT、Service、Handler）
func (s *Server) initDependencies(cfg *config.Config, db *gorm.DB, logger *zap.Logger) {
	// 检查并生成 RSA 密钥对（如果不存在）
	if err := jwtx.EnsureRSAKeyPair(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath, cfg.JWT.AutoGenerateKeys); err != nil {
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
	s.sourceService = sourceService
	s.sourceHandler = handler.NewSourceHandler(sourceService)

	// 初始化 iLink 客户端（全局共享实例，多 bot 场景下复用 HTTP 连接池）
	ilinkClient := ilink.NewClient(s.logger)

	// 初始化渠道服务及处理器
	// ilinkClient 注入到 SenderFactory 和 ChannelService，统一管理 iLink 协议通信
	senderFactory := sender.NewSenderFactory(s.logger, ilinkClient)
	channelService := service.NewChannelService(db, logger, senderFactory, ilinkClient)
	s.channelHandler = handler.NewChannelHandler(channelService, logger)

	// 初始化消息服务及处理器
	messageService := service.NewMessageService(db, logger, senderFactory)
	s.messageHandler = handler.NewMessageHandler(messageService)
	s.pushHandler = handler.NewPushHandler(messageService)

	s.accessJWTHelper = accessJWTHelper
}

// setupMiddleware 注册全局中间件
func (s *Server) setupMiddleware() {
	s.engine.Use(middleware.RequestID())                                  // 请求 ID 生成与注入
	s.engine.Use(middleware.RequestLogger(s.logger, s.cfg.Log.DebugBody)) // 请求日志记录
	s.engine.Use(middleware.CustomRecovery(s.logger))                     // 自定义 Panic 恢复
	s.engine.Use(middleware.ErrorHandler(s.logger))                       // 统一错误处理
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

// Run 启动 HTTP 服务器
func (s *Server) Run() error {
	s.logger.Info("服务器启动", zap.String("addr", s.addr))
	if err := s.engine.Run(s.addr); err != nil {
		return err
	}
	return nil
}

// GetEngine 获取 Gin 引擎实例（用于 http.Server 包装）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// Close 优雅关闭服务器资源
func (s *Server) Close() {
	s.logger.Info("服务器资源已关闭")
}

// Shutdown 优雅关闭服务器资源
// 参数:
//   - ctx: 关闭超时上下文
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("服务器资源已关闭")
	return nil
}

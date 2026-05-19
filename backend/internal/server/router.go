package server

import (
	"github.com/gin-gonic/gin"

	"octotify/internal/handler"
	"octotify/internal/middleware"
)

// setupRoutes 注册所有路由
func (s *Server) setupRoutes() {
	// 健康检查接口
	s.engine.GET("/ping", handler.Ping(s.serverName))

	api := s.engine.Group("/api")

	// 注册各业务模块路由
	s.setupAuthRoutes(api)    // 用户认证
	s.setupUserRoutes(api)    // 用户管理
	s.setupSourceRoutes(api)  // 消息来源管理
	s.setupChannelRoutes(api) // 推送渠道管理
	s.setupMessageRoutes(api) // 消息管理
	s.setupPushRoutes(api)    // 消息推送
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
		source.POST("", s.sourceHandler.CreateSource)                                                                     // 创建消息来源
		source.PUT("/:id", s.sourceHandler.UpdateSource)                                                                  // 编辑消息来源
		source.GET("", s.sourceHandler.ListSources)                                                                       // 查看来源列表
		source.GET("/:id", s.sourceHandler.GetSourceDetail)                                                               // 查看来源详情
		source.POST("/:id/token", middleware.StepUpAuth(s.sourceService.VerifyPassword), s.sourceHandler.GetSourceToken)  // 查看来源令牌（需密码二次验证）
		source.PUT("/:id/token", middleware.StepUpAuth(s.sourceService.VerifyPassword), s.sourceHandler.ResetSourceToken) // 重置来源令牌（需密码二次验证）
		source.PUT("/:id/disable", middleware.StepUpAuth(s.sourceService.VerifyPassword), s.sourceHandler.DisableSource)  // 停用消息来源（需密码二次验证）
		source.PUT("/:id/enable", middleware.StepUpAuth(s.sourceService.VerifyPassword), s.sourceHandler.EnableSource)    // 启用消息来源（需密码二次验证）
		source.DELETE("/:id", middleware.StepUpAuth(s.sourceService.VerifyPassword), s.sourceHandler.DeleteSource)        // 删除消息来源（需密码二次验证）
	}
}

// setupChannelRoutes 注册推送渠道管理相关路由
func (s *Server) setupChannelRoutes(api *gin.RouterGroup) {
	channel := api.Group("/channels")
	channel.Use(middleware.JWTAuth(s.accessJWTHelper)) // 所有渠道接口均需 JWT 认证
	{
		channel.POST("", s.channelHandler.CreateChannel)             // 创建渠道
		channel.PUT("/:id", s.channelHandler.UpdateChannel)          // 编辑渠道
		channel.GET("", s.channelHandler.ListChannels)               // 查看渠道列表
		channel.GET("/:id", s.channelHandler.GetChannelDetail)       // 查看渠道详情
		channel.POST("/:id/test", s.channelHandler.TestChannel)      // 测试渠道连接
		channel.PUT("/:id/disable", s.channelHandler.DisableChannel) // 停用渠道
		channel.PUT("/:id/enable", s.channelHandler.EnableChannel)   // 启用渠道
		channel.DELETE("/:id", s.channelHandler.DeleteChannel)       // 删除渠道
	}

	// 渠道类型元数据接口（在 channels 组外，避免与 :id 路由冲突）
	api.GET("/channel-types", middleware.JWTAuth(s.accessJWTHelper), s.channelHandler.GetChannelTypes) // 获取渠道类型元数据

	// 微信ClawBot绑定路由
	wechatClawbot := api.Group("/channels/wechat-clawbot")
	wechatClawbot.Use(middleware.JWTAuth(s.accessJWTHelper)) // 需 JWT 认证
	{
		wechatClawbot.POST("/bind", s.channelHandler.StartBind)            // 发起扫码绑定
		wechatClawbot.POST("/bind/status", s.channelHandler.GetBindStatus) // 查询绑定状态
		wechatClawbot.POST("/check-activation", s.channelHandler.CheckActivation) // 检查激活消息
	}
}

// setupMessageRoutes 注册消息管理相关路由
func (s *Server) setupMessageRoutes(api *gin.RouterGroup) {
	message := api.Group("/messages")
	message.Use(middleware.JWTAuth(s.accessJWTHelper)) // 所有消息接口均需 JWT 认证
	{
		message.GET("", s.messageHandler.ListMessages)          // 查看消息列表
		message.GET("/filter", s.messageHandler.FilterMessages) // 筛选消息
		message.GET("/:id", s.messageHandler.GetMessageDetail)  // 查看消息详情
		message.DELETE("/:id", s.messageHandler.DeleteMessage)  // 删除消息
	}
}

// setupPushRoutes 注册消息推送相关路由
func (s *Server) setupPushRoutes(api *gin.RouterGroup) {
	push := api.Group("/push")
	{
		push.POST("/:token", s.pushHandler.PushMessage) // 推送消息（通过来源 Token）
	}
}

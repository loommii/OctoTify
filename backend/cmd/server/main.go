package main

//go:generate sh -c "cd ../.. && go run tools/generate/main.go"
//go:generate sh -c "cd ../.. && swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal"

// @title           OctoTify API
// @version         1.0
// @description     OctoTify 是一个消息总线平台，支持多种消息来源和推送渠道。
// @description     核心功能：消息来源管理、推送渠道管理、消息推送与记录。
// @description     本文档为前端开发人员提供完整的 API 接口说明，包含请求参数、响应格式、错误码等详细信息。
// @termsOfService  http://swagger.io/terms/

// @contact.name    OctoTify API Support
// @contact.url     https://github.com/OctoTify

// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host            localhost:34123
// @BasePath        /api

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 输入格式：Bearer {access_token}，例如：Bearer eyJhbGciOiJSUzI1NiIs...

// @securityDefinitions.apikey  SourceTokenAuth
// @in                          header
// @name                        Authorization
// @description                 推送消息时使用 Source Token，输入格式：Bearer src{uuid}，例如：Bearer src0196a3b2c4d50000a1b2c3d4e5f67890

// @tag.name                    系统
// @tag.description             系统健康检查等基础接口
// @tag.name                    用户认证
// @tag.description             用户登录、退出登录、刷新令牌等认证相关接口
// @tag.name                    用户管理
// @tag.description             用户注册、修改密码、查询用户信息等管理接口
// @tag.name                    消息来源管理
// @tag.description             创建、编辑、删除消息来源，管理来源令牌等接口
// @tag.name                    推送渠道管理
// @tag.description             创建、编辑、删除推送渠道，测试渠道连接等接口
// @tag.name                    消息管理
// @tag.description             查询消息列表、筛选消息、查看消息详情、删除消息等接口
// @tag.name                    消息推送
// @tag.description             外部系统通过 Source Token 推送消息到平台
// @tag.name                    微信ClawBot绑定
// @tag.description             发起绑定、轮询绑定状态等微信ClawBot绑定相关接口

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"octotify/internal/config"
	"octotify/internal/database"
	"octotify/internal/log"
	"octotify/internal/model"
	"octotify/internal/server"
)

// main 是应用程序的入口函数
// 负责初始化配置、日志、数据库，并启动 HTTP 服务器
func main() {
	// 解析命令行参数，指定配置文件路径
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	// 加载配置文件
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志系统
	logger, err := log.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.LogFile, cfg.Log.ErrorFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 初始化数据库连接
	db, err := database.NewDB(cfg.Database.Path, logger)
	if err != nil {
		logger.Fatal("init database failed", zap.Error(err))
	}

	// 执行数据库自动迁移，创建或更新表结构
	if err := db.AutoMigrate(
		&model.User{},          // 用户表
		&model.Source{},        // 消息来源表
		&model.Channel{},       // 推送渠道表
		&model.Message{},       // 消息表
		&model.SourceChannel{}, // 来源与渠道关联表
		&model.RefreshToken{},  // 刷新令牌表
	); err != nil {
		logger.Fatal("auto migrate failed", zap.Error(err))
	}

	// 构建服务器监听地址
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	// 创建服务器实例
	srv := server.New(addr, cfg, db, logger)

	// 启动 HTTP 服务器（使用标准库 http.Server 包装，支持优雅关闭）
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: srv.GetEngine(),
	}

	// 在独立 goroutine 中启动服务，避免阻塞主线程
	go func() {
		logger.Info("服务器启动", zap.String("addr", addr))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server run failed", zap.Error(err))
		}
	}()

	// 监听系统信号，实现优雅关闭
	// 当收到 SIGINT (Ctrl+C) 或 SIGTERM 信号时，触发关闭流程
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到关闭信号，开始优雅关闭...")

	// 设置关闭超时时间，确保服务在规定时间内完成现有请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器（停止接收新请求，等待已有请求完成）
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error("HTTP 服务器关闭失败", zap.Error(err))
	}

	// 关闭后台资源（数据库连接、定时任务等）
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器资源关闭失败", zap.Error(err))
	}

	logger.Info("服务器已优雅关闭")
}

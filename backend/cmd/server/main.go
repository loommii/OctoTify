package main

//go:generate go run ../../tools/generate/main.go
//go:generate swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal

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

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"octotify/internal/config"
	"octotify/internal/database"
	"octotify/internal/log"
	"octotify/internal/model"
	"octotify/internal/server"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config failed: %v\n", err)
		os.Exit(1)
	}

	logger, err := log.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output, cfg.Log.LogFile, cfg.Log.ErrorFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	db, err := database.NewDB(cfg.Database.Path, logger)
	if err != nil {
		logger.Fatal("init database failed", zap.Error(err))
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Source{},
		&model.Channel{},
		&model.Message{},
		&model.SourceChannel{},
		&model.RefreshToken{},
	); err != nil {
		logger.Fatal("auto migrate failed", zap.Error(err))
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := server.New(addr, cfg, db, logger)

	if err := srv.Run(); err != nil {
		logger.Fatal("server run failed", zap.Error(err))
	}
}

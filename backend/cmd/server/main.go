package main

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

	logger, err := log.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.Log.Output)
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
	); err != nil {
		logger.Fatal("auto migrate failed", zap.Error(err))
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := server.New(addr, cfg.Server.Mode, logger)

	if err := srv.Run(); err != nil {
		logger.Fatal("server run failed", zap.Error(err))
	}
}

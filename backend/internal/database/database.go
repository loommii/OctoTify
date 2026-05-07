package database

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDB(dbPath string, log *zap.Logger) (*gorm.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	gormLogger := logger.New(
		&zapLogger{log: log},
		logger.Config{
			LogLevel: logger.Warn,
		},
	)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

type zapLogger struct {
	log *zap.Logger
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.log.Sugar().Infof(format, args...)
}

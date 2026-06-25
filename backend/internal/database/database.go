package database

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	// 使用 WAL 模式提升并发读写性能，busy_timeout 防止 SQLITE_BUSY 错误
	// mattn/go-sqlite3 驱动支持通过 DSN query 参数设置 PRAGMA
	dsn := buildDSN(dbPath)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// buildDSN 构造 SQLite DSN，附加 WAL 模式和 busy_timeout
// 支持两种路径格式：相对路径（data/db.sqlite）和 file: 前缀（file:data/db.sqlite）
func buildDSN(dbPath string) string {
	pragma := "_journal_mode=WAL&_busy_timeout=5000"
	if strings.HasPrefix(dbPath, "file:") {
		// file: 格式：插入 query 参数
		if strings.Contains(dbPath, "?") {
			return dbPath + "&" + pragma
		}
		return dbPath + "?" + pragma
	}
	// 普通路径：加 file: 前缀
	return fmt.Sprintf("file:%s?%s", dbPath, pragma)
}

type zapLogger struct {
	log *zap.Logger
}

func (l *zapLogger) Printf(format string, args ...interface{}) {
	l.log.Sugar().Infof(format, args...)
}

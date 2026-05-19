package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger 创建并配置一个 zap.Logger 实例。
//
// 参数说明：
//   - level: 日志级别（debug/info/warn/error/fatal）
//   - format: 输出格式（"console" 为控制台格式，其他为 JSON 格式）
//   - output: 输出目标（"stdout"/"stderr"/"file"）
//   - logFile: 日志文件路径，仅在 output="file" 时生效
//   - errorFile: 错误日志文件路径，仅在 output="file" 时生效，仅记录 ERROR 及以上级别
//
// JSON 格式的时间字段为 Unix 时间戳，控制台格式为 "2006-01-02 15:04:05"。
// 当 output="file" 时，日志会同时输出到文件和控制台，并启用日志轮转：
//   - 单文件最大 100MB
//   - 保留最近 30 天的日志
//   - 最多保留 30 个备份文件
func NewLogger(level, format, output, logFile, errorFile string) (*zap.Logger, error) {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	var encoder zapcore.Encoder
	if format == "console" {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
		encoderCfg.TimeKey = "time"
		encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
		encoderCfg.EncodeDuration = zapcore.StringDurationEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	var cores []zapcore.Core

	switch output {
	case "stdout":
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl))
	case "stderr":
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stderr), lvl))
	case "file":
		if logFile != "" {
			fileWriter := &lumberjack.Logger{
				Filename:   logFile,
				MaxSize:    100,
				MaxBackups: 30,
				MaxAge:     30,
				Compress:   false,
				LocalTime:  true,
			}
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(fileWriter), lvl))
			cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl))
		}
	default:
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl))
	}

	if errorFile != "" && output == "file" {
		errorWriter := &lumberjack.Logger{
			Filename:   errorFile,
			MaxSize:    100,
			MaxBackups: 30,
			MaxAge:     30,
			Compress:   false,
			LocalTime:  true,
		}
		errorCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(errorWriter),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			}),
		)
		cores = append(cores, errorCore)
	}

	core := zapcore.NewTee(cores...)
	return zap.New(core, zap.AddCaller()), nil
}

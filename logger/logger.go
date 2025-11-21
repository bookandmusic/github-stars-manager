package logger

import (
	"github-stars-manager/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewDevelopmentLogger 创建开发环境的日志记录器
func NewDevelopmentLogger(cfg *config.Config) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	
	// 根据配置设置日志级别
	level := zap.DebugLevel
	switch cfg.LoggerLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	case "panic":
		level = zap.PanicLevel
	}
	
	config.Level = zap.NewAtomicLevelAt(level)
	logger, _ := config.Build()
	return logger
}

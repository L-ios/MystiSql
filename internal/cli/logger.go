package cli

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// InitLogger 初始化日志系统
// verbose: 是否启用详细日志
func InitLogger(verbose bool) error {
	// 配置日志级别
	var level zapcore.Level
	if verbose {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建 core
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	// 创建 logger
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()

	return nil
}

// GetLogger 获取 zap.Logger
func GetLogger() *zap.Logger {
	return logger
}

// GetSugar 获取 zap.SugaredLogger
func GetSugar() *zap.SugaredLogger {
	return sugar
}

// Sync 刷新日志缓冲
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

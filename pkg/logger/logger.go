package logger

import (
	"context"
	"os"
	"path/filepath"

	"github.com/yunhanshu-net/function-server/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.Logger

// contextKey 是上下文键类型
type contextKey string

// ContextKey 日志上下文键
const ContextKey contextKey = "trace_id"

// Init 初始化日志
func Init(cfg config.LogConfig) error {
	// 确保日志目录存在
	logDir := filepath.Dir("logs/api-server.log")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}
	}

	// 设置日志级别
	var level zapcore.Level
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置日志输出
	hook := lumberjack.Logger{
		Filename:   cfg.Filename,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// 创建编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建Core
	fileWriter := zapcore.AddSync(&hook)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 开发模式下同时输出到控制台和文件
	var core zapcore.Core
	if cfg.Level == "debug" {
		core = zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				fileWriter,
				level,
			),
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				consoleWriter,
				level,
			),
		)
	} else {
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			level,
		)
	}

	// 创建Logger
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

// getLoggerWithContext 从上下文中获取带有trace_id的日志记录器
func getLoggerWithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return log
	}

	if traceID, ok := ctx.Value(ContextKey).(string); ok && traceID != "" {
		return log.With(zap.String("trace_id", traceID))
	}

	return log
}

// Debug 输出Debug级别日志
func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerWithContext(ctx).Debug(msg, fields...)
}

// Debugf 输出格式化的Debug级别日志
func Debugf(ctx context.Context, format string, args ...interface{}) {
	getLoggerWithContext(ctx).Sugar().Debugf(format, args...)
}

// Info 输出Info级别日志
func Info(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerWithContext(ctx).Info(msg, fields...)
}

// Infof 输出格式化的Info级别日志
func Infof(ctx context.Context, format string, args ...interface{}) {
	getLoggerWithContext(ctx).Sugar().Infof(format, args...)
}

// Warn 输出Warn级别日志
func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerWithContext(ctx).Warn(msg, fields...)
}

// Warnf 输出格式化的Warn级别日志
func Warnf(ctx context.Context, format string, args ...interface{}) {
	getLoggerWithContext(ctx).Sugar().Warnf(format, args...)
}

// Error 输出Error级别日志
func Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	getLoggerWithContext(ctx).Error(msg, fields...)
}

// Errorf 输出格式化的Error级别日志
func Errorf(ctx context.Context, format string, args ...interface{}) {
	getLoggerWithContext(ctx).Sugar().Errorf(format, args...)
}

// Fatal 输出Fatal级别日志并退出程序
func Fatal(ctx context.Context, msg string, err error, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))
	}
	getLoggerWithContext(ctx).Fatal(msg, fields...)
}

// Fatalf 输出格式化的Fatal级别日志并退出程序
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	getLoggerWithContext(ctx).Sugar().Fatalf(format, args...)
}

// With 创建带有指定字段的新日志记录器
func With(fields ...zap.Field) *zap.Logger {
	return log.With(fields...)
}

// WithContext 向上下文中添加trace_id
func WithContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextKey, traceID)
}

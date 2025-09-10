package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel 定义日志级别类型
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogConfig 日志配置
type LogConfig struct {
	Level string `yaml:"level"` // 日志级别
}

// DefaultLogConfig 默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level: "info",
	}
}

// Logger 统一日志接口
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
	config    *LogConfig
}

// NewLogger 创建新的日志实例
func NewLogger(config *LogConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	switch config.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 使用控制台编码器
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 输出到标准输出
	writeSyncer := zapcore.AddSync(os.Stdout)

	// 创建核心
	core := zapcore.NewCore(encoder, writeSyncer, zap.NewAtomicLevelAt(level))

	// 创建logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := zapLogger.Sugar()

	return &Logger{
		zapLogger: zapLogger,
		sugar:     sugar,
		config:    config,
	}, nil
}

// Debug 调试日志
func (l *Logger) Debug(args ...interface{}) {
	l.sugar.Debug(args...)
}

// Debugf 格式化调试日志
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Info 信息日志
func (l *Logger) Info(args ...interface{}) {
	l.sugar.Info(args...)
}

// Infof 格式化信息日志
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warn 警告日志
func (l *Logger) Warn(args ...interface{}) {
	l.sugar.Warn(args...)
}

// Warnf 格式化警告日志
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Error 错误日志
func (l *Logger) Error(args ...interface{}) {
	l.sugar.Error(args...)
}

// Errorf 格式化错误日志
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(args ...interface{}) {
	l.sugar.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// Printf 兼容fmt.Printf的接口，用于用户交互输出
func (l *Logger) Printf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// Print 兼容fmt.Print的接口，用于用户交互输出
func (l *Logger) Print(args ...interface{}) {
	fmt.Print(args...)
}

// Println 兼容fmt.Println的接口，用于用户交互输出
func (l *Logger) Println(args ...interface{}) {
	fmt.Println(args...)
}

// Sync 同步日志缓冲区
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Close 关闭日志
func (l *Logger) Close() error {
	return l.zapLogger.Sync()
}

// 全局日志实例
var defaultLogger *Logger

// InitDefaultLogger 初始化默认日志实例
func InitDefaultLogger(config *LogConfig) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetDefaultLogger 获取默认日志实例
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger, _ = NewLogger(DefaultLogConfig())
	}
	return defaultLogger
}

// 全局便捷函数
func Debug(args ...interface{}) {
	GetDefaultLogger().Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	GetDefaultLogger().Debugf(template, args...)
}

func Info(args ...interface{}) {
	GetDefaultLogger().Info(args...)
}

func Infof(template string, args ...interface{}) {
	GetDefaultLogger().Infof(template, args...)
}

func Warn(args ...interface{}) {
	GetDefaultLogger().Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	GetDefaultLogger().Warnf(template, args...)
}

func Error(args ...interface{}) {
	GetDefaultLogger().Error(args...)
}

func Errorf(template string, args ...interface{}) {
	GetDefaultLogger().Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	GetDefaultLogger().Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	GetDefaultLogger().Fatalf(template, args...)
}

// Printf 兼容fmt.Printf的全局函数
func Printf(format string, args ...interface{}) {
	GetDefaultLogger().Printf(format, args...)
}

// Print 兼容fmt.Print的全局函数
func Print(args ...interface{}) {
	GetDefaultLogger().Print(args...)
}

// Println 兼容fmt.Println的全局函数
func Println(args ...interface{}) {
	GetDefaultLogger().Println(args...)
}

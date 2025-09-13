package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel defines log level types
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
)

// LogConfig logger configuration
type LogConfig struct {
	Level string `yaml:"level"` // log level
}

// DefaultLogConfig default logger configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level: "info",
	}
}

// Logger unified logging interface
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
	config    *LogConfig
}

// NewLogger creates a new logger instance
func NewLogger(config *LogConfig) (*Logger, error) {
	if config == nil {
		config = DefaultLogConfig()
	}

	// Set log level
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

	// Create encoder configuration
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

	// Use console encoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Output to stdout
	writeSyncer := zapcore.AddSync(os.Stdout)

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, zap.NewAtomicLevelAt(level))

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar := zapLogger.Sugar()

	return &Logger{
		zapLogger: zapLogger,
		sugar:     sugar,
		config:    config,
	}, nil
}

// Debug debug log
func (l *Logger) Debug(args ...interface{}) {
	l.sugar.Debug(args...)
}

// Debugf formatted debug log
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Info info log
func (l *Logger) Info(args ...interface{}) {
	l.sugar.Info(args...)
}

// Infof formatted info log
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warn warning log
func (l *Logger) Warn(args ...interface{}) {
	l.sugar.Warn(args...)
}

// Warnf formatted warning log
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Error error log
func (l *Logger) Error(args ...interface{}) {
	l.sugar.Error(args...)
}

// Errorf formatted error log
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatal fatal error log
func (l *Logger) Fatal(args ...interface{}) {
	l.sugar.Fatal(args...)
}

// Fatalf formatted fatal error log
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// Printf compatible interface with fmt.Printf for user interaction output
// Deprecated: Use UserOutput methods instead for better output management
func (l *Logger) Printf(format string, args ...interface{}) {
	GetDefaultUserOutput().Printf(format, args...)
}

// Print compatible interface with fmt.Print for user interaction output
// Deprecated: Use UserOutput methods instead for better output management
func (l *Logger) Print(args ...interface{}) {
	GetDefaultUserOutput().Print(args...)
}

// Println compatible interface with fmt.Println for user interaction output
// Deprecated: Use UserOutput methods instead for better output management
func (l *Logger) Println(args ...interface{}) {
	GetDefaultUserOutput().Println(args...)
}

// Sync log buffer
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Close logger
func (l *Logger) Close() error {
	return l.zapLogger.Sync()
}

// Global logger instance
var defaultLogger *Logger

// InitDefaultLogger initializes default logger instance
func InitDefaultLogger(config *LogConfig) error {
	logger, err := NewLogger(config)
	if err != nil {
		return err
	}
	defaultLogger = logger
	return nil
}

// GetDefaultLogger gets default logger instance
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger, _ = NewLogger(DefaultLogConfig())
	}
	return defaultLogger
}

// Global convenience functions
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

// Printf global function compatible with fmt.Printf
// Deprecated: Use UserOutput methods instead for better output management
func Printf(format string, args ...interface{}) {
	GetDefaultUserOutput().Printf(format, args...)
}

// Print global function compatible with fmt.Print
// Deprecated: Use UserOutput methods instead for better output management
func Print(args ...interface{}) {
	GetDefaultUserOutput().Print(args...)
}

// Println global function compatible with fmt.Println
// Deprecated: Use UserOutput methods instead for better output management
func Println(args ...interface{}) {
	GetDefaultUserOutput().Println(args...)
}

package log

import (
	"os"

	config "astron-xmod-shim/internal/dto/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	sugarLogger  *zap.SugaredLogger
)

// Init Initialize logging system based on configuration
func Init(cfg *config.LogConfig) error {
	// 1. Validate and process default configuration values
	if err := setDefaultConfig(cfg); err != nil {
		return err
	}

	// 2. Build log core (stdout only)
	core := buildLogCore(cfg)

	// 3. Configure log options (caller line number, etc.)
	options := buildZapOptions(cfg)

	// 4. Initialize global Logger
	globalLogger = zap.New(core, options...)
	sugarLogger = globalLogger.Sugar()

	return nil
}

// setDefaultConfig Set default values for missing configuration items
func setDefaultConfig(cfg *config.LogConfig) error {
	if cfg.Level == "" {
		cfg.Level = "info" // Default to info level
	}
	// No need for file path, MaxSize, MaxAge in cloud-native environment
	return nil
}

// buildLogCore Build log core: stdout only, using JSON encoding
func buildLogCore(cfg *config.LogConfig) zapcore.Core {
	// Use stdout as output target
	consoleSyncer := zapcore.Lock(zapcore.AddSync(zapcore.Lock(os.Stdout)))

	// Log encoding configuration (JSON structured logging)
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // Lowercase level (info/warn/error)
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 time format
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // Short caller format (file.go:line)
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel // Default to info for invalid levels
	}

	// Build core: stdout only
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // JSON format
		consoleSyncer,
		level,
	)
}

// buildZapOptions Build zap options (caller line number, etc.)
func buildZapOptions(cfg *config.LogConfig) []zap.Option {
	var options []zap.Option
	if cfg.ShowLine {
		options = append(options, zap.AddCaller())      // Show caller information
		options = append(options, zap.AddCallerSkip(1)) // Skip current package level
	}
	return options
}

// Common logging method wrappers (SugaredLogger)

func Debug(template string, args ...interface{}) {
	sugarLogger.Debugf(template, args...)
}

func Info(template string, args ...interface{}) {
	sugarLogger.Infof(template, args...)
}

func Warn(template string, args ...interface{}) {
	sugarLogger.Warnf(template, args...)
}

func Error(template string, args ...interface{}) {
	sugarLogger.Errorf(template, args...)
}

func Fatal(template string, args ...interface{}) {
	sugarLogger.Fatalf(template, args...)
}

// Sync Flush log buffer (call before program exit)
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}
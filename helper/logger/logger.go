package logger

import (
	"context"
	"os"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aluo/gomono/edgecom/config"
)

// For mapping config logger to app logger levels
var logLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

var once sync.Once

var (
	instance *logger
)

type logger struct {
	logger  *zap.Logger
	logger2 *zap.Logger
}

func NewLogger(cfg *config.Config) *zap.Logger {
	once.Do(func() { // <-- atomic, does not allow repeating

		zapLogger := initLogger(cfg) // <-- thread safe
		instance = &logger{
			logger:  zapLogger,
			logger2: zapLogger.WithOptions(zap.AddCaller(), zap.AddCallerSkip(-1)),
		}
	})

	return instance.logger
}

func getLoggerLevel(cfg *config.Config) zapcore.Level {
	level, exist := logLevelMap[cfg.Logger.Level]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}

// InitLogger is to initiate logger
func initLogger(cfg *config.Config) *zap.Logger {
	logLevel := getLoggerLevel(cfg)

	logWriter := zapcore.AddSync(os.Stderr)

	var encoderCfg zapcore.EncoderConfig
	if cfg.Server.Mode == "Development" {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	var encoder zapcore.Encoder
	encoderCfg.LevelKey = "LEVEL"
	encoderCfg.CallerKey = "CALLER"
	encoderCfg.TimeKey = "TIME"
	encoderCfg.NameKey = "NAME"
	encoderCfg.MessageKey = "MESSAGE"

	if cfg.Logger.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger
}

func SyncLogger(logger *zap.Logger) {
	if err := logger.Sync(); err != nil {
		logger.Error(err.Error())
	}
}

func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	ctxzap.Debug(ctx, msg, fields...)
}

func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	ctxzap.Info(ctx, msg, fields...)
}

func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	ctxzap.Warn(ctx, msg, fields...)
}

func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	ctxzap.Error(ctx, msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	instance.logger2.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	instance.logger2.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	instance.logger2.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	instance.logger2.Error(msg, fields...)
}

package logger

import (
	"context"

	"github.com/presnalex/go-micro/v3/wrapper/requestid"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	zaplogger "go.unistack.org/micro-logger-zap/v3"
	"go.unistack.org/micro/v3/logger"
)

const (
	LoggerField = "x-request-id"
)

type LoggerConfig struct {
	LogLevel string `json:"loglevel"`
}

func DefaultLogger(cfg *LoggerConfig) logger.Logger {
	level := logger.InfoLevel
	if cfg != nil {
		level = logger.ParseLevel(cfg.LogLevel)
	}
	enccfg := zap.NewProductionEncoderConfig()
	enccfg.EncodeTime = zapcore.ISO8601TimeEncoder
	enccfg.TimeKey = "@timestamp"
	zapcfg := zap.NewProductionConfig()

	l := zaplogger.NewLogger(
		zaplogger.WithCallerSkip(2), // hardcode 2 skip to use with helper
		zaplogger.WithConfig(zapcfg),
		zaplogger.WithEncoderConfig(enccfg),
		logger.WithLevel(level),
	)

	if err := l.Init(); err != nil {
		panic(err)
	}

	logger.DefaultLogger = l

	return l
}

type LoggerKey struct{}

func FromOutgoingContext(ctx context.Context) logger.Logger {
	if l, ok := ctx.Value(LoggerKey{}).(logger.Logger); ok {
		return l
	}
	reqid, ok := requestid.GetOutgoingRequestId(ctx)
	if !ok {
		uid, err := uuid.NewRandom()
		if err != nil {
			uid = uuid.Nil
		}
		reqid = uid.String()
	}
	return logger.DefaultLogger.Fields(map[string]interface{}{LoggerField: reqid})
}

func FromIncomingContext(ctx context.Context) logger.Logger {
	if l, ok := ctx.Value(LoggerKey{}).(logger.Logger); ok {
		return l
	}
	reqid, ok := requestid.GetIncomingRequestId(ctx)
	if !ok {
		uid, err := uuid.NewRandom()
		if err != nil {
			uid = uuid.Nil
		}
		reqid = uid.String()
	}
	return logger.DefaultLogger.Fields(map[string]interface{}{LoggerField: reqid})
}

func InjectLogger(ctx context.Context, reqid string) context.Context {
	fieldHelper := logger.DefaultLogger.Fields(map[string]interface{}{LoggerField: reqid})
	lCtx := context.WithValue(ctx, LoggerKey{}, fieldHelper)
	return context.WithValue(lCtx, LoggerKey{}, fieldHelper)
}

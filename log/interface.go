package log

import (
	"context"

	"github.com/Hyingerrr/mirco-esim/config"

	"go.uber.org/zap"
)

var _log Logger

func NewNullLogger() Logger {
	opt := LoggerOptions{}
	return NewLogger(opt.WithLoggerConf(config.NewNullConfig()))
}

type Logger interface {
	Error(msg string)

	Debugf(string, ...interface{})

	Infof(string, ...interface{})

	Info(...interface{})

	InfoW(string, ...interface{})

	Warnf(string, ...interface{})

	WarnW(string, ...interface{})

	Errorf(string, ...interface{})

	ErrorW(string, ...interface{})

	//Errorfo(string, ...zapcore.Field)

	DPanicf(string, ...interface{})

	Panicf(string, ...interface{})

	Fatalf(string, ...interface{})

	Debugc(context.Context, string, ...interface{})

	Infoc(context.Context, string, ...interface{})

	Warnc(context.Context, string, ...interface{})

	Errorc(context.Context, string, ...interface{})

	DPanicc(context.Context, string, ...interface{})

	Panicc(context.Context, string, ...interface{})

	Fatalc(context.Context, string, ...interface{})

	WithFields(Field) *zap.SugaredLogger
}

func Error(msg string) {
	_log.Error(msg)
}

func Errorc(ctx context.Context, format string, args ...interface{}) {
	_log.Errorc(ctx, format, args...)
}

func Errorf(format string, args ...interface{}) {
	_log.Errorf(format, args...)
}

func ErrorW(format string, args ...interface{}) {
	_log.ErrorW(format, args...)
}

func Debugf(format string, args ...interface{}) {
	_log.Debugf(format, args...)
}

func Debugc(ctx context.Context, format string, args ...interface{}) {
	_log.Debugc(ctx, format, args...)
}

func Infof(format string, args ...interface{}) {
	_log.Infof(format, args...)
}

func Info(args ...interface{}) {
	_log.Info(args...)
}

func InfoW(format string, args ...interface{}) {
	_log.InfoW(format, args...)
}

func Infoc(ctx context.Context, format string, args ...interface{}) {
	_log.Infoc(ctx, format, args...)
}

func Warnf(format string, args ...interface{}) {
	_log.Warnf(format, args...)
}

func Warnc(ctx context.Context, format string, args ...interface{}) {
	_log.Warnc(ctx, format, args...)
}

func DPanicf(format string, args ...interface{}) {
	_log.DPanicf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	_log.Panicf(format, args...)
}

func DPanicc(ctx context.Context, format string, args ...interface{}) {
	_log.DPanicc(ctx, format, args...)
}

func Panicc(ctx context.Context, format string, args ...interface{}) {
	_log.Panicc(ctx, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	_log.Fatalf(format, args...)
}

func Fatalc(ctx context.Context, format string, args ...interface{}) {
	_log.Fatalc(ctx, format, args...)
}

func WithFields(field Field) *zap.SugaredLogger {
	return _log.WithFields(field)
}

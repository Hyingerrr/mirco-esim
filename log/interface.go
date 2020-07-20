package log

import (
	"context"
	"github.com/jukylin/esim/config"
	"go.uber.org/zap"
)

var (
	Log Logger
)

func NewNullLogger() Logger {
	opt := LoggerOptions{}
	Log = NewLogger(opt.WithLoggerConf(config.NewNullConfig()))
	return Log
}

type Logger interface {
	Error(msg string)

	Debugf(string, ...interface{})

	Infof(string, ...interface{})

	Info( ...interface{})

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

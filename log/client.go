package log

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/jukylin/esim/config"
	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type logger struct {
	*Config
	log   *zap.Logger
	debug bool
	sugar *zap.SugaredLogger
	conf  config.Config
}

type Field map[string]interface{}

type LoggerOptions struct{}

type Option func(*logger)

func NewLogger(options ...Option) Logger {
	var (
		writer []zapcore.WriteSyncer
		l      = &logger{Config: &Config{}}
		opts   = make([]zap.Option, 0)
		core   []zapcore.Core
	)

	for _, opt := range options {
		opt(l)
	}

	if l.conf == nil {
		l.conf = config.NewMemConfig()
	}

	l.Config.fillWithDefaultConfig(l.conf)

	hook := &lumberjack.Logger{
		Filename:   l.Config.File,
		MaxSize:    l.Config.MaxSize,
		MaxBackups: l.Config.BackupCount,
		MaxAge:     l.Config.MaxAge,
		Compress:   false,
	}

	switch {
	case l.Config.IsBothFileStdout():
		writer = append(writer, zapcore.AddSync(hook), zapcore.AddSync(os.Stdout))
	case l.Config.IsOutFile():
		writer = append(writer, zapcore.AddSync(hook))
	case l.Config.IsOutStdout():
		writer = append(writer, zapcore.AddSync(os.Stdout))
	}

	lever := ParseLevel(l.Config.Level)

	if l.Config.ReportCaller {
		opts = append(opts, zap.AddCaller())
	}

	if l.Config.Stacktrace {
		opts = append(opts, zap.AddStacktrace(lever))
	}

	if l.Config.ErrStats {
		opts = append(opts, zap.Hooks(addErrMetric))
	}

	for _, w := range writer {
		core = append(core, zapcore.NewCore(l.buildEncoder(), w, zap.NewAtomicLevelAt(lever)))
	}

	l.log = zap.New(zapcore.NewTee(core...), opts...)
	l.sugar = l.log.Sugar()
	Log = l
	return Log
}

func (LoggerOptions) WithLoggerConf(conf config.Config) Option {
	return func(l *logger) {
		l.conf = conf
	}
}

func (LoggerOptions) WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

func (l *logger) buildEncoder() zapcore.Encoder {
	var (
		encoder zapcore.EncoderConfig
	)

	if l.debug {
		encoder = zap.NewDevelopmentEncoderConfig()
	} else {
		encoder = zap.NewProductionEncoderConfig()
	}

	encoder.TimeKey = "time"
	encoder.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.999999"))
	}
	encoder.EncodeCaller = zapcore.FullCallerEncoder
	encoder.EncodeName = zapcore.FullNameEncoder

	if l.Config.Format == "json" {
		return zapcore.NewJSONEncoder(encoder)
	} else {
		return zapcore.NewConsoleEncoder(encoder)
	}
}

//func (l *logger) Errorfo(msg string, field ...zapcore.Field) {
//	l.logger.Error(msg, field...)
//}

func (l *logger) Infof(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Infof(template, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Info(args...)
}

func (l *logger) Infoc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).Infof(template, args...)
}

func (l *logger) InfoW(msg string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Infow(msg, args...)
}

func (l *logger) Warnf(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Warnf(template, args...)
}

func (l *logger) WarnW(msg string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Warnw(msg, args...)
}

func (l *logger) Warnc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).Warnf(template, args...)
}

func (l *logger) Error(msg string) {
	l.log.Error(msg)
}

func (l *logger) Errorf(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Errorf(template, args...)
}

func (l *logger) Errorc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).Errorf(template, args...)
}

func (l *logger) ErrorW(msg string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Errorw(msg, args...)
}

func (l *logger) Debugf(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Debugf(template, args...)
}

func (l *logger) DPanicf(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).DPanicf(template, args...)
}

func (l *logger) Panicf(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Panicf(template, args...)
}

func (l *logger) Debugc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).Debugf(template, args...)
}

func (l *logger) Fatalf(template string, args ...interface{}) {
	l.sugar.With(l.getArgs(context.TODO())...).Fatalf(template, args...)
}

func (l *logger) DPanicc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).DPanicf(template, args...)
}

func (l *logger) Panicc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).Panicf(template, args...)
}

func (l *logger) Fatalc(ctx context.Context, template string, args ...interface{}) {
	l.sugar.With(l.getArgs(ctx)...).Fatalf(template, args...)
}

func (l *logger) standardTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func (l *logger) WithFields(field Field) *zap.SugaredLogger {
	return l.withFields(context.TODO(), field)
}

func (l *logger) withFields(ctx context.Context, field Field) *zap.SugaredLogger {
	return l.sugar.With(l.getArgs(ctx, field)...)
}

func (l *logger) getArgs(ctx context.Context, field ...Field) []interface{} {
	args := make([]interface{}, 0)

	args = append(args, "caller", l.getCaller(runtime.Caller(2)))

	if len(field) > 0 {
		for k, v := range field[0] {
			args = append(args, k, v)
		}
	}

	tracerID := l.getTracerID(ctx)
	if tracerID != "" {
		args = append(args, "tracer_id", tracerID)
	}

	return args
}

func (l *logger) getCaller(pc uintptr, file string, line int, ok bool) string {
	return zapcore.NewEntryCaller(pc, file, line, ok).TrimmedPath()
}

// getTracerID get tracer_id from context.
func (l *logger) getTracerID(ctx context.Context) string {
	sp := opentracing.SpanFromContext(ctx)
	if sp != nil {
		if jaegerSpanContext, ok := sp.Context().(jaeger.SpanContext); ok {
			return jaegerSpanContext.TraceID().String()
		}
	}

	val := ctx.Value(tracerid.ActiveEsimKey)
	if tracerID, ok := val.(string); ok {
		return tracerID
	}

	return ""
}

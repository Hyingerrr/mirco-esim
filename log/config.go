package log

import (
	"strings"

	"github.com/Hyingerrr/mirco-esim/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Output       string // 日志的位置 file|both|stdout
	Level        string
	Format       string
	ReportCaller bool
	Stacktrace   bool
	ErrStats     bool
	File         string
	MaxSize      int  // 单个文件最大size
	MaxAge       int  // 保留旧文件的最大天数
	BackupCount  int  // 保留旧文件的最大个数
	Compress     bool // 是否压缩/归档旧文件
}

func (c *Config) fillWithDefaultConfig(conf config.Config) {
	c.Output = conf.GetString("log_output")
	if c.Output == "" {
		c.Output = "stdout"
	}

	c.File = conf.GetString("log_file")
	if c.File == "" {
		c.File = "./logs/esim.log"
	}

	c.Level = conf.GetString("log_level")
	if c.Level == "" {
		c.Level = "debug"
	}

	c.Compress = conf.GetBool("log_compress")
	c.BackupCount = conf.GetInt("log_backup_count")
	if c.BackupCount == 0 {
		c.BackupCount = 20
	}
	c.MaxAge = conf.GetInt("log_max_age")
	if c.MaxAge == 0 {
		c.MaxAge = 15
	}
	c.MaxSize = conf.GetInt("log_max_size")
	if c.MaxSize == 0 {
		c.MaxSize = 1000 // 10G
	}
	c.Stacktrace = conf.GetBool("log_stack_trace")
	c.ReportCaller = conf.GetBool("log_report_caller")
	c.Format = conf.GetString("log_format")
	if c.Format == "" {
		c.Format = "text"
	}

	c.ErrStats = conf.GetBool("log_err_stats")
}

func (c *Config) IsOutStdout() bool {
	return c.Output == "stdout"
}

func (c *Config) IsBothFileStdout() bool {
	return c.Output == "both"
}

func (c *Config) IsOutFile() bool {
	return c.Output == "file"
}

func ParseLevel(lvl string) zapcore.Level {
	switch strings.ToLower(lvl) {
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	case "error":
		return zap.ErrorLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	default:
		return zap.InfoLevel
	}
}

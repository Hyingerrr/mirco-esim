package container

import (
	"github.com/Hyingerrr/mirco-esim/config"
	logx "github.com/Hyingerrr/mirco-esim/log"
)

var (
	loggerFunc = func(conf config.Config) logx.Logger {
		var loggerOptions logx.LoggerOptions
		logger := logx.NewLogger(
			loggerOptions.WithLoggerConf(conf),
			loggerOptions.WithDebug(conf.GetBool("debug")),
		)
		return logger
	}

	confFunc = func() config.Config {
		return config.NewMemConfig()
	}
)

func SetConfFunc(conf func() config.Config) {
	confFunc = conf
}

func SetLogger(logger func(config.Config) logx.Logger) {
	loggerFunc = logger
}

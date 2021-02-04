package mysql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/jukylin/esim/container"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jinzhu/gorm"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/proxy"
)

var clientOnce sync.Once

var onceClient *Client

type Client struct {
	gdbs map[string]*gorm.DB

	sqlDbs map[string]*sql.DB

	proxy []func() interface{}

	conf config.Config

	logger log.Logger

	dbConfigs []DbConfig

	closeChan chan bool

	stateTicker time.Duration

	// for integration tests
	db *sql.DB

	traceOnce sync.Once

	lock sync.Mutex
}

type Option func(c *Client)

type ClientOptions struct{}

type DbConfig struct {
	Db          string `json:"db" yaml:"db"`
	Dsn         string `json:"dsn" yaml:"dsn"`
	MaxIdle     int    `json:"max_idle" yaml:"maxidle"`
	MaxOpen     int    `json:"max_open" yaml:"maxopen"`
	MaxLifetime int    `json:"max_lifetime" yaml:"maxlifetime"`
}

func NewClient(options ...Option) *Client {
	clientOnce.Do(func() {
		onceClient = &Client{
			gdbs:        make(map[string]*gorm.DB),
			sqlDbs:      make(map[string]*sql.DB),
			proxy:       make([]func() interface{}, 0),
			stateTicker: 10 * time.Second,
			closeChan:   make(chan bool, 1),
		}

		for _, option := range options {
			option(onceClient)
		}

		if onceClient.conf == nil {
			onceClient.conf = config.NewNullConfig()
		}

		if onceClient.logger == nil {
			onceClient.logger = log.NewLogger()
		}

		onceClient.init()
	})

	return onceClient
}

func (ClientOptions) WithConf(conf config.Config) Option {
	return func(m *Client) {
		m.conf = conf
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(m *Client) {
		m.logger = logger
	}
}

func (ClientOptions) WithDbConfig(dbConfigs []DbConfig) Option {
	return func(m *Client) {
		m.dbConfigs = dbConfigs
	}
}

func (ClientOptions) WithProxy(proxys ...func() interface{}) Option {
	return func(m *Client) {
		m.proxy = append(m.proxy, proxys...)
	}
}

func (ClientOptions) WithStateTicker(stateTicker time.Duration) Option {
	return func(m *Client) {
		m.stateTicker = stateTicker
	}
}

func (ClientOptions) WithDB(db *sql.DB) Option {
	return func(m *Client) {
		m.db = db
	}
}

// initializes Client.
func (c *Client) init() {
	dbConfigs := make([]DbConfig, 0)
	err := c.conf.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		c.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(c.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, c.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		if len(c.proxy) == 0 {
			var DB *gorm.DB

			if c.db != nil {
				DB, err = gorm.Open("mysql", c.db)
			} else {
				DB, err = gorm.Open("mysql", dbConfig.Dsn)
			}

			if err != nil {
				c.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			DB.DB().SetMaxIdleConns(dbConfig.MaxIdle)
			DB.DB().SetMaxOpenConns(dbConfig.MaxOpen)
			DB.DB().SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))

			c.setDb(dbConfig.Db, DB, DB.DB())

			if c.conf.GetBool("debug") {
				DB.LogMode(true)
			}
		} else {
			var DB *gorm.DB
			var dbSQL *sql.DB

			if c.db == nil {
				dbSQL, err = sql.Open("mysql", dbConfig.Dsn)
				if err != nil {
					c.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
				}
			} else {
				dbSQL = c.db
			}

			firstProxy := proxy.NewProxyFactory().
				GetFirstInstance("db_"+dbConfig.Db, dbSQL, c.proxy...)

			DB, err = gorm.Open("mysql", firstProxy)
			if err != nil {
				c.logger.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			err = dbSQL.Ping()
			if err != nil {
				c.logger.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			dbSQL.SetMaxIdleConns(dbConfig.MaxIdle)
			dbSQL.SetMaxOpenConns(dbConfig.MaxOpen)
			dbSQL.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))

			c.setDb(dbConfig.Db, DB, dbSQL)

			if c.conf.GetBool("debug") {
				DB.LogMode(true)
			}
		}

		go c.Stats()
		c.logger.Infof("[mysql] %s init success", dbConfig.Db)
	}
}

func (c *Client) setDb(dbName string, gdb *gorm.DB, db *sql.DB) {
	c.lock.Lock()
	defer c.lock.Unlock()
	dbName = strings.ToLower(dbName)

	c.gdbs[dbName] = gdb
	c.sqlDbs[dbName] = db
}

func (c *Client) GetDb(dbName string) *gorm.DB {
	return c.getDb(context.Background(), dbName)
}

func (c *Client) getDb(ctx context.Context, dbName string) *gorm.DB {
	if db, ok := c.gdbs[strings.ToLower(dbName)]; ok {
		// inject monitor
		if c.conf.GetBool("mysql_metric") {
			c.RegisterMetricsCallbacks(ctx, db)
		}

		// inject tracer
		if c.conf.GetBool("mysql_tracer") {
			db = c.Trace(ctx, db).DB
		}

		return db
	}

	c.logger.Errorc(ctx, "[db] %s not found", dbName)

	return nil
}

func (c *Client) GetCtxDb(ctx context.Context, dbName string) *gorm.DB {
	return c.getDb(ctx, dbName)
}

func (c *Client) Ping() []error {
	var errs []error
	var err error
	for _, db := range c.sqlDbs {
		err = db.Ping()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *Client) Close() {
	var err error
	for _, db := range c.gdbs {
		err = db.Close()
		if err != nil {
			c.logger.Errorf(err.Error())
		}
	}
}

func (c *Client) Stats() {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Infof(err.(error).Error())
		}
	}()

	ticker := time.NewTicker(c.stateTicker)
	var stats sql.DBStats

	for {
		select {
		case <-ticker.C:
			serviceName := container.AppName()
			for dbName, db := range c.sqlDbs {
				stats = db.Stats()
				dbSchema := c.gdbs[dbName].Scopes().Dialect().CurrentDatabase()
				mysqlDBStats.Set(float64(stats.MaxOpenConnections), serviceName, dbSchema, "max_open_conn")
				mysqlDBStats.Set(float64(stats.OpenConnections), serviceName, dbSchema, "open_conn")
				mysqlDBStats.Set(float64(stats.InUse), serviceName, dbSchema, "in_use")
				mysqlDBStats.Set(float64(stats.Idle), serviceName, dbSchema, "idle")
				mysqlDBStats.Set(float64(stats.WaitCount), serviceName, dbSchema, "wait_count")
			}
		case <-c.closeChan:
			c.logger.Infof("stop stats")
			goto Stop
		}
	}

Stop:
	ticker.Stop()
	return
}

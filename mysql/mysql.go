package mysql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/Hyingerrr/mirco-esim/config"
	"github.com/Hyingerrr/mirco-esim/container"
	logx "github.com/Hyingerrr/mirco-esim/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var clientOnce sync.Once

var onceClient *Client

type Client struct {
	gdbs map[string]*gorm.DB

	sqlDbs map[string]*sql.DB

	dbConfigs []DbConfig

	closeChan chan bool

	stateTicker time.Duration

	db *sql.DB

	traceOnce sync.Once

	lock sync.Mutex
}

type Option func(c *Client)

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
			stateTicker: 10 * time.Second,
			closeChan:   make(chan bool, 1),
		}

		for _, option := range options {
			option(onceClient)
		}

		onceClient.init()
	})

	return onceClient
}

func WithDbConfig(dbConfigs []DbConfig) Option {
	return func(m *Client) {
		m.dbConfigs = dbConfigs
	}
}

func WithStateTicker(stateTicker time.Duration) Option {
	return func(m *Client) {
		m.stateTicker = stateTicker
	}
}

func WithDB(db *sql.DB) Option {
	return func(m *Client) {
		m.db = db
	}
}

// initializes Client.
func (c *Client) init() {
	dbConfigs := make([]DbConfig, 0)
	err := config.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		logx.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(c.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, c.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		var DB *gorm.DB

		if c.db != nil {
			DB, err = gorm.Open("mysql", c.db)
		} else {
			DB, err = gorm.Open("mysql", dbConfig.Dsn)
		}
		if err != nil {
			logx.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
		}

		if err = DB.DB().Ping(); err != nil {
			logx.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
		}

		DB.DB().SetMaxIdleConns(dbConfig.MaxIdle)
		DB.DB().SetMaxOpenConns(dbConfig.MaxOpen)
		DB.DB().SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))

		c.setDb(dbConfig.Db, DB, DB.DB())

		if config.GetBool("debug") {
			DB.LogMode(true)
		}

		go c.Stats()
		logx.Infof("[mysql] %s init success", dbConfig.Db)
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
		if config.GetBool("mysql_metric") {
			c.RegisterMetricsCallbacks(ctx, db)
		}

		// inject tracer
		if config.GetBool("mysql_tracer") {
			db = c.Trace(ctx, db).DB
		}

		return db
	}

	logx.Errorc(ctx, "[db] %s not found", dbName)

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
			logx.Errorf(err.Error())
		}
	}
}

func (c *Client) Stats() {
	defer func() {
		if err := recover(); err != nil {
			logx.Infof(err.(error).Error())
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
			logx.Infof("stop stats")
			goto Stop
		}
	}

Stop:
	ticker.Stop()
	return
}

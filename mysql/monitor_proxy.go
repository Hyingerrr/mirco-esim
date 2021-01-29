package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
)

type MonitorProxy struct {
	// proxy name
	name string

	nextProxy SQLCommon

	tracer opentracing2.Tracer

	conf config.Config

	logger log.Logger

	afterEvents []afterEvents
}

type afterEvents func(string, time.Time, time.Time)

type MonitorProxyOption func(c *MonitorProxy)

type MonitorProxyOptions struct{}

func NewMonitorProxy(options ...MonitorProxyOption) *MonitorProxy {
	MonitorProxy := &MonitorProxy{}

	for _, option := range options {
		option(MonitorProxy)
	}

	if MonitorProxy.conf == nil {
		MonitorProxy.conf = config.NewNullConfig()
	}

	if MonitorProxy.logger == nil {
		MonitorProxy.logger = log.NewLogger()
	}

	if MonitorProxy.tracer == nil {
		MonitorProxy.tracer = opentracing.NewTracer("mysql", MonitorProxy.logger)
	}

	MonitorProxy.name = "monitor_proxy"

	MonitorProxy.registerAfterEvent()

	return MonitorProxy
}

func (MonitorProxyOptions) WithConf(conf config.Config) MonitorProxyOption {
	return func(r *MonitorProxy) {
		r.conf = conf
	}
}

func (MonitorProxyOptions) WithLogger(logger log.Logger) MonitorProxyOption {
	return func(r *MonitorProxy) {
		r.logger = logger
	}
}

func (MonitorProxyOptions) WithTracer(tracer opentracing2.Tracer) MonitorProxyOption {
	return func(r *MonitorProxy) {
		r.tracer = tracer
	}
}

// Implement Proxy interface.
func (mp *MonitorProxy) NextProxy(db interface{}) {
	mp.nextProxy = db.(SQLCommon)
}

// Implement Proxy interface.
func (mp *MonitorProxy) ProxyName() string {
	return mp.name
}

func (mp *MonitorProxy) Exec(query string, args ...interface{}) (sql.Result, error) {
	startTime := time.Now()
	result, err := mp.nextProxy.Exec(query, args...)
	mp.after(query, startTime)
	return result, err
}

func (mp *MonitorProxy) Prepare(query string) (*sql.Stmt, error) {
	startTime := time.Now()
	stmt, err := mp.nextProxy.Prepare(query)
	mp.after(query, startTime)

	return stmt, err
}

func (mp *MonitorProxy) Query(query string, args ...interface{}) (*sql.Rows, error) {
	startTime := time.Now()
	rows, err := mp.nextProxy.Query(query, args...)
	mp.after(query, startTime)

	return rows, err
}

func (mp *MonitorProxy) QueryRow(query string, args ...interface{}) *sql.Row {
	startTime := time.Now()
	row := mp.nextProxy.QueryRow(query, args...)
	mp.after(query, startTime)

	return row
}

func (mp *MonitorProxy) Close() error {
	return mp.nextProxy.Close()
}

func (mp *MonitorProxy) Begin() (*sql.Tx, error) {
	return mp.nextProxy.Begin()
}

func (mp *MonitorProxy) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return mp.nextProxy.BeginTx(ctx, opts)
}

func (mp *MonitorProxy) registerAfterEvent() {
	if mp.conf.GetBool("mysql_tracer") {
		mp.afterEvents = append(mp.afterEvents, mp.withMysqlTracer)
	}

	if mp.conf.GetBool("mysql_check_slow") {
		mp.afterEvents = append(mp.afterEvents, mp.withSlowSQL)
	}

	if mp.conf.GetBool("mysql_metrics") {
		mp.afterEvents = append(mp.afterEvents, mp.withMysqlMetrics)
	}
}

func (mp *MonitorProxy) after(query string, beginTime time.Time) {
	now := time.Now()
	for _, event := range mp.afterEvents {
		event(query, beginTime, now)
	}
}

func (mp *MonitorProxy) withSlowSQL(query string, beginTime, endTime time.Time) {
	mysqlSlowTime := mp.conf.GetInt64("mysql_slow_time")

	if mysqlSlowTime != 0 {
		if endTime.Sub(beginTime) > time.Duration(mysqlSlowTime)*time.Millisecond {
			mp.logger.Warnf("slow sql %s", query)
		}
	}
}

func (mp *MonitorProxy) withMysqlMetrics(query string, beginTime, endTime time.Time) {
	// 阿里云dms后台可以看的很详细
	// sql太多，防止metric被打爆

	//mysqlDBTotal.Inc(container.GetServiceName(), query)
	//mysqlDBDuration.Observe(endTime.Sub(beginTime).Seconds(), container.GetServiceName(), query)
}

// Waiting for version 2.0 .
func (mp *MonitorProxy) withMysqlTracer(query string, beginTime, endTime time.Time) {
	// span := opentracing.GetSpan(ctx, m.tracer,
	//	query, beginTime)
	// span.LogKV("sql", query)
	// span.FinishWithOptions(opentracing2.FinishOptions{FinishTime: endTime})
}

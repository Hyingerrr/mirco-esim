package mysql

import (
	"context"
	"fmt"
	"strings"
	"time"

	logx "github.com/Hyingerrr/mirco-esim/log"

	"github.com/Hyingerrr/mirco-esim/core/tracer"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/Hyingerrr/mirco-esim/container"

	"github.com/jinzhu/gorm"
)

func (c *Client) RegisterMetricsCallbacks(ctx context.Context, db *gorm.DB) {
	db.Callback().Create().After("gorm:after_create").Register("esim:metrics_after_create", func(scope *gorm.Scope) {
		if scope.HasError() {
			c.handleError(ctx, scope)
		}
	})

	db.Callback().Update().After("gorm:after_update").Register("esim:metrics_after_update", func(scope *gorm.Scope) {
		if scope.HasError() {
			c.handleError(ctx, scope)
		}
	})

	db.Callback().Query().Before("gorm:query").Register("esim:before_query", func(scope *gorm.Scope) {
		scope.InstanceSet("now_query_key", time.Now())
	})

	db.Callback().Query().After("gorm:after_query").Register("esim:metrics_after_query", func(scope *gorm.Scope) {
		if scope.HasError() {
			c.handleError(ctx, scope)
		}
	})

	db.Callback().Delete().After("gorm:after_delete").Register("esim:metrics_after_delete", func(scope *gorm.Scope) {
		if scope.HasError() {
			c.handleError(ctx, scope)
		}
	})

	db.Callback().RowQuery().After("gorm:row_query").Register("esim:metrics_after_sql", func(scope *gorm.Scope) {
		if scope.HasError() {
			c.handleError(ctx, scope)
		}
	})
}

func (c *Client) RegisterTraceCallbacks(db *gorm.DB) {
	c.traceOnce.Do(func() {
		// for create
		db.Callback().Create().Before("gorm:before_create").Register("esim:trace_before_create", func(scope *gorm.Scope) {
			iface, ok := scope.Get(OpenTracingContextKey)
			if !ok {
				return
			}
			ctx, ok := iface.(TraceContext)
			if !ok {
				return
			}
			spanner := ctx.StartSpan("CREATE",
				opentracing.ChildOf(ctx.spanCtx),
				tracer.TagDbSchema(scope.Dialect().CurrentDatabase()),
				tracer.TagStartTime(),
			)
			ext.SpanKindRPCClient.Set(spanner)
			ext.PeerService.Set(spanner, "mysql")
			ext.DBType.Set(spanner, "mysql")
			ext.DBInstance.Set(spanner, scope.TableName())

			scope.InstanceSet(OpenTracingSpanContextKey, spanner)
		})

		db.Callback().Create().After("gorm:after_create").Register("esim:trace_after_create", func(scope *gorm.Scope) {
			iface, ok := scope.InstanceGet(OpenTracingSpanContextKey)
			if !ok {
				return
			}

			spanner, ok := iface.(opentracing.Span)
			if !ok {
				return
			}

			if scope.HasError() {
				ext.DBStatement.Set(spanner, fmt.Sprintf("%v", scope.DB().QueryExpr()))
				ext.Error.Set(spanner, true)
				spanner.LogKV("event", "error", "message", scope.DB().Error.Error())
			} else {
				ext.DBStatement.Set(spanner, scope.SQL)
			}

			spanner.FinishWithOptions(tracer.TagFinishTime())
		})

		// for update
		db.Callback().Update().Before("gorm:before_update").Register("esim:trace_before_update", func(scope *gorm.Scope) {
			iface, ok := scope.Get(OpenTracingContextKey)
			if !ok {
				return
			}

			ctx, ok := iface.(TraceContext)
			if !ok {
				return
			}

			spanner := ctx.StartSpan("UPDATE",
				opentracing.ChildOf(ctx.spanCtx),
				tracer.TagDbSchema(scope.Dialect().CurrentDatabase()),
				tracer.TagStartTime(),
			)
			ext.SpanKindRPCClient.Set(spanner)
			ext.PeerService.Set(spanner, "mysql")
			ext.DBType.Set(spanner, "mysql")
			ext.DBInstance.Set(spanner, scope.TableName())

			scope.InstanceSet(OpenTracingSpanContextKey, spanner)
		})

		db.Callback().Update().After("gorm:after_update").Register("esim:trace_after_update", func(scope *gorm.Scope) {
			iface, ok := scope.InstanceGet(OpenTracingSpanContextKey)
			if !ok {
				return
			}

			spanner, ok := iface.(opentracing.Span)
			if !ok {
				return
			}

			if scope.HasError() {
				ext.DBStatement.Set(spanner, fmt.Sprintf("%v", scope.DB().QueryExpr()))
				ext.Error.Set(spanner, true)

				spanner.LogKV("event", "error", "message", scope.DB().Error.Error())
			} else {
				ext.DBStatement.Set(spanner, scope.SQL)
			}

			spanner.FinishWithOptions(tracer.TagFinishTime())
		})

		// for query
		db.Callback().Query().Before("gorm:query").Register("esim:trace_before_query", func(scope *gorm.Scope) {
			iface, ok := scope.Get(OpenTracingContextKey)
			if !ok {
				return
			}

			ctx, ok := iface.(TraceContext)
			if !ok {
				return
			}

			spanner := ctx.StartSpan("QUERY",
				opentracing.ChildOf(ctx.spanCtx),
				tracer.TagDbSchema(scope.Dialect().CurrentDatabase()),
				tracer.TagStartTime(),
			)
			ext.SpanKindRPCClient.Set(spanner)
			ext.PeerService.Set(spanner, "mysql")
			ext.DBType.Set(spanner, "mysql")
			ext.DBInstance.Set(spanner, scope.TableName())

			scope.InstanceSet(OpenTracingSpanContextKey, spanner)
		})

		db.Callback().Query().After("gorm:after_query").Register("esim:trace_after_query", func(scope *gorm.Scope) {
			iface, ok := scope.InstanceGet(OpenTracingSpanContextKey)
			if !ok {
				return
			}

			spanner, ok := iface.(opentracing.Span)
			if !ok {
				return
			}

			if scope.HasError() && scope.DB().Error != gorm.ErrRecordNotFound {
				ext.DBStatement.Set(spanner, fmt.Sprintf("%v", scope.DB().QueryExpr()))
				ext.Error.Set(spanner, true)

				spanner.LogKV("event", "error", "message", scope.DB().Error.Error())
			} else {
				ext.DBStatement.Set(spanner, scope.SQL)
			}

			spanner.FinishWithOptions(tracer.TagFinishTime())
		})

		// for raw
		db.Callback().RowQuery().Before("gorm:row_query").Register("esim:trace_before_sql", func(scope *gorm.Scope) {
			iface, ok := scope.Get(OpenTracingContextKey)
			if !ok {
				return
			}

			ctx, ok := iface.(TraceContext)
			if !ok {
				return
			}

			name := "SQL"
			switch strings.SplitN(scope.SQL, " ", 2)[0] {
			case "select", "SELECT", "Select":
				name = "QUERY"
			case "insert", "INSERT", "Insert", "replace", "REPLACE", "Replace":
				name = "CREATE"
			case "update", "UPDATE", "Update":
				name = "UPDATE"
			case "delete", "DELETE", "Delete":
				name = "DELETE"
			}

			name = "mysql_" + name

			spanner := ctx.StartSpan(name,
				opentracing.ChildOf(ctx.spanCtx),
				tracer.TagDbSchema(scope.Dialect().CurrentDatabase()),
				tracer.TagStartTime(),
			)
			ext.SpanKindRPCClient.Set(spanner)
			ext.PeerService.Set(spanner, "mysql")
			ext.DBType.Set(spanner, "mysql")
			ext.DBInstance.Set(spanner, scope.TableName())

			scope.InstanceSet(OpenTracingSpanContextKey, spanner)
		})

		db.Callback().RowQuery().After("gorm:row_query").Register("esim:trace_after_sql", func(scope *gorm.Scope) {
			iface, ok := scope.InstanceGet(OpenTracingSpanContextKey)
			if !ok {
				return
			}

			spanner, ok := iface.(opentracing.Span)
			if !ok {
				return
			}

			if scope.HasError() {
				ext.DBStatement.Set(spanner, scope.SQL)
				ext.Error.Set(spanner, true)
				spanner.LogKV("event", "error", "message", scope.DB().Error.Error())
			} else {
				ext.DBStatement.Set(spanner, scope.SQL)
			}

			spanner.FinishWithOptions(tracer.TagFinishTime())
		})

		// for delete
		db.Callback().Delete().Before("gorm:before_delete").Register("esim:trace_before_delete", func(scope *gorm.Scope) {
			iface, ok := scope.Get(OpenTracingContextKey)
			if !ok {
				return
			}

			ctx, ok := iface.(TraceContext)
			if !ok {
				return
			}

			spanner := ctx.StartSpan("DELETE",
				opentracing.ChildOf(ctx.spanCtx),
				tracer.TagDbSchema(scope.Dialect().CurrentDatabase()),
				tracer.TagStartTime(),
			)
			ext.SpanKindRPCClient.Set(spanner)
			ext.PeerService.Set(spanner, "mysql")
			ext.DBType.Set(spanner, "mysql")
			ext.DBInstance.Set(spanner, scope.TableName())

			scope.InstanceSet(OpenTracingSpanContextKey, spanner)
		})

		db.Callback().Delete().After("gorm:after_delete").Register("esim:trace_after_delete", func(scope *gorm.Scope) {
			iface, ok := scope.InstanceGet(OpenTracingSpanContextKey)
			if !ok {
				return
			}

			spanner, ok := iface.(opentracing.Span)
			if !ok {
				return
			}

			if scope.HasError() {
				ext.DBStatement.Set(spanner, fmt.Sprintf("%v", scope.DB().QueryExpr()))
				ext.Error.Set(spanner, true)
				spanner.LogKV("event", "error", "message", scope.DB().Error.Error())
			} else {
				ext.DBStatement.Set(spanner, scope.SQL)
			}

			spanner.FinishWithOptions(tracer.TagFinishTime())
		})
	})
}

func (c *Client) handleError(ctx context.Context, scope *gorm.Scope) {
	var (
		schema = scope.Dialect().CurrentDatabase()
	)

	if err := scope.DB().Error; err == gorm.ErrRecordNotFound {
		// db miss
		logx.Errorc(ctx, "mysql_miss, schema[%v], sql_desc[%v]", schema, fmt.Sprintf("%v", scope.DB().QueryExpr()))
		mysqlDBMiss.Inc(container.AppName(), schema, scope.QuotedTableName())
	} else {
		// db error
		logx.Errorc(ctx, "mysql_error: schema[%v], sql_desc[%v], err: %v", schema, fmt.Sprintf("%v", scope.DB().QueryExpr()), err)
		mysqlDBError.Inc(container.AppName(), schema, scope.QuotedTableName())
	}
}

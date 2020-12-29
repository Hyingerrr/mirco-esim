package mysql

import (
	"context"
	"fmt"
	"time"

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

func (c *Client) handleError(ctx context.Context, scope *gorm.Scope) {
	var (
		schema = scope.Dialect().CurrentDatabase()
	)

	if err := scope.DB().Error; err == gorm.ErrRecordNotFound {
		// db miss
		c.logger.Errorc(ctx, "mysql_miss, schema[%v], sql_desc[%v]", schema, fmt.Sprintf("%v", scope.DB().QueryExpr()))
		mysqlDBMiss.Inc(schema, scope.QuotedTableName())
	} else {
		// db error
		c.logger.Errorc(ctx, "mysql_error: schema[%v], sql_desc[%v], err: %v", schema, fmt.Sprintf("%v", scope.DB().QueryExpr()), err)
		mysqlDBError.Inc(schema, scope.QuotedTableName())
	}
}

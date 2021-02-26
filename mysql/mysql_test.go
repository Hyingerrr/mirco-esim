package mysql

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jukylin/esim/config"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

var (
	test1Config = DbConfig{
		Db:      "test_1",
		Dsn:     "root:root@tcp(localhost:3306)/test_1?charset=utf8&parseTime=True&loc=Local",
		MaxIdle: 10,
		MaxOpen: 100}

	test2Config = DbConfig{
		Db:      "test_2",
		Dsn:     "root:123456@tcp(localhost:3306)/test_1?charset=utf8&parseTime=True&loc=Local",
		MaxIdle: 10,
		MaxOpen: 100}
)

type TestStruct struct {
	ID     int    `json:"id"`
	Title1 string `json:"title1"`
}

type UserStruct struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

var db *sql.DB

func TestMain(m *testing.M) {
	//	logger := log.NewLogger()
	//
	//	pool, err := dockertest.NewPool("")
	//	if err != nil {
	//		logger.Fatalf("Could not connect to docker: %s", err)
	//	}
	//
	//	opt := &dockertest.RunOptions{
	//		Repository: "mysql",
	//		Tag:        "latest",
	//		Env:        []string{"MYSQL_ROOT_PASSWORD=123456"},
	//	}
	//
	//	// pulls an image, creates a container based on it and runs it
	//	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
	//		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
	//			"3306/tcp": {{HostIP: "", HostPort: "3306"}},
	//		}
	//	})
	//	if err != nil {
	//		logger.Fatalf("Could not start resource: %s", err.Error())
	//	}
	//
	//	err = resource.Expire(120)
	//	if err != nil {
	//		logger.Fatalf(err.Error())
	//	}
	//
	//	if err := pool.Retry(func() error {
	//		var err error
	//		db, err = sql.Open("mysql",
	//			"root:123456@tcp(localhost:3306)/mysql?charset=utf8&parseTime=True&loc=Local")
	//		if err != nil {
	//			return err
	//		}
	//		db.SetMaxOpenConns(100)
	//
	//		return db.Ping()
	//	}); err != nil {
	//		logger.Fatalf("Could not connect to docker: %s", err)
	//	}
	//
	//	sqls := []string{
	//		`create database test_1;`,
	//		`CREATE TABLE IF NOT EXISTS test_1.test(
	//		  id int not NULL auto_increment,
	//		  title VARCHAR(10) not NULL DEFAULT '',
	//		  PRIMARY KEY (id)
	//		)engine=innodb;`,
	//		`create database test_2;`,
	//		`CREATE TABLE IF NOT EXISTS test_2.user(
	//		  id int not NULL auto_increment,
	//		  username VARCHAR(10) not NULL DEFAULT '',
	//			PRIMARY KEY (id)
	//		)engine=innodb;`}
	//
	//	for _, execSQL := range sqls {
	//		res, err := db.Exec(execSQL)
	//		if err != nil {
	//			logger.Errorf(err.Error())
	//		}
	//		_, err = res.RowsAffected()
	//		if err != nil {
	//			logger.Errorf(err.Error())
	//		}
	//	}
	code := m.Run()
	//
	//	db.Close()
	//	// You can't defer this because os.Exit doesn't care for defer
	//	if err := pool.Purge(resource); err != nil {
	//		logger.Fatalf("Could not purge resource: %s", err)
	//	}
	os.Exit(code)
}

func TestInitAndSingleInstance(t *testing.T) {
	clientOptions := ClientOptions{}

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}),
		clientOptions.WithDB(db),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	_, ok := client.gdbs["test_1"]
	assert.True(t, ok)

	assert.Equal(t, client, NewClient())

	client.Close()
}

func TestProxyPatternWithTwoInstance(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	_ = config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
	)

	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Len(t, db1.GetErrors(), 0)

	db2 := client.GetCtxDb(ctx, "test_2")
	db2.Exec("use test_2;")
	assert.NotNil(t, db2)

	us := &UserStruct{}
	db2.Table("user").First(us)
	assert.Len(t, db1.GetErrors(), 0)

	client.Close()
}

func TestMulProxyPatternWithOneInstance(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	_ = config.NewMemConfig()
	// memConfig.Set("debug", true)
	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}))

	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Len(t, db1.GetErrors(), 0)

	client.Close()
}

func TestMulProxyPatternWithTwoInstance(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	_ = config.NewMemConfig()
	// memConfig.Set("debug", true)

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
	)

	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Len(t, db1.GetErrors(), 0)

	db2 := client.GetCtxDb(ctx, "test_2")
	db2.Exec("use test_2;")
	assert.NotNil(t, db2)

	us := &UserStruct{}
	db2.Table("user").First(us)

	assert.Len(t, db2.GetErrors(), 0)

	client.Close()
}

func BenchmarkParallelGetDB(b *testing.B) {
	clientOnce = sync.Once{}

	b.ReportAllocs()
	b.ResetTimer()

	clientOptions := ClientOptions{}
	_ = config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			client.GetCtxDb(ctx, "test_1")

			client.GetCtxDb(ctx, "test_2")
		}
	})

	client.Close()

	b.StopTimer()
}

func TestDummyProxy_Exec(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	_ = config.NewMemConfig()
	// memConfig.Set("debug", true)

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	db1.Table("test").Create(&TestStruct{})

	assert.Equal(t, db1.RowsAffected, int64(0))

	client.Close()
}

func TestClient_GetStats(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithStateTicker(10*time.Millisecond),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	time.Sleep(100 * time.Millisecond)

	mysqlDBStats.Set(100, "test_1", "max_open_conn")
	metric := &io_prometheus_client.Metric{}

	assert.Equal(t, float64(100), metric.Gauge.GetValue())

	mysqlDBStats.Set(1, "test_1", "idle")
	metric = &io_prometheus_client.Metric{}

	assert.Equal(t, float64(1), metric.Gauge.GetValue())

	client.Close()
}

//nolint:dupl
func TestClient_TxCommit(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	tx := db1.Begin()
	tx.Exec("insert into test values (1, 'test')")
	tx.Commit()
	if len(tx.GetErrors()) > 0 {
		assert.Error(t, tx.GetErrors()[0])
	}

	test := &TestStruct{}

	db1.Table("test").First(test)

	assert.Equal(t, 1, test.ID)

	client.Close()
}

//nolint:dupl
func TestClient_TxRollBack(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	db1.Exec("use test_1;")
	assert.NotNil(t, db1)

	tx := db1.Begin()
	tx.Exec("insert into test values (1, 'test')")
	tx.Rollback()
	if len(tx.GetErrors()) > 0 {
		assert.Error(t, tx.GetErrors()[0])
	}

	test := &TestStruct{}

	db1.Table("test").First(test)

	assert.Equal(t, 1, test.ID)

	client.Close()
}

func TestClient_RegisterMetricsCallbacks(t *testing.T) {
	clientOptions := ClientOptions{}
	_ = config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "test_1")
	//db1.Exec("use local_db;")
	assert.NotNil(t, db1)

	db1.Table("test_1").Create(&TestStruct{})

	assert.Equal(t, db1.RowsAffected, int64(0))
	metric := &io_prometheus_client.Metric{}

	c, err := mysqlDBError.GetMetric([]string{"test_1", "test_1"}...)
	err = c.Write(metric)
	assert.Nil(t, err)

	assert.Equal(t, float64(1), metric.Counter.GetValue())
	client.Close()
}

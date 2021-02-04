package new

var (
	configfc1 = &FileContent{
		FileName: "conf.yaml",
		Dir:      "conf",
		Content: `
runmode : 'dev'
appname : {{.ServerName}}

#HTTP 服务
httpport : 8080

#服务端
grpc_server_tcp : 50055
grpc_server_kp_time : 60 # s
grpc_server_kp_time_out : 5 # s
#链接超时 单位：ms
grpc_server_conn_time_out : 500

#客户端
grpc_client_kp_time : 60 # s
grpc_client_kp_time_out : 5 # s
#链接超时 单位：ms
grpc_client_conn_time_out : 500
grpc_client_permit_without_stream: true

jaeger_disabled: '${JAEGER_DISABLED}'
jaeger_local_agent_host_port: '0.0.0.0:6831'

#mysql
dbs:
#- {db: 'test', dsn: 'root:123456@tcp(:3306)/config?charset=utf8&parseTime=True&loc=Local',
#  maxidle: 10, maxopen: 100, maxlifetime : 10}
#- {db: 'test_slave', dsn: 'root:123456@tcp(:3306)/config?charset=utf8&parseTime=True&loc=Local',


#mongodb
#mgos:
#- {db: 'test', uri: 'mongodb://0.0.0.0:27017/admin?connect=direct'}

#ms
mgo_connect_timeout : 500
#min
mgo_max_conn_idle_time : 10
mgo_max_pool_size : 100
mgo_min_pool_size : 10

# http请求 单位：s
http_client_time_out : 3


#redis
redis_max_active : 50
redis_max_idle : 100
redis_idle_time_out : 600
redis_host : 0.0.0.0
redis_port : 6379
redis_password :

#redis 读超时 单位：ms
redis_read_time_out : 500
#redis 写超时 单位：ms
redis_write_time_out : 500
#redis 连接超时 单位：ms
redis_conn_time_out : 500


#prometheus http addr
prometheus_http_addr : 9002

# logger
log_output: stdout  # 日志位置，file 文件|both 文件和终端|stdout 终端
log_file: ./logs/{{.ServerName}}.log  # 文件地址，建议写绝对路径
log_level: INFO   # 日志等级  panic|fatal|error|warn|info|debug
log_format: json  # 日志格式  json|text
log_report_caller: false  # 是否显示文件:行号
log_stack_trace: false  # 是否打印堆栈
log_max_size: 100   # 单个文件最大size
log_max_age: 15   # 保留旧文件的最大天数
log_backup_count: 10  # 保留旧文件的最大个数
log_compress: true  # 是否压缩/归档旧文件

# 微信报警 配置样例
wx_web_hook: e0c3df32-547b-4699-a887-67c5ae8ea877  # wx群ID
wx_retries: 3
wx_interval: 3
`,
	}

	configfc2 = &FileContent{
		FileName: "monitoring.yaml",
		Dir:      "conf",
		Content: `
#grpc 服务端
#开启慢检查 true/false
grpc_server_check_slow : {{.Monitoring}}
#单位: ms
grpc_server_slow_time : 500
#开启 tracer true/false
grpc_server_tracer : {{.Monitoring}}
#启动metrice true/false
grpc_server_metrics : {{.Monitoring}}
#启动debug true/false
grpc_server_debug: {{.Monitoring}}
# 单位ms handle
grpc_server_timeout: 5000
# 启动字段验证
grpc_server_validate: true

#grpc 客户端
#开启慢检查 true/false
grpc_client_check_slow : {{.Monitoring}}
#单位: ms
grpc_client_slow_time : 500
#开启 tracer true/false
grpc_client_tracer : {{.Monitoring}}
#启动metrice true/false
grpc_client_metrics : {{.Monitoring}}
#启动debug true/false
grpc_client_debug: {{.Monitoring}}
# 单位ms handle
grpc_client_timeout: 5000

#mysql
#开启慢检查 true/false
mysql_check_slow : {{.Monitoring}}
# 大于 slow_sql_time 为慢sql 单位：ms
mysql_slow_time : 500
#开启 tracer true/false
mysql_tracer : {{.Monitoring}}
#启动metrice true/false
mysql_metrics : {{.Monitoring}}

#mongodb
#开启慢检查 true/false
mgo_check_slow : {{.Monitoring}}
# 大于 slow_sql_time 为慢命令 单位：ms
mgo_slow_time : 250
#开启 tracer true/false
mgo_tracer : {{.Monitoring}}
#启动metrice true/false
mgo_metrics : {{.Monitoring}}


# http client
#开启慢检查 true/false
http_client_check_slow : {{.Monitoring}}
#  单位：ms
http_client_slow_time : 500
#开启 tracer true/false
http_client_tracer : {{.Monitoring}}
#启动metrice true/false
http_client_metrics : {{.Monitoring}}

# http server
#开启 tracer true/false
http_tracer: {{.Monitoring}}
#启动metrice true/false
http_metrics: {{.Monitoring}}

#redis
#开启慢检查 true/false
redis_check_slow : {{.Monitoring}}
#慢命令 单位 ms
redis_slow_time : 50
#开启 tracer true/false
redis_tracer : {{.Monitoring}}
#启动metrice true/false
redis_metrics : {{.Monitoring}}

# tracer collect server
tracer_jaeger_upd: ':6831'
`,
	}
)

func initConfigFiles() {
	Files = append(Files, configfc1, configfc2)
}

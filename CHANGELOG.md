##  v2.0.1 2021-02-24 
1. 新增gRPC客户端请求超时自定义设置
    - 超时设置原则：对于客户端，自定义超时 优先于 配置文件中设置默认的grpc_client_timeout超时；对于服务端，从ctx中取出timeout与配置文件中的比较，取较小值；
    - 自定义超时设置示例：
 ```go
      test.NewHelloServerClient(conn).SayGoodbye(ctx, &test.HelloRequest{Name: "HH"}, WithTimeout(10*time.Second))
 ```

2. 新增gRPC服务端字段验证
    - (1) 安装gogoproto
    ```bash
    go get github.com/gogo/protobuf/gogoproto/gogo.proto
    ```

      [gogo proto写法参照](https://kj_test.bhecard.com:8443/gitlab/go-dev/esim/-/blob/hy/mertrics/grpc/test/hello.proto)
        - 配置规则参考validator.v10 (github.com/go-playground/validator/v10)
    - (2) 配置文件开启： grpc_server_validate = true
    - (3) 生成proto对应的pb文件
 ```bash
    esim proto -v=true ../path/test.proto
 ```

3. 重写了gRPC客户端和服务端，添加了metric、tracer、timeout、字段验证、recovery等的处理，可以自定义gRPC拦截器、根据业务场景自定义metric收集；
4. 新增监控和链路追踪相关
     - [具体参见](https://kj_test.bhecard.com:8443/confluence/pages/viewpage.action?pageId=1310725)

5. 新增开箱即用的config、logger对象，减少基础对象Infra层层嵌套；

6. 新增了服务环境变量，如：

   ```go
   // 是否是生产环境
   if xenv.IsPro() || xenv.IsDev() {
     // todo
   }
   ```

   


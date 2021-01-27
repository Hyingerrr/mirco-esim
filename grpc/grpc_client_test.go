package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jukylin/esim/grpc/test"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	address      = "0.0.0.0"
	isTest       = "is test"
	callPanic    = "call_panic"
	callPanicArr = "callPanciArr"
	esim         = "esim"
)

var (
	logger    log.Logger
	tcpAddr   = &net.TCPAddr{IP: net.ParseIP(address).To4(), Port: 50250}
	memConfig config.Config

	clientOpt *ClientOptions
)

type server struct{}

func (s *server) SayGoodbye(ctx context.Context, in *test.HelloRequest) (*test.HelloResponse, error) {
	resp := &test.HelloResponse{
		Head:   &test.InternalResponse{RespCode: "0000", RespMsg: "SUCCESS"},
		NameEn: in.Name + "_en",
		AgeEn:  in.Age,
	}
	return resp, nil
}

func TestMain(m *testing.M) {
	logger = log.NewLogger()
	options := config.ViperConfOptions{}
	memConfig = config.NewViperConfig(options.WithConfigType("yaml"),
		options.WithConfFile([]string{"../config/a.yaml", "../config/b.yaml"}))
	opts := ClientOptionals{}
	clientOpt = NewClientOptions(
		opts.WithDialOptions())

	svr := NewServer()

	test.RegisterHelloServerServer(svr.server, &server{})

	svr.Use(panicResp())
	svr.Start()

	code := m.Run()

	os.Exit(code)
}

func TestGrpcClient(t *testing.T) {
	ctx := context.Background()
	conn := NewClient(clientOpt).DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := test.NewHelloServerClient(conn)

	r, err := c.SayGoodbye(ctx, &test.HelloRequest{Head: &test.InternalHeader{
		AppId:   "QY0002",
		TermNo:  "",
		MerchNo: "Qy08032324ccs",
		TraceId: "x1x2x3x4x5",
	}, Name: esim, Age: -1})
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				logger.Errorc(ctx, "grpc dial timeout deadline: %v", statusErr.Message())
			}
		}
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.String())
	}
}

func TestSlowClient_TimeOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	conn := NewClient(clientOpt).DialContext(ctx, tcpAddr.String())
	defer conn.Close()
	c := test.NewHelloServerClient(conn)

	fmt.Println("Start Time: " + time.Now().String())
	r, err := c.SayGoodbye(ctx, &test.HelloRequest{Name: "HuangYin"}, WithTimeout(10*time.Second))
	if err != nil {
		logger.Errorf(err.Error())
	} else {
		assert.NotEmpty(t, r.NameEn)
	}
}

func panicResp() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		if req.(*test.HelloRequest).Name == callPanic {
			panic(isTest)
		} else if req.(*test.HelloRequest).Name == callPanicArr {
			var arr [1]string
			arr[0] = isTest
			panic(arr)
		}
		resp, err = handler(ctx, req)

		return resp, err
	}
}

func TestMetaData(t *testing.T) {
	ctx := context.Background()
	fmt.Println("Start Time: " + time.Now().String())
	conn := NewClient(clientOpt).DialContext(ctx, tcpAddr.String())

	defer conn.Close()
	c := test.NewHelloServerClient(conn)

	r, err := c.SayGoodbye(ctx, &test.HelloRequest{Head: &test.InternalHeader{
		AppId:    "AppID1",
		TermNo:   "TermNo1",
		MerchNo:  "商户A",
		DstSysId: "",
		SrcSysId: "SrcB",
		ProdCd:   "SM101",
		TranCd:   "MP010",
		TranSeq:  "dsfsgdsggfhj",
		TraceId:  "C1c2c3vv44",
	}, Name: "call_panic1", Age: 30, Address: "上海市"})
	if err != nil {
		logger.Errorf(err.Error())
	}
	assert.Error(t, err)
	assert.Nil(t, r)
}

package grpc

import (
	"context"
	"fmt"
	"runtime"

	"github.com/jukylin/esim/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (gs *Server) recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, info.FullMethod)
			}
		}()

		resp, err = handler(ctx, req)
		return
	}
}

func recoverFrom(r interface{}, fullMethod string) error {
	var stacktrace string
	for i := 1; ; i++ {
		_, f, l, got := runtime.Caller(i)
		if !got {
			break
		}

		stacktrace += fmt.Sprintf("%s:%d\n", f, l)
	}

	log.Errorf("handle gprc panic. fullMethod[%v], panic: %v, stack: %v",
		fullMethod, r, stacktrace)

	return status.Errorf(codes.Unknown, "%s", r)
}

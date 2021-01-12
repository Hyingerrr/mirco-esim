package grpc

import (
	"context"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	opentracing2 "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// tracerId If not found opentracing's tracer_id then generate a new tracer_id.
// Recommend to the end of the Interceptor.
func (gs *Server) tracerID() grpc.UnaryServerInterceptor {
	tracerID := tracerid.TracerID()
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		sp := opentracing2.SpanFromContext(ctx)
		if sp == nil {
			ctx = context.WithValue(ctx, tracerid.ActiveEsimKey, tracerID())
		}

		resp, err = handler(ctx, req)

		return resp, err
	}
}

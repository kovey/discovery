package krpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	Ko_Trace_Id = "ko_trace_id"
)

func trace(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	traceId, ok := ctx.Value(Ko_Trace_Id).(string)
	if !ok {
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	traceInfo := metadata.Pairs(Ko_Trace_Id, traceId)
	return invoker(metadata.NewOutgoingContext(ctx, traceInfo), method, req, reply, cc, opts...)
}

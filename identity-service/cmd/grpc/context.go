package grpc

import (
	"context"

	"google.golang.org/grpc"
)

// StreamContextWrapper injects context into a server stream.
type StreamContextWrapper interface {
	grpc.ServerStream
	SetContext(context.Context)
}

type wrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrapper) Context() context.Context {
	return w.ctx
}

func (w *wrapper) SetContext(ctx context.Context) {
	w.ctx = ctx
}

func newStreamContextWrapper(stream grpc.ServerStream) StreamContextWrapper {
	ctx := stream.Context()
	return &wrapper{
		ServerStream: stream,
		ctx:          ctx,
	}
}

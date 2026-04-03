package ai

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func deadlineExceeded(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return status.Error(codes.Canceled, "The client canceled the request!")
	}
	return nil
}
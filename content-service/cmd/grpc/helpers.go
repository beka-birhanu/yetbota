package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	"github.com/google/uuid"
)

func makeUnaryServerInterceptor(sessionManager domainAuth.SessionManager) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		appCtx := ctxYB.New()
		appCtx.SetMethod(info.FullMethod)
		appCtx.SetXCorrelationID(uuid.NewString())

		if _, skip := constants.SkipAuthGrpc[info.FullMethod]; !skip {
			token, err := extractBearerToken(ctx)
			if err != nil {
				return nil, err
			}

			userSess, err := sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
				TokenType: domainAuth.AccessToken,
				Token:     token,
			})
			if err != nil {
				return nil, status.Error(codes.Unauthenticated, "invalid token")
			}

			appCtx.UserSession = *userSess
			appCtx.SetGrpcAuthToken(token)
		}

		ctx = context.WithValue(ctx, ctxYB.AppSession, appCtx)
		return handler(ctx, req)
	}
}

func makeStreamServerInterceptor(sessionManager domainAuth.SessionManager) grpc.StreamServerInterceptor {
	return func(
		srv any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()
		appCtx := ctxYB.New()
		appCtx.SetMethod(info.FullMethod)
		appCtx.SetXCorrelationID(uuid.NewString())

		if _, skip := constants.SkipAuthGrpc[info.FullMethod]; !skip {
			token, err := extractBearerToken(ctx)
			if err != nil {
				return err
			}

			userSess, err := sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
				TokenType: domainAuth.AccessToken,
				Token:     token,
			})
			if err != nil {
				return status.Error(codes.Unauthenticated, "invalid token")
			}

			appCtx.UserSession = *userSess
			appCtx.SetGrpcAuthToken(token)
		}

		newCtx := context.WithValue(ctx, ctxYB.AppSession, appCtx)
		wrapped := newStreamContextWrapper(stream)
		wrapped.SetContext(newCtx)
		return handler(srv, wrapped)
	}
}

func extractBearerToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	return values[0], nil
}

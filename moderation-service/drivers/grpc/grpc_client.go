package grpc

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	ctxRP "github.com/beka-birhanu/yetbota/moderation-service/internal/domain/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	CtxSession     = "Context_session"
	XCorrelationID = "X-Correlation-ID"
	Header         = "header"
	Authorization  = "Authorization"
	defaultTimeout = 10 // second
)

type RpcConnection struct {
	Conn    *grpc.ClientConn
	options Options
}

func NewGrpcConnection(options Options) *RpcConnection {
	var conn *grpc.ClientConn
	var err error

	if len(options.Cert) == 0 {
		conn, err = grpc.Dial(options.Address, grpc.WithInsecure(), withClientUnaryInterceptor(), withClientStreamInterceptor())
	} else {
		certPEM := strings.ReplaceAll(options.Cert, "\\n", "\n")
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM([]byte(certPEM)) {
			log.Fatalf("failed to append certificate in GRPC client")
		}
		// Create the TransportCredentials using the CertPool
		creds := credentials.NewClientTLSFromCert(certPool, options.ServerName)
		conn, err = grpc.Dial(options.Address, grpc.WithTransportCredentials(creds), withClientUnaryInterceptor())
	}
	if err != nil {
		panic(err)
	}

	return &RpcConnection{
		Conn:    conn,
		options: options,
	}
}

func (rpc *RpcConnection) CreateContext(parentCtx context.Context, ctxSess *ctxRP.Context) context.Context {
	ctx := context.WithValue(parentCtx, ctxRP.AppSession, ctxSess)
	auth := ""
	if head, ok := ctxSess.Header.(http.Header); ok { // get auth from rest call
		auth = head.Get(Authorization)
	}
	if ctxSess.Method == "GRPC" { // get auth from grpc call
		auth = fmt.Sprintf("Bearer %s", ctxSess.GrpcAuthToken)
	}
	ctxWithMetadata := metadata.NewOutgoingContext(ctx,
		metadata.Pairs(
			XCorrelationID, ctxSess.XCorrelationID,
			Authorization, auth,
		),
	)
	return ctxWithMetadata
}

func (rpc *RpcConnection) Timeout() time.Duration {
	if rpc.options.Timeout > 0 {
		return rpc.options.Timeout * time.Second
	}
	return time.Duration(defaultTimeout) * time.Second
}

func clientInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	var ctxSess *ctxRP.Context
	if _, ok := ctx.Value(CtxSession).(*ctxRP.Context); ok {
		ctxSess = ctx.Value(CtxSession).(*ctxRP.Context)
	}
	if _, ok := ctx.Value(ctxRP.AppSession).(*ctxRP.Context); ok {
		ctxSess = ctx.Value(ctxRP.AppSession).(*ctxRP.Context)
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, metadata.Pairs(XCorrelationID, ctxSess.XCorrelationID))
	err = invoker(ctxWithMetadata, method, req, reply, cc, opts...)

	return
}

func clientStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctxSess := ctx.Value(CtxSession).(*ctxRP.Context)
	auth := ""
	if head, ok := ctxSess.Header.(http.Header); ok { // get auth from rest call
		auth = head.Get(Authorization)
	}
	if ctxSess.Method == "GRPC" { // get auth from grpc call
		auth = fmt.Sprintf("Bearer %s", ctxSess.GrpcAuthToken)
	}
	ctxWithMetadata := metadata.NewOutgoingContext(ctx,
		metadata.Pairs(
			XCorrelationID, ctxSess.XCorrelationID,
			Authorization, auth,
		),
	)
	return streamer(ctxWithMetadata, desc, cc, method, opts...)
}

func withClientUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(clientInterceptor)
}

func withClientStreamInterceptor() grpc.DialOption {
	return grpc.WithStreamInterceptor(clientStreamInterceptor)
}
